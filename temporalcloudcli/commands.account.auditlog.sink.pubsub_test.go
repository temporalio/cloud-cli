package temporalcloudcli_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	accountv1 "go.temporal.io/cloud-sdk/api/account/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	operationv1 "go.temporal.io/cloud-sdk/api/operation/v1"
	sinkv1 "go.temporal.io/cloud-sdk/api/sink/v1"
)

func auditLogSink() *accountv1.AuditLogSink {
	return &accountv1.AuditLogSink{
		Name:            "my-sink",
		ResourceVersion: "rv-1",
		Spec:            auditLogSinkPubSubSpec(),
	}
}

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

// TestCreateAuditLogSinkPubSub_Success verifies CreateAccountAuditLogSink is called and HandleOperation receives the async op.
func TestCreateAuditLogSinkPubSub_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	spec := auditLogSinkPubSubSpec()
	op := &operationv1.AsyncOperation{Id: "op-123"}

	mockPrompter.EXPECT().
		PromptApply(&accountv1.AuditLogSinkSpec{}, spec, false).
		Return(nil)

	mockCloud.EXPECT().
		CreateAccountAuditLogSink(context.Background(), &cloudservice.CreateAccountAuditLogSinkRequest{Spec: spec}).
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
		PromptApply(&accountv1.AuditLogSinkSpec{}, spec, false).
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
		PromptApply(&accountv1.AuditLogSinkSpec{}, spec, false).
		Return(nil)

	mockCloud.EXPECT().
		CreateAccountAuditLogSink(context.Background(), &cloudservice.CreateAccountAuditLogSinkRequest{Spec: spec}).
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

// TestUpdateAuditLogSinkPubSub_Success verifies UpdateAccountAuditLogSink is called with the merged spec and HandleOperation receives the async op.
func TestUpdateAuditLogSinkPubSub_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	existing := auditLogSink()
	op := &operationv1.AsyncOperation{Id: "op-456"}

	oldSpec := auditLogSinkPubSubSpec()
	newSpec := auditLogSinkPubSubSpec()
	newSpec.GetPubSubSink().ServiceAccountId = "new-sa@project.iam.gserviceaccount.com"
	newSpec.GetPubSubSink().TopicName = "new-topic"

	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{Name: "my-sink"}).
		Return(&cloudservice.GetAccountAuditLogSinkResponse{Sink: existing}, nil)

	mockPrompter.EXPECT().
		PromptApply(oldSpec, newSpec, false).
		Return(nil)

	mockCloud.EXPECT().
		UpdateAccountAuditLogSink(context.Background(), &cloudservice.UpdateAccountAuditLogSinkRequest{
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateAccountAuditLogSinkResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().
		HandleOperation(op, "my-sink").
		Return(nil)

	err := temporalcloudcli.UpdateAuditLogSinkPubSub(context.Background(), temporalcloudcli.UpdateAuditLogSinkPubSubParams{
		Name:             "my-sink",
		ServiceAccountID: "new-sa@project.iam.gserviceaccount.com",
		TopicName:        "new-topic",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestUpdateAuditLogSinkPubSub_ResourceVersionOverride verifies an explicit --resource-version overrides the fetched one.
func TestUpdateAuditLogSinkPubSub_ResourceVersionOverride(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	existing := auditLogSink()
	op := &operationv1.AsyncOperation{Id: "op-789"}

	spec := auditLogSinkPubSubSpec()

	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{Name: "my-sink"}).
		Return(&cloudservice.GetAccountAuditLogSinkResponse{Sink: existing}, nil)

	mockPrompter.EXPECT().
		PromptApply(spec, spec, false).
		Return(nil)

	mockCloud.EXPECT().
		UpdateAccountAuditLogSink(context.Background(), &cloudservice.UpdateAccountAuditLogSinkRequest{
			Spec:            spec,
			ResourceVersion: "rv-override",
		}).
		Return(&cloudservice.UpdateAccountAuditLogSinkResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().
		HandleOperation(op, "my-sink").
		Return(nil)

	err := temporalcloudcli.UpdateAuditLogSinkPubSub(context.Background(), temporalcloudcli.UpdateAuditLogSinkPubSubParams{
		Name:            "my-sink",
		ResourceVersion: "rv-override",
		Cloud:           mockCloud,
		Prompter:        mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestUpdateAuditLogSinkPubSub_GetError verifies that a GetAccountAuditLogSink error is returned immediately.
func TestUpdateAuditLogSinkPubSub_GetError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("not found")

	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{Name: "my-sink"}).
		Return(nil, apiErr)

	err := temporalcloudcli.UpdateAuditLogSinkPubSub(context.Background(), temporalcloudcli.UpdateAuditLogSinkPubSubParams{
		Name:             "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

// TestUpdateAuditLogSinkPubSub_PromptDeclined verifies UpdateAccountAuditLogSink is never called when the prompt is declined.
func TestUpdateAuditLogSinkPubSub_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	existing := auditLogSink()
	spec := auditLogSinkPubSubSpec()

	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{Name: "my-sink"}).
		Return(&cloudservice.GetAccountAuditLogSinkResponse{Sink: existing}, nil)

	mockPrompter.EXPECT().
		PromptApply(spec, spec, false).
		Return(promptErr)

	err := temporalcloudcli.UpdateAuditLogSinkPubSub(context.Background(), temporalcloudcli.UpdateAuditLogSinkPubSubParams{
		Name:             "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, promptErr)
}

// TestUpdateAuditLogSinkPubSub_APIError verifies that an UpdateAccountAuditLogSink error is forwarded to HandleUpdateErr.
func TestUpdateAuditLogSinkPubSub_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("api error")

	existing := auditLogSink()
	spec := auditLogSinkPubSubSpec()

	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{Name: "my-sink"}).
		Return(&cloudservice.GetAccountAuditLogSinkResponse{Sink: existing}, nil)

	mockPrompter.EXPECT().
		PromptApply(spec, spec, false).
		Return(nil)

	mockCloud.EXPECT().
		UpdateAccountAuditLogSink(context.Background(), &cloudservice.UpdateAccountAuditLogSinkRequest{
			Spec:            spec,
			ResourceVersion: "rv-1",
		}).
		Return(nil, apiErr)

	mockHandler.EXPECT().
		HandleUpdateErr(apiErr).
		Return(apiErr)

	err := temporalcloudcli.UpdateAuditLogSinkPubSub(context.Background(), temporalcloudcli.UpdateAuditLogSinkPubSubParams{
		Name:             "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

// TestValidateAuditLogSinkPubSub_Success verifies ValidateAccountAuditLogSink is called and "Validation successful." is printed.
func TestValidateAuditLogSinkPubSub_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	spec := &accountv1.AuditLogSinkSpec{
		SinkType: &accountv1.AuditLogSinkSpec_PubSubSink{
			PubSubSink: &sinkv1.PubSubSpec{
				ServiceAccountId: "my-sa@project.iam.gserviceaccount.com",
				TopicName:        "my-topic",
				GcpProjectId:     "my-project",
			},
		},
	}

	mockCloud.EXPECT().
		ValidateAccountAuditLogSink(context.Background(), &cloudservice.ValidateAccountAuditLogSinkRequest{Spec: spec}).
		Return(&cloudservice.ValidateAccountAuditLogSinkResponse{}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.ValidateAuditLogSinkPubSub(context.Background(), temporalcloudcli.ValidateAuditLogSinkPubSubParams{
		ServiceAccountID: "my-sa@project.iam.gserviceaccount.com",
		TopicName:        "my-topic",
		GcpProjectID:     "my-project",
		Cloud:            mockCloud,
		Printer:          &printer.Printer{Output: &buf},
	})
	require.NoError(t, err)
	require.Contains(t, buf.String(), "Validation successful.")
}

// TestValidateAuditLogSinkPubSub_APIError verifies that a ValidateAccountAuditLogSink error is returned.
func TestValidateAuditLogSinkPubSub_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	apiErr := errors.New("validation failed")

	spec := &accountv1.AuditLogSinkSpec{
		SinkType: &accountv1.AuditLogSinkSpec_PubSubSink{
			PubSubSink: &sinkv1.PubSubSpec{
				ServiceAccountId: "my-sa@project.iam.gserviceaccount.com",
				TopicName:        "my-topic",
				GcpProjectId:     "my-project",
			},
		},
	}

	mockCloud.EXPECT().
		ValidateAccountAuditLogSink(context.Background(), &cloudservice.ValidateAccountAuditLogSinkRequest{Spec: spec}).
		Return(nil, apiErr)

	var buf bytes.Buffer
	err := temporalcloudcli.ValidateAuditLogSinkPubSub(context.Background(), temporalcloudcli.ValidateAuditLogSinkPubSubParams{
		ServiceAccountID: "my-sa@project.iam.gserviceaccount.com",
		TopicName:        "my-topic",
		GcpProjectID:     "my-project",
		Cloud:            mockCloud,
		Printer:          &printer.Printer{Output: &buf},
	})
	require.ErrorIs(t, err, apiErr)
}
