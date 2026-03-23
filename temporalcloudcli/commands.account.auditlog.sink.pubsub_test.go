package temporalcloudcli_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	accountv1 "go.temporal.io/cloud-sdk/api/account/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	operationv1 "go.temporal.io/cloud-sdk/api/operation/v1"
	sinkv1 "go.temporal.io/cloud-sdk/api/sink/v1"
	"google.golang.org/protobuf/proto"
)

func auditLogSinkPubSubSpec() *accountv1.AuditLogSinkSpec {
	return &accountv1.AuditLogSinkSpec{
		Name:    "my-sink",
		Enabled: true,
		SinkType: &accountv1.AuditLogSinkSpec_PubSubSink{
			PubSubSink: &sinkv1.PubSubSpec{
				ServiceAccountId: "my-sa@project.iam.gserviceaccount.com",
				TopicName:        "my-topic",
				GcpProjectId:     "my-project",
			},
		},
	}
}

func createSinkReqMatcher(expected *accountv1.AuditLogSinkSpec) interface{} {
	return mock.MatchedBy(func(req *cloudservice.CreateAccountAuditLogSinkRequest) bool {
		return proto.Equal(req.Spec, expected)
	})
}

// TestCreateAuditLogSinkPubSub_Success verifies CreateAccountAuditLogSink is called and HandleOperation receives the async op.
func TestCreateAuditLogSinkPubSub_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	spec := auditLogSinkPubSubSpec()
	op := &operationv1.AsyncOperation{Id: "op-123"}

	mockPrompter.EXPECT().
		PromptApply(&accountv1.AuditLogSinkSpec{}, mock.MatchedBy(func(s proto.Message) bool { return proto.Equal(s, spec) }), false).
		Return(nil)

	mockCloud.EXPECT().
		CreateAccountAuditLogSink(context.Background(), createSinkReqMatcher(spec)).
		Return(&cloudservice.CreateAccountAuditLogSinkResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().
		HandleOperation(op, "my-sink").
		Return(nil)

	err := temporalcloudcli.CreateAuditLogSinkPubSub(context.Background(), temporalcloudcli.CreateAuditLogSinkPubSubParams{
		Name:             "my-sink",
		ServiceAccountID: "my-sa@project.iam.gserviceaccount.com",
		TopicName:        "my-topic",
		GcpProjectID:     "my-project",
		Enabled:          true,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestCreateAuditLogSinkPubSub_PromptDeclined verifies CreateAccountAuditLogSink is never called when the prompt is declined.
func TestCreateAuditLogSinkPubSub_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	spec := auditLogSinkPubSubSpec()

	mockPrompter.EXPECT().
		PromptApply(&accountv1.AuditLogSinkSpec{}, mock.MatchedBy(func(s proto.Message) bool { return proto.Equal(s, spec) }), false).
		Return(promptErr)

	err := temporalcloudcli.CreateAuditLogSinkPubSub(context.Background(), temporalcloudcli.CreateAuditLogSinkPubSubParams{
		Name:             "my-sink",
		ServiceAccountID: "my-sa@project.iam.gserviceaccount.com",
		TopicName:        "my-topic",
		GcpProjectID:     "my-project",
		Enabled:          true,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, promptErr)
}

// TestCreateAuditLogSinkPubSub_APIError verifies that a CreateAccountAuditLogSink error is forwarded to HandleCreateErr.
func TestCreateAuditLogSinkPubSub_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("api error")

	spec := auditLogSinkPubSubSpec()

	mockPrompter.EXPECT().
		PromptApply(&accountv1.AuditLogSinkSpec{}, mock.MatchedBy(func(s proto.Message) bool { return proto.Equal(s, spec) }), false).
		Return(nil)

	mockCloud.EXPECT().
		CreateAccountAuditLogSink(context.Background(), createSinkReqMatcher(spec)).
		Return(nil, apiErr)

	mockHandler.EXPECT().
		HandleCreateErr(apiErr).
		Return(apiErr)

	err := temporalcloudcli.CreateAuditLogSinkPubSub(context.Background(), temporalcloudcli.CreateAuditLogSinkPubSubParams{
		Name:             "my-sink",
		ServiceAccountID: "my-sa@project.iam.gserviceaccount.com",
		TopicName:        "my-topic",
		GcpProjectID:     "my-project",
		Enabled:          true,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}
