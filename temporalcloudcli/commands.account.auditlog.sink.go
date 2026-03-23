package temporalcloudcli

import (
	"context"

	accountv1 "go.temporal.io/cloud-sdk/api/account/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	"google.golang.org/protobuf/proto"

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

	EnableAuditLogSinkParams struct {
		Name             string
		ResourceVersion  string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	DisableAuditLogSinkParams struct {
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
			Fields:     []string{"Name", "State", "Health"},
			SpecFields: []string{"SinkType", "Enabled"},
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

func EnableAuditLogSink(ctx context.Context, params EnableAuditLogSinkParams) error {
	return setAuditLogSinkEnabled(ctx, setAuditLogSinkEnabledParams{
		Name:             params.Name,
		Enabled:          true,
		ResourceVersion:  params.ResourceVersion,
		AsyncOperationID: params.AsyncOperationID,
		Cloud:            params.Cloud,
		Prompter:         params.Prompter,
		OperationHandler: params.OperationHandler,
	})
}

func DisableAuditLogSink(ctx context.Context, params DisableAuditLogSinkParams) error {
	return setAuditLogSinkEnabled(ctx, setAuditLogSinkEnabledParams{
		Name:             params.Name,
		Enabled:          false,
		ResourceVersion:  params.ResourceVersion,
		AsyncOperationID: params.AsyncOperationID,
		Cloud:            params.Cloud,
		Prompter:         params.Prompter,
		OperationHandler: params.OperationHandler,
	})
}

type setAuditLogSinkEnabledParams struct {
	Name             string
	Enabled          bool
	ResourceVersion  string
	AsyncOperationID string

	Cloud            cloudservice.CloudServiceClient
	Prompter         Prompter
	OperationHandler AsyncOperationHandler
}

func setAuditLogSinkEnabled(ctx context.Context, params setAuditLogSinkEnabledParams) error {
	res, err := params.Cloud.GetAccountAuditLogSink(ctx, &cloudservice.GetAccountAuditLogSinkRequest{
		Name: params.Name,
	})
	if err != nil {
		return err
	}
	sink := res.GetSink()
	newSpec := proto.Clone(sink.Spec).(*accountv1.AuditLogSinkSpec)
	newSpec.Enabled = params.Enabled

	if err := params.Prompter.PromptApply(sink.Spec, newSpec, false); err != nil {
		return err
	}

	rv := sink.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}

	updateSink := wrapUpdateOperation(params.Cloud.UpdateAccountAuditLogSink, params.OperationHandler, params.Name)
	return updateSink(ctx, &cloudservice.UpdateAccountAuditLogSinkRequest{
		Spec:             newSpec,
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

func (c *CloudAccountAuditLogSinkEnableCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return EnableAuditLogSink(cctx.Context, EnableAuditLogSinkParams{
		Name:             c.Name,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudAccountAuditLogSinkDisableCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return DisableAuditLogSink(cctx.Context, DisableAuditLogSinkParams{
		Name:             c.Name,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}
