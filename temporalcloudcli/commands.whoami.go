package temporalcloudcli

import (
	"go.temporal.io/cloud-sdk/api/cloudservice/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

// run calls GetCurrentIdentity and prints the authenticated principal (User or
// ServiceAccount) along with the associated API key, if any.
func (c *CloudWhoamiCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := cloudClient.CloudService().GetCurrentIdentity(cctx.Context, &cloudservice.GetCurrentIdentityRequest{})
	if err != nil {
		return err
	}

	if cctx.Printer.JSON {
		return cctx.Printer.PrintStructured(res, printer.StructuredOptions{})
	}

	switch res.GetPrincipal().(type) {
	case *cloudservice.GetCurrentIdentityResponse_User:
		cctx.Printer.Println("Authenticated as User:")
		if err := cctx.Printer.PrintResource(res.GetUser(), printer.PrintResourceOptions{}); err != nil {
			return err
		}
	case *cloudservice.GetCurrentIdentityResponse_ServiceAccount:
		cctx.Printer.Println("Authenticated as Service Account:")
		if err := cctx.Printer.PrintResource(res.GetServiceAccount(), printer.PrintResourceOptions{}); err != nil {
			return err
		}
	default:
		cctx.Printer.Println("Authenticated with unknown principal type")
	}

	if res.GetPrincipalApiKey() != nil {
		cctx.Printer.Println("Using API Key:")
		if err := cctx.Printer.PrintResource(res.GetPrincipalApiKey(), printer.PrintResourceOptions{}); err != nil {
			return err
		}
	}
	return nil
}
