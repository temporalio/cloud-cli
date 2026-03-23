package temporalcloudcli

import (
	"context"

	accountv1 "go.temporal.io/cloud-sdk/api/account/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	sinkv1 "go.temporal.io/cloud-sdk/api/sink/v1"
)

type CreateAuditLogSinkPubSubParams struct {
	Name             string
	ServiceAccountID string
	TopicName        string
	GcpProjectID     string
	Enabled          bool
	AsyncOperationID string

	Cloud            cloudservice.CloudServiceClient
	Prompter         Prompter
	OperationHandler AsyncOperationHandler
}

func CreateAuditLogSinkPubSub(ctx context.Context, params CreateAuditLogSinkPubSubParams) error {
	spec := &accountv1.AuditLogSinkSpec{
		Name:    params.Name,
		Enabled: params.Enabled,
		SinkType: &accountv1.AuditLogSinkSpec_PubSubSink{
			PubSubSink: &sinkv1.PubSubSpec{
				ServiceAccountId: params.ServiceAccountID,
				TopicName:        params.TopicName,
				GcpProjectId:     params.GcpProjectID,
			},
		},
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

func (c *CloudAccountAuditLogSinkPubsubCreateCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return CreateAuditLogSinkPubSub(cctx.Context, CreateAuditLogSinkPubSubParams{
		Name:             c.Name,
		ServiceAccountID: c.ServiceAccountId,
		TopicName:        c.TopicName,
		GcpProjectID:     c.GcpProjectId,
		Enabled:          c.Enabled,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}
