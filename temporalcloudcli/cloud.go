package temporalcloudcli

import (
	"context"
	"errors"
	"fmt"

	"github.com/temporalio/cli/cliext"
	"go.temporal.io/cloud-sdk/cloudclient"
)

func (c *CloudCommand) GetAPIKey(ctx context.Context) (string, error) {
	loadClientOauthRes, err := cliext.LoadClientOAuth(cliext.LoadClientOAuthOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to load login configuration: %w, please run `temporal cloud login --reset`", err)
	}

	// check if we have had a valid token in the past
	if loadClientOauthRes.OAuth == nil || loadClientOauthRes.OAuth.ClientConfig == nil {
		return "", fmt.Errorf("no login session found, please run `temporal cloud login`")
	}

	token, refreshed, err := GetToken(ctx, loadClientOauthRes.OAuth.ClientConfig, loadClientOauthRes.OAuth.Token)
	if err != nil {
		if errors.Is(err, ErrLoginRequired) {
			return "", fmt.Errorf("login session expired, please run `temporal cloud login`: %w", err)
		}
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	if refreshed {
		loadClientOauthRes.OAuth.Token = token
		if err := cliext.StoreClientOAuth(cliext.StoreClientOAuthOptions{
			OAuth: loadClientOauthRes.OAuth,
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
