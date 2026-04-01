package temporalcloudcli

import (
	"fmt"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNexusEndpointGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	// The singular GetNexusEndpoint RPC requires an endpoint ID, but this command takes a name.
	// Use the list RPC with a name filter instead.
	res, err := client.GetNexusEndpoints(cctx, &cloudservice.GetNexusEndpointsRequest{
		Name: c.Name,
	})
	if err != nil {
		return err
	}

	if len(res.Endpoints) == 0 {
		return fmt.Errorf("endpoint %q not found", c.Name)
	}

	return cctx.Printer.PrintResource(res.Endpoints[0], printer.PrintResourceOptions{})
}
