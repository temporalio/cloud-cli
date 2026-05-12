package temporalcloudcli

import (
	"context"

	accountv1 "go.temporal.io/cloud-sdk/api/account/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	sinkv1 "go.temporal.io/cloud-sdk/api/sink/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type CreateAuditLogSinkPubSubParams struct {
	Name             string
	ServiceAccountID string
	TopicName        string
	GCPProjectID     string
	AsyncOperationID string

	Cloud            cloudservice.CloudServiceClient
	Prompter         Prompter
	OperationHandler AsyncOperationHandler
}

func CreateAuditLogSinkPubSub(ctx context.Context, params CreateAuditLogSinkPubSubParams) error {
	spec := &accountv1.AuditLogSinkSpec{
		Name:    params.Name,
		Enabled: true,
		SinkType: &accountv1.AuditLogSinkSpec_PubSubSink{
			PubSubSink: &sinkv1.PubSubSpec{
				ServiceAccountId: params.ServiceAccountID,
				TopicName:        params.TopicName,
				GcpProjectId:     params.GCPProjectID,
			},
		},
	}

	if err := params.Prompter.PromptApply(&accountv1.AuditLogSinkSpec{}, spec, false); err != nil {
		return err
	}

	createSink := wrapCreateOperation(
		params.Cloud.CreateAccountAuditLogSink,
		params.OperationHandler,
		func(res *cloudservice.CreateAccountAuditLogSinkResponse) string { return params.Name },
	)
	return createSink(ctx, &cloudservice.CreateAccountAuditLogSinkRequest{
		Spec:             spec,
		AsyncOperationId: params.AsyncOperationID,
	})
}

type UpdateAuditLogSinkPubSubParams struct {
	Name             string
	ServiceAccountID string
	TopicName        string
	GCPProjectID     string
	ResourceVersion  string
	AsyncOperationID string

	Cloud            cloudservice.CloudServiceClient
	Prompter         Prompter
	OperationHandler AsyncOperationHandler
}

func UpdateAuditLogSinkPubSub(ctx context.Context, params UpdateAuditLogSinkPubSubParams) error {
	res, err := params.Cloud.GetAccountAuditLogSink(ctx, &cloudservice.GetAccountAuditLogSinkRequest{Name: params.Name})
	if err != nil {
		return err
	}
	sink := res.Sink
	newSpec := proto.Clone(sink.Spec).(*accountv1.AuditLogSinkSpec)
	pubSub := newSpec.GetPubSubSink()
	if pubSub == nil {
		pubSub = &sinkv1.PubSubSpec{}
		newSpec.SinkType = &accountv1.AuditLogSinkSpec_PubSubSink{PubSubSink: pubSub}
	}
	if params.ServiceAccountID != "" {
		pubSub.ServiceAccountId = params.ServiceAccountID
	}
	if params.TopicName != "" {
		pubSub.TopicName = params.TopicName
	}
	if params.GCPProjectID != "" {
		pubSub.GcpProjectId = params.GCPProjectID
	}

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

type ValidateAuditLogSinkPubSubParams struct {
	ServiceAccountID string
	TopicName        string
	GCPProjectID     string

	Cloud   cloudservice.CloudServiceClient
	Printer *printer.Printer
}

func ValidateAuditLogSinkPubSub(ctx context.Context, params ValidateAuditLogSinkPubSubParams) error {
	spec := &accountv1.AuditLogSinkSpec{
		SinkType: &accountv1.AuditLogSinkSpec_PubSubSink{
			PubSubSink: &sinkv1.PubSubSpec{
				ServiceAccountId: params.ServiceAccountID,
				TopicName:        params.TopicName,
				GcpProjectId:     params.GCPProjectID,
			},
		},
	}
	if _, err := params.Cloud.ValidateAccountAuditLogSink(ctx, &cloudservice.ValidateAccountAuditLogSinkRequest{Spec: spec}); err != nil {
		return err
	}

	return params.Printer.PrintStructured(struct {
		Status string `json:"status"`
	}{
		Status: "valid",
	}, printer.StructuredOptions{})
}

func (c *CloudAccountAuditLogSinkPubsubValidateCommand) run(cctx *CommandContext, _ []string) error {
	saID, projectID, err := ParseServiceAccountEmail(c.ServiceAccountEmail)
	if err != nil {
		return err
	}
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return ValidateAuditLogSinkPubSub(cctx.Context, ValidateAuditLogSinkPubSubParams{
		ServiceAccountID: saID,
		TopicName:        c.TopicName,
		GCPProjectID:     projectID,
		Cloud:            cloudClient.CloudService(),
		Printer:          cctx.Printer,
	})
}

func (c *CloudAccountAuditLogSinkPubsubCreateCommand) run(cctx *CommandContext, _ []string) error {
	saID, projectID, err := ParseServiceAccountEmail(c.ServiceAccountEmail)
	if err != nil {
		return err
	}
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return CreateAuditLogSinkPubSub(cctx.Context, CreateAuditLogSinkPubSubParams{
		Name:             c.Name,
		ServiceAccountID: saID,
		TopicName:        c.TopicName,
		GCPProjectID:     projectID,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudAccountAuditLogSinkPubsubUpdateCommand) run(cctx *CommandContext, _ []string) error {
	// service-account-email is optional on update; if omitted, leave both fields
	// empty so UpdateAuditLogSinkPubSub treats them as "no change".
	var saID, projectID string
	if c.ServiceAccountEmail != "" {
		var err error
		saID, projectID, err = ParseServiceAccountEmail(c.ServiceAccountEmail)
		if err != nil {
			return err
		}
	}
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return UpdateAuditLogSinkPubSub(cctx.Context, UpdateAuditLogSinkPubSubParams{
		Name:             c.Name,
		ServiceAccountID: saID,
		TopicName:        c.TopicName,
		GCPProjectID:     projectID,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}
