package temporalcloudcli

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
)

const (
	// OAuth error defined in RFC-6749.
	// https://datatracker.ietf.org/doc/html/rfc6749#section-5.2
	invalidGrantErr = "invalid_grant"
)

const (
	loginDefaultClientID = "d7V5bZMLCbRLfRVpqC567AqjAERaWHhl"
)

func login(cctx *CommandContext, tokenConfig *TokenConfig) (*TokenConfig, error) {
	if tokenConfig == nil {
		defaultConfig, err := defaultTokenConfig(cctx)
		if err != nil {
			return nil, err
		}
		tokenConfig = defaultConfig
	}

	resp, err := tokenConfig.OAuthConfig.DeviceAuth(cctx.Context, oauth2.SetAuthURLParam("audience", tokenConfig.Audience))
	if err != nil {
		return nil, fmt.Errorf("failed to perform device auth: %w", err)
	}

	domainURL, err := parseURL(tokenConfig.Domain)
	if err != nil {
		return nil, fmt.Errorf("failed to parse domain: %w", err)
	}

	verificationURL, err := parseURL(resp.VerificationURIComplete)
	if err != nil {
		return nil, fmt.Errorf("failed to parse verification URL: %w", err)
	} else if verificationURL.Hostname() != domainURL.Hostname() {
		// We expect the verification URL to be the same host as the domain URL.
		// Otherwise the response could have us POST to any arbitrary URL.
		return nil, fmt.Errorf("domain URL `%s` does not match verification URL `%s` in response", domainURL.Hostname(), verificationURL.Hostname())
	}

	err = openBrowser(cctx, "Login via this url", verificationURL.String())
	if err != nil {
		// Notify the user but ensure they can continue the process.
		fmt.Printf("Failed to open the browser, click the link to continue: %v", err)
	}

	token, err := tokenConfig.OAuthConfig.DeviceAccessToken(cctx.Context, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve access token: %w", err)
	}
	// Print to stderr so other tooling can parse the command output.
	fmt.Fprintln(os.Stderr, "Successfully logged in!")

	tokenConfig.OAuthToken = token
	tokenConfig.cctx = cctx

	err = tokenConfig.Store()
	if err != nil {
		return nil, fmt.Errorf("failed to store token config: %w", err)
	}

	return tokenConfig, nil
}

func defaultTokenConfig(cctx *CommandContext) (*TokenConfig, error) {
	domainURL, err := parseURL(cctx.RootCommand.Domain)
	if err != nil {
		return nil, fmt.Errorf("failed to parse domain URL: %w", err)
	}
	clientID := loginDefaultClientID
	if cctx.RootCommand.ClientId != "" {
		clientID = cctx.RootCommand.ClientId
	}

	return &TokenConfig{
		Audience: cctx.RootCommand.Audience,
		Domain:   domainURL.String(),
		OAuthConfig: oauth2.Config{
			ClientID: clientID,
			Endpoint: oauth2.Endpoint{
				DeviceAuthURL: domainURL.JoinPath("oauth", "device", "code").String(),
				TokenURL:      domainURL.JoinPath("oauth", "token").String(),
				AuthStyle:     oauth2.AuthStyleInParams,
			},
			Scopes: []string{"openid", "profile", "user", "offline_access"},
		},
		cctx: cctx,
	}, nil
}
