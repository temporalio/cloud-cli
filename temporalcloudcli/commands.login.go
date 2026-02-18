package temporalcloudcli

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/pkg/browser"
	"github.com/temporalio/cli/cliext"
	"golang.org/x/oauth2"
)

func (c *CloudLoginCommand) run(cctx *CommandContext, _ []string) error {
	var oauthConfig cliext.OAuthConfig
	// First load the config to see if we have an existing config
	loadClientOauthRes, err := cliext.LoadClientOAuth(cliext.LoadClientOAuthOptions{})
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}
	if loadClientOauthRes.OAuth != nil &&
		loadClientOauthRes.OAuth.Token != nil &&
		loadClientOauthRes.OAuth.ClientConfig != nil &&
		!reflect.DeepEqual(*loadClientOauthRes.OAuth, cliext.OAuthConfig{}) &&
		!reflect.DeepEqual(*loadClientOauthRes.OAuth.ClientConfig, oauth2.Config{}) &&
		!c.Reset {
		// Existing OAuth config found, use it as a base
		oauthConfig = *loadClientOauthRes.OAuth
	} else {
		var err error
		oauthConfig.ClientConfig, err = c.generateOauthClientConfig()
		if err != nil {
			return fmt.Errorf("failed to generate OAuth client config: %w", err)
		}
	}

	oauthToken, err := Login(cctx, oauthConfig.ClientConfig, oauth2.SetAuthURLParam("audience", c.Audience))
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	oauthConfig.Token = oauthToken
	if err := cliext.StoreClientOAuth(cliext.StoreClientOAuthOptions{
		OAuth: &oauthConfig,
	}); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	fmt.Println("Login successful!")
	return nil
}

func parseURL(s string) (*url.URL, error) {
	// Without a scheme, url.Parse would interpret the path as a relative file path.
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = fmt.Sprintf("%s%s", "https://", s)
	}

	u, err := url.ParseRequestURI(s)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		u.Scheme = "https"
	}

	return u, err
}

func (c *CloudLoginCommand) generateOauthClientConfig() (*oauth2.Config, error) {
	domainURL, err := parseURL(c.Domain)
	if err != nil {
		return nil, fmt.Errorf("failed to parse domain: %w", err)
	}

	return &oauth2.Config{
		ClientID: c.ClientId,
		Endpoint: oauth2.Endpoint{
			AuthURL:   domainURL.JoinPath("authorize").String(),
			TokenURL:  domainURL.JoinPath("oauth", "token").String(),
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: c.RedirectUrl,
		Scopes:      []string{"openid", "profile", "user", "offline_access"},
	}, nil
}

func (c *CloudLogoutCommand) run(cctx *CommandContext, _ []string) error {
	domainURL, err := parseURL(c.Domain)
	if err != nil {
		return fmt.Errorf("failed to parse domain: %w", err)
	}

	if err := cliext.StoreClientOAuth(cliext.StoreClientOAuthOptions{
		OAuth: &cliext.OAuthConfig{},
	}); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	logoutURL := domainURL.JoinPath("v2", "logout")
	fmt.Printf("Opening browser to logout. If it doesn't open, visit: %s\n", logoutURL.String())
	_ = browser.OpenURL(logoutURL.String())

	return nil
}
