package temporalcloudcli

import (
	"context"

	accountv1 "go.temporal.io/cloud-sdk/api/account/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type ListAuditLogSinksParams struct {
	PageSize  int32
	PageToken string

	Cloud   cloudservice.CloudServiceClient
	Printer *printer.Printer
}

func ListAuditLogSinks(ctx context.Context, params ListAuditLogSinksParams) error {
	res, err := params.Cloud.GetAccountAuditLogSinks(ctx, &cloudservice.GetAccountAuditLogSinksRequest{
		PageSize:  params.PageSize,
		PageToken: params.PageToken,
	})
	if err != nil {
		return err
	}
	return params.Printer.PrintResourceList(
		struct {
			Sinks         []*accountv1.AuditLogSink
			NextPageToken string
		}{
			Sinks:         res.GetSinks(),
			NextPageToken: res.GetNextPageToken(),
		},
		printer.PrintResourceOptions{
			Fields: []string{"Name", "State", "Health"},
		},
		printer.TableOptions{},
	)
}

func (c *CloudAccountAuditLogSinkListCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return ListAuditLogSinks(cctx.Context, ListAuditLogSinksParams{
		PageSize:  int32(c.PageSize),
		PageToken: c.PageToken,
		Cloud:     cloudClient.CloudService(),
		Printer:   cctx.Printer,
	})
}
