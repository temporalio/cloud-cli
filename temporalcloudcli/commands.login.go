package temporalcloudcli

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/temporalio/cli/cliext"
)

func (c *CloudLoginCommand) run(cctx *CommandContext, _ []string) error {
	var oauth cliext.OAuthConfig
	// First load the config to see if we have an existing config
	loadProfileResult, err := cliext.LoadProfile(cliext.LoadProfileOptions{})
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}
	if loadProfileResult.Profile.OAuth != nil &&
		!reflect.DeepEqual(*loadProfileResult.Profile.OAuth, cliext.OAuthConfig{}) &&
		!c.Reset {
		// Existing OAuth config found, use it as a base
		oauth = *loadProfileResult.Profile.OAuth
	} else {
		var err error
		oauth.OAuthClientConfig, err = c.generateClientConfig()
		if err != nil {
			return fmt.Errorf("failed to generate OAuth client config: %w", err)
		}
	}

	oauthClient, err := cliext.NewOAuthClient(oauth.OAuthClientConfig)
	if err != nil {
		return fmt.Errorf("failed to create OAuth client: %w", err)
	}

	oauthToken, err := oauthClient.Login(cctx)
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}

	oauth.OAuthToken = oauthToken
	loadProfileResult.Profile.OAuth = &oauth
	loadProfileResult.Config.Profiles[loadProfileResult.ProfileName] = loadProfileResult.Profile
	if err := cliext.WriteConfig(loadProfileResult.Config, ""); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
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

func (c *CloudLoginCommand) generateClientConfig() (cliext.OAuthClientConfig, error) {
	domainURL, err := parseURL(c.Domain)
	if err != nil {
		return cliext.OAuthClientConfig{}, fmt.Errorf("failed to parse domain: %w", err)
	}

	return cliext.OAuthClientConfig{
		ClientID: c.ClientId,
		// AuthURL:  domainURL.JoinPath("authorize").String(),
		AuthURL:  domainURL.JoinPath("oauth", "device", "code").String(),
		TokenURL: domainURL.JoinPath("oauth", "token").String(),
		RequestParams: map[string]string{
			"audience": c.Audience,
		},
		Scopes: []string{"openid", "profile", "user", "offline_access"},
	}, nil
}
