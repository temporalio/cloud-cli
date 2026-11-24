package temporalcloudcli

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/temporalio/cloud-cli/internal/onboarding"
	onboardingservice "github.com/temporalio/cloud-cli/internal/onboardingservice/v1"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cliprompter "github.com/temporalio/cloud-cli/temporalcloudcli/prompter"
)

// initCommand hides the command from help while it is in development.
func (c *CloudSignupCommand) initCommand(*CommandContext) {
	c.Command.Hidden = true
}

func (c *CloudSignupCommand) run(cctx *CommandContext, _ []string) error {
	oauthConfig, err := newOAuthClientConfig(c.Domain, c.ClientId, c.RedirectUrl)
	if err != nil {
		return fmt.Errorf("failed to generate OAuth client config: %w", err)
	}

	connection, err := c.chooseLoginMethod(cctx)
	if err != nil {
		return err
	}
	authParams := append([]oauth2.AuthCodeOption{oauth2.SetAuthURLParam("audience", c.Audience)}, c.connectionAuthParams(connection)...)
	token, err := c.login(cctx, oauthConfig, authParams)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	conn, err := c.dialOnboarding(oauthConfig.TokenSource(cctx, token))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := onboardingservice.NewOnboardingServiceClient(conn)
	req := &onboardingservice.GetOrCreateCLIOnboardingStateRequest{}
	var resp *onboardingservice.GetOrCreateCLIOnboardingStateResponse
	if c.NoWait {
		resp, err = client.GetOrCreateCLIOnboardingState(cctx, req)
		if err != nil {
			return fmt.Errorf("failed to get onboarding state: %w", err)
		}
	} else {
		resp, err = onboarding.Wait(cctx, client, req, onboarding.DefaultPollConfig, func(step onboardingservice.StepType) {
			if !cctx.JSONOutput {
				cctx.Printer.Printlnf("Waiting for your account (step: %s)...", step)
			}
		})
		if err != nil {
			return err
		}
	}

	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
	}
	switch {
	case resp.GetUserAlreadyExists():
		cctx.Printer.Println("You already have a Temporal Cloud account. Run `temporal cloud login` to sign in.")
	case onboarding.IsComplete(resp):
		cctx.Printer.Println("Your Temporal Cloud account is ready. Run `temporal cloud login` to sign in.")
	default:
		cctx.Printer.Printlnf("Onboarding step: %s", resp.GetState().GetCurrentStep())
	}
	return nil
}

// login runs the browser login. A new email + password user is denied until
// they verify their email, so in a terminal it waits for them to click the
// emailed link and then logs in again, instead of making them rerun signup.
func (c *CloudSignupCommand) login(cctx *CommandContext, config *oauth2.Config, authParams []oauth2.AuthCodeOption) (*oauth2.Token, error) {
	for {
		token, err := Login(cctx, config, !c.Parent.DisablePopUp, authParams...)
		var authErr *AuthorizationError
		if !errors.As(err, &authErr) || !authErr.IsEmailNotVerified() {
			return token, err
		}
		if !cliprompter.IsTerminal(cctx.Options.Stdin, cctx.Options.Stderr) {
			return nil, errors.New("verify your email using the link we sent you, then run signup again")
		}
		fmt.Fprint(cctx.Options.Stderr,
			"Verify your email using the link we sent you, then press Enter to continue.")
		if err := waitForEnter(cctx, cctx.Options.Stdin); err != nil {
			return nil, err
		}
		// The account now exists, so show the login form rather than sign-up.
		authParams = append(authParams, oauth2.SetAuthURLParam("screen_hint", "login"))
	}
}

// waitForEnter blocks until a line is read from in or ctx is canceled. Ctrl-C
// cancels ctx rather than killing the process, so a plain read would hang.
func waitForEnter(ctx context.Context, in io.Reader) error {
	done := make(chan error, 1)
	go func() {
		_, err := bufio.NewReader(in).ReadString('\n')
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to read from terminal: %w", err)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

var loginMethods = []struct {
	value string
	label string
}{
	{value: "google", label: "Google"},
	{value: "microsoft", label: "Microsoft"},
	{value: "password", label: "Email + Password"},
}

// chooseLoginMethod returns --login-method if set. Otherwise, in an interactive
// terminal, it asks with an arrow-key menu. In any other case (scripts, agents,
// JSON output) it returns "" and Auth0 shows its own provider picker.
func (c *CloudSignupCommand) chooseLoginMethod(cctx *CommandContext) (string, error) {
	if c.LoginMethod.Value != "" {
		return c.LoginMethod.Value, nil
	}
	if cctx.JSONOutput || !cliprompter.IsTerminal(cctx.Options.Stdin, cctx.Options.Stdout) {
		// Stderr, so JSON on stdout stays parseable and agents still see the hint.
		fmt.Fprintln(cctx.Options.Stderr,
			"No --login-method set: choose one in the browser, or pass --login-method google|microsoft|password.")
		return "", nil
	}
	labels := make([]string, len(loginMethods))
	for i, conn := range loginMethods {
		labels[i] = conn.label
	}
	i, err := cliprompter.SelectInTerminal(cctx.Options.Stdin, cctx.Options.Stdout, "Sign up with:", labels)
	if err != nil {
		return "", err
	}
	return loginMethods[i].value, nil
}

// connectionAuthParams maps a connection value to Auth0 authorize parameters,
// the same way the /get-cloud sign-up buttons do.
func (c *CloudSignupCommand) connectionAuthParams(connection string) []oauth2.AuthCodeOption {
	switch connection {
	case "google":
		return []oauth2.AuthCodeOption{oauth2.SetAuthURLParam("connection", "google-oauth2")}
	case "microsoft":
		// The Microsoft connection is named differently per tenant
		// (microsoft-v2 in prod, microsoft elsewhere), mirroring the website.
		name := "microsoft"
		if c.Domain == "login.tmprl.cloud" {
			name = "microsoft-v2"
		}
		return []oauth2.AuthCodeOption{oauth2.SetAuthURLParam("connection", name)}
	case "password":
		return []oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("connection", "Username-Password-Authentication"),
			oauth2.SetAuthURLParam("screen_hint", "signup"),
		}
	default:
		return nil
	}
}

func (c *CloudSignupCommand) dialOnboarding(tokenSource oauth2.TokenSource) (*grpc.ClientConn, error) {
	transport := credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
	if c.OnboardingInsecure {
		transport = insecure.NewCredentials()
	}
	conn, err := grpc.NewClient(
		c.OnboardingAddress,
		grpc.WithTransportCredentials(transport),
		grpc.WithPerRPCCredentials(bearerTokenCredentials{
			tokenSource:   tokenSource,
			requireSecure: !c.OnboardingInsecure,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to onboarding service: %w", err)
	}
	return conn, nil
}

// bearerTokenCredentials attaches an OAuth access token to each RPC. Unlike
// grpc/credentials/oauth.TokenSource, it can be used without TLS so the
// command can reach a local onboarding service.
type bearerTokenCredentials struct {
	tokenSource   oauth2.TokenSource
	requireSecure bool
}

func (b bearerTokenCredentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	token, err := b.tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	return map[string]string{"authorization": "Bearer " + token.AccessToken}, nil
}

func (b bearerTokenCredentials) RequireTransportSecurity() bool {
	return b.requireSecure
}
