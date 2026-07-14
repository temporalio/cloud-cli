package temporalcloudcli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/temporalio/cli/cliext"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/log"
)

type CloudOptions struct {
	ApiKey string
	Server string
	cliext.CommonOptions
	Logger log.Logger
}

type CloudOptionsBuilder struct {
	// CommonOptions contains common CLI options including profile config.
	CommonOptions cliext.CommonOptions
	// ClientOptions contains the client configuration from flags.
	ClientOptions ClientOptions
	// EnvLookup is the environment variable lookup function.
	// If nil, environment variables are not used for profile loading.
	EnvLookup envconfig.EnvLookup
	// Logger is the slog logger to use for the client. If set, it will be
	// wrapped with the SDK's structured logger adapter.
	Logger *slog.Logger
}

func (b *CloudOptionsBuilder) Build(ctx context.Context) (*CloudOptions, error) {
	cfg := b.ClientOptions
	common := b.CommonOptions

	// Load a client config profile if configured
	var profile envconfig.ClientConfigProfile
	if !common.DisableConfigFile || !common.DisableConfigEnv {
		var err error
		profile, err = envconfig.LoadClientConfigProfile(envconfig.LoadClientConfigProfileOptions{
			ConfigFilePath:    common.ConfigFile,
			ConfigFileProfile: common.Profile,
			DisableFile:       common.DisableConfigFile,
			DisableEnv:        common.DisableConfigEnv,
			EnvLookup:         b.EnvLookup,
		})
		if err != nil {
			return nil, fmt.Errorf("failed loading client config: %w", err)
		}
	}

	cloudOpts := &CloudOptions{
		CommonOptions: common,
	}

	// Set logger if provided.
	if b.Logger != nil {
		cloudOpts.Logger = log.NewStructuredLogger(b.Logger)
	}

	// Set API key on profile if provided
	if cfg.ApiKey != "" {
		cloudOpts.ApiKey = cfg.ApiKey
	} else if profile.APIKey != "" {
		cloudOpts.ApiKey = profile.APIKey
	}

	if cfg.Server != "" {
		cloudOpts.Server = cfg.Server
	}

	return cloudOpts, nil
}

type (
	errSlot    struct{ err error }
	errSlotKey struct{}
)

// errorContext will add err to the errSlot on ctx if it has one before returning v and err
func errorContext[T any](ctx context.Context, v T, err error) (T, error) {
	if ctx != nil {
		if slot, ok := ctx.Value(errSlotKey{}).(*errSlot); ok {
			slot.err = err
		}
	}
	return v, err
}

func (c *CloudOptions) GetAPIKey(ctx context.Context) (string, error) {
	loadClientOauthRes, err := cliext.LoadClientOAuth(cliext.LoadClientOAuthOptions{
		ConfigFilePath: c.ConfigFile,
		ProfileName:    c.Profile,
		EnvLookup:      envconfig.EnvLookupOS,
	})
	if err != nil {
		return errorContext(ctx, "", NewFriendlyError("failed to load login configuration, please run `temporal cloud login --reset`", err))
	}

	// check if we have had a valid token in the past
	if loadClientOauthRes.OAuth == nil || loadClientOauthRes.OAuth.ClientConfig == nil {
		return errorContext(ctx, "", NewFriendlyError("no login session found, please run `temporal cloud login`", nil))
	}

	token, refreshed, err := GetToken(ctx, loadClientOauthRes.OAuth.ClientConfig, loadClientOauthRes.OAuth.Token)
	if err != nil {
		if errors.Is(err, ErrLoginRequired) {
			return errorContext(ctx, "", NewFriendlyError("login session expired, please run `temporal cloud login`", err))
		}
		return errorContext(ctx, "", NewFriendlyErrorf("failed to get access token: %v", err))
	}
	if refreshed {
		loadClientOauthRes.OAuth.Token = token
		if err := cliext.StoreClientOAuth(cliext.StoreClientOAuthOptions{
			OAuth:          loadClientOauthRes.OAuth,
			ConfigFilePath: c.ConfigFile,
			ProfileName:    c.Profile,
			EnvLookup:      envconfig.EnvLookupOS,
		}); err != nil {
			return errorContext(ctx, "", NewFriendlyErrorf("failed to write config file: %v", err))
		}
	}
	return token.AccessToken, nil
}
