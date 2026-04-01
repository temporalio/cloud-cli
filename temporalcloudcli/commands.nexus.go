package temporalcloudcli

import (
	"fmt"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	nexusv1 "go.temporal.io/cloud-sdk/api/nexus/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNexusEndpointListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	var endpoints []*nexusv1.Endpoint
	pageToken := ""
	for {
		res, err := client.GetNexusEndpoints(cctx, &cloudservice.GetNexusEndpointsRequest{
			PageToken: pageToken,
		})
		if err != nil {
			return err
		}
		endpoints = append(endpoints, res.Endpoints...)
		pageToken = res.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return cctx.Printer.PrintResourceList(
		struct {
			Endpoints []*nexusv1.Endpoint
		}{
			Endpoints: endpoints,
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Id", "State"},
			SpecFields: []string{"Name"},
		},
		printer.TableOptions{},
	)
}

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
