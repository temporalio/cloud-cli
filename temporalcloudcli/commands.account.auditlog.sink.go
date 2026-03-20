package temporalcloudcli

import (
	"context"

	accountv1 "go.temporal.io/cloud-sdk/api/account/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type (
	ListAuditLogSinksParams struct {
		PageSize  int32
		PageToken string

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
	}

	DeleteAuditLogSinkParams struct {
		Name             string
		ResourceVersion  string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}
)

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

func DeleteAuditLogSink(ctx context.Context, params DeleteAuditLogSinkParams) error {
	res, err := params.Cloud.GetAccountAuditLogSink(ctx, &cloudservice.GetAccountAuditLogSinkRequest{
		Name: params.Name,
	})
	if err != nil {
		return params.OperationHandler.HandleDeleteErr(err)
	}
	sink := res.GetSink()

	if err := params.Prompter.PromptApply(sink.Spec, &accountv1.AuditLogSinkSpec{}, false); err != nil {
		return err
	}

	rv := sink.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}

	deleteSink := wrapDeleteOperation(params.Cloud.DeleteAccountAuditLogSink, params.OperationHandler, params.Name)
	return deleteSink(ctx, &cloudservice.DeleteAccountAuditLogSinkRequest{
		Name:             params.Name,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
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

func (c *CloudAccountAuditLogSinkDeleteCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return DeleteAuditLogSink(cctx.Context, DeleteAuditLogSinkParams{
		Name:             c.Name,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}
