package temporalcloudcli

import (
	"context"
	"fmt"

	"github.com/temporalio/cli/cliext"
	"go.temporal.io/cloud-sdk/cloudclient"
)

func (c *CloudCommand) GetAPIKey(ctx context.Context) (string, error) {
	loadProfileResult, err := cliext.LoadProfile(cliext.LoadProfileOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to load login configuration: %w, please run `temporal cloud login`", err)
	}

	// check if we have had a valid token in the past
	if loadProfileResult.Profile == nil || loadProfileResult.Profile.OAuth == nil {
		return "", fmt.Errorf("no login configurations found, please run `temporal cloud login`")
	}

	token, err := cliext.NewOAuthClient(loadProfileResult.Profile.OAuth.OAuthClientConfig).Token(ctx, loadProfileResult.Profile.OAuth)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve access token: %w, please run `temporal cloud login`", err)
	}
	if token.AccessTokenRefreshed {
		token.AccessTokenRefreshed = false // reset the flag before saving
		loadProfileResult.Profile.OAuth.OAuthToken = token
		loadProfileResult.Config.Profiles[loadProfileResult.ProfileName] = loadProfileResult.Profile
		if err := cliext.WriteConfig(cliext.WriteConfigOptions{
			Config: loadProfileResult.Config,
		}); err != nil {
			return "", fmt.Errorf("failed to write config file: %w", err)
		}
	}
	return token.AccessToken, nil
}

func newCloudClient(cctx *CommandContext) (*cloudclient.Client, error) {
	opts := cloudclient.Options{}
	if cctx.RootCommand.Server != "" {
		opts.HostPort = cctx.RootCommand.Server
	}
	if cctx.RootCommand.ApiKey != "" {
		// an explicit api key was provided, use it
		opts.APIKey = cctx.RootCommand.ApiKey
	} else {
		// fallaback to the oauth based sso token provider
		opts.APIKeyReader = cctx.RootCommand
	}

	cloudClient, err := cloudclient.New(opts)
	if err != nil {
		return nil, err
	}
	return cloudClient, nil
}
