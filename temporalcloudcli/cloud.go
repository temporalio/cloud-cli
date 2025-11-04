package temporalcloudcli

import (
	"go.temporal.io/cloud-sdk/cloudclient"
)

func newCloudClient(cctx *CommandContext) (*cloudclient.Client, error) {
	opts := cloudclient.Options{}
	if cctx.RootCommand.ApiKey != "" {
		opts.APIKey = cctx.RootCommand.ApiKey
	} else {
		tokenConfig, err := loadTokenConfig(cctx)
		if err != nil {
			return nil, err
		}
		opts.APIKeyReader = tokenConfig
	}

	cloudClient, err := cloudclient.New(opts)
	if err != nil {
		return nil, err
	}

	return cloudClient, nil
}
