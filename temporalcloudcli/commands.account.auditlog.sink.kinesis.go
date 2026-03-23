package temporalcloudcli

import (
	"context"

	accountv1 "go.temporal.io/cloud-sdk/api/account/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	sinkv1 "go.temporal.io/cloud-sdk/api/sink/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type (
	CreateAuditLogSinkKinesisParams struct {
		Name             string
		RoleName         string
		DestinationURI   string
		Region           string
		Enabled          bool
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	UpdateAuditLogSinkKinesisParams struct {
		Name             string
		RoleName         string
		DestinationURI   string
		Region           string
		ResourceVersion  string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	ValidateAuditLogSinkKinesisParams struct {
		RoleName       string
		DestinationURI string
		Region         string

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
	}
)

func CreateAuditLogSinkKinesis(ctx context.Context, params CreateAuditLogSinkKinesisParams) error {
	spec := &accountv1.AuditLogSinkSpec{
		Name: params.Name,
		SinkType: &accountv1.AuditLogSinkSpec_KinesisSink{
			KinesisSink: &sinkv1.KinesisSpec{
				RoleName:       params.RoleName,
				DestinationUri: params.DestinationURI,
				Region:         params.Region,
			},
		},
		Enabled: params.Enabled,
	}

	if err := params.Prompter.PromptApply(&accountv1.AuditLogSinkSpec{}, spec, false); err != nil {
		return err
	}

	createSink := wrapCreateOperation(
		params.Cloud.CreateAccountAuditLogSink,
		params.OperationHandler,
		func(_ *cloudservice.CreateAccountAuditLogSinkResponse) string { return params.Name },
	)
	return createSink(ctx, &cloudservice.CreateAccountAuditLogSinkRequest{
		Spec:             spec,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func UpdateAuditLogSinkKinesis(ctx context.Context, params UpdateAuditLogSinkKinesisParams) error {
	res, err := params.Cloud.GetAccountAuditLogSink(ctx, &cloudservice.GetAccountAuditLogSinkRequest{
		Name: params.Name,
	})
	if err != nil {
		return err
	}
	sink := res.GetSink()
	newSpec := proto.Clone(sink.Spec).(*accountv1.AuditLogSinkSpec)

	// Apply non-empty string overrides; keep existing values for omitted flags.
	kinesis := proto.Clone(newSpec.GetKinesisSink()).(*sinkv1.KinesisSpec)
	if params.RoleName != "" {
		kinesis.RoleName = params.RoleName
	}
	if params.DestinationURI != "" {
		kinesis.DestinationUri = params.DestinationURI
	}
	if params.Region != "" {
		kinesis.Region = params.Region
	}
	newSpec.SinkType = &accountv1.AuditLogSinkSpec_KinesisSink{KinesisSink: kinesis}

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

func ValidateAuditLogSinkKinesis(ctx context.Context, params ValidateAuditLogSinkKinesisParams) error {
	spec := &accountv1.AuditLogSinkSpec{
		SinkType: &accountv1.AuditLogSinkSpec_KinesisSink{
			KinesisSink: &sinkv1.KinesisSpec{
				RoleName:       params.RoleName,
				DestinationUri: params.DestinationURI,
				Region:         params.Region,
			},
		},
	}

	_, err := params.Cloud.ValidateAccountAuditLogSink(ctx, &cloudservice.ValidateAccountAuditLogSinkRequest{
		Spec: spec,
	})
	if err != nil {
		return err
	}

	return params.Printer.PrintStructured(struct {
		Status string `json:"status"`
	}{
		Status: "valid",
	}, printer.StructuredOptions{})
}

func (c *CloudAccountAuditLogSinkKinesisCreateCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return CreateAuditLogSinkKinesis(cctx.Context, CreateAuditLogSinkKinesisParams{
		Name:             c.Name,
		RoleName:         c.RoleName,
		DestinationURI:   c.DestinationUri,
		Region:           c.Region,
		Enabled:          c.Enabled,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudAccountAuditLogSinkKinesisUpdateCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return UpdateAuditLogSinkKinesis(cctx.Context, UpdateAuditLogSinkKinesisParams{
		Name:             c.Name,
		RoleName:         c.RoleName,
		DestinationURI:   c.DestinationUri,
		Region:           c.Region,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudAccountAuditLogSinkKinesisValidateCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return ValidateAuditLogSinkKinesis(cctx.Context, ValidateAuditLogSinkKinesisParams{
		RoleName:       c.RoleName,
		DestinationURI: c.DestinationUri,
		Region:         c.Region,
		Cloud:          cloudClient.CloudService(),
		Printer:        cctx.Printer,
	})
}
