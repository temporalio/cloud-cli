package temporalcloudcli

import (
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	regionv1 "go.temporal.io/cloud-sdk/api/region/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudRegionGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetRegion(cctx, &cloudservice.GetRegionRequest{Region: c.Region})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintStructured(res.Region, printer.StructuredOptions{})
}

func (c *CloudRegionListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetRegions(cctx, &cloudservice.GetRegionsRequest{})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResourceList(
		struct {
			Regions []*regionv1.Region
		}{
			Regions: res.Regions,
		},
		printer.PrintResourceOptions{
			Fields: []string{"Id", "CloudProvider", "CloudProviderRegion", "Location"},
		},
		printer.TableOptions{},
	)
}
