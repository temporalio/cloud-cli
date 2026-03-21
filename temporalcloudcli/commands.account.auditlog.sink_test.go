package temporalcloudcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	accountv1 "go.temporal.io/cloud-sdk/api/account/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAuditLogSinks_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	sinks := []*accountv1.AuditLogSink{
		{Name: "sink-1"},
		{Name: "sink-2"},
	}
	mockCloud.EXPECT().
		GetAccountAuditLogSinks(context.Background(), &cloudservice.GetAccountAuditLogSinksRequest{}).
		Return(&cloudservice.GetAccountAuditLogSinksResponse{
			Sinks:         sinks,
			NextPageToken: "next-token",
		}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.ListAuditLogSinks(context.Background(), temporalcloudcli.ListAuditLogSinksParams{
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)

	type listResponse struct {
		Sinks         []*accountv1.AuditLogSink `json:"sinks"`
		NextPageToken string                    `json:"nextPageToken"`
	}
	var out listResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, listResponse{
		Sinks:         sinks,
		NextPageToken: "next-token",
	}, out)
}

func TestListAuditLogSinks_WithPagination(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	mockCloud.EXPECT().
		GetAccountAuditLogSinks(context.Background(), &cloudservice.GetAccountAuditLogSinksRequest{
			PageSize:  50,
			PageToken: "some-token",
		}).
		Return(&cloudservice.GetAccountAuditLogSinksResponse{}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.ListAuditLogSinks(context.Background(), temporalcloudcli.ListAuditLogSinksParams{
		PageSize:  50,
		PageToken: "some-token",
		Cloud:     mockCloud,
		Printer:   &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)
}

func TestListAuditLogSinks_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetAccountAuditLogSinks(context.Background(), &cloudservice.GetAccountAuditLogSinksRequest{}).
		Return(nil, apiErr)

	var buf bytes.Buffer
	err := temporalcloudcli.ListAuditLogSinks(context.Background(), temporalcloudcli.ListAuditLogSinksParams{
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.ErrorIs(t, err, apiErr)
	assert.Empty(t, buf.String())
}

func TestDeleteAuditLogSink_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)

	sinkSpec := &accountv1.AuditLogSinkSpec{}
	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{
			Name: "my-sink",
		}).
		Return(&cloudservice.GetAccountAuditLogSinkResponse{
			Sink: &accountv1.AuditLogSink{
				Name:            "my-sink",
				ResourceVersion: "rv-1",
				Spec:            sinkSpec,
			},
		}, nil)

	mockPrompter.EXPECT().
		PromptApply(sinkSpec, &accountv1.AuditLogSinkSpec{}, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-1"}
	mockCloud.EXPECT().
		DeleteAccountAuditLogSink(context.Background(), &cloudservice.DeleteAccountAuditLogSinkRequest{
			Name:            "my-sink",
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.DeleteAccountAuditLogSinkResponse{
			AsyncOperation: op,
		}, nil)

	mockHandler.EXPECT().HandleOperation(op, "my-sink").Return(nil)

	err := temporalcloudcli.DeleteAuditLogSink(context.Background(), temporalcloudcli.DeleteAuditLogSinkParams{
		Name:             "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestDeleteAuditLogSink_GetSinkError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	apiErr := errors.New("not found")

	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{
			Name: "my-sink",
		}).
		Return(nil, apiErr)

	mockHandler.EXPECT().HandleDeleteErr(apiErr).Return(apiErr)

	err := temporalcloudcli.DeleteAuditLogSink(context.Background(), temporalcloudcli.DeleteAuditLogSinkParams{
		Name:             "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

func TestDeleteAuditLogSink_DeleteError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	deleteErr := errors.New("delete error")

	sinkSpec := &accountv1.AuditLogSinkSpec{}
	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{
			Name: "my-sink",
		}).
		Return(&cloudservice.GetAccountAuditLogSinkResponse{
			Sink: &accountv1.AuditLogSink{
				Name:            "my-sink",
				ResourceVersion: "rv-1",
				Spec:            sinkSpec,
			},
		}, nil)

	mockPrompter.EXPECT().
		PromptApply(sinkSpec, &accountv1.AuditLogSinkSpec{}, false).
		Return(nil)

	mockCloud.EXPECT().
		DeleteAccountAuditLogSink(context.Background(), &cloudservice.DeleteAccountAuditLogSinkRequest{
			Name:            "my-sink",
			ResourceVersion: "rv-1",
		}).
		Return(nil, deleteErr)

	mockHandler.EXPECT().HandleDeleteErr(deleteErr).Return(deleteErr)

	err := temporalcloudcli.DeleteAuditLogSink(context.Background(), temporalcloudcli.DeleteAuditLogSinkParams{
		Name:             "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, deleteErr)
}

func TestEnableAuditLogSink_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)

	sinkSpec := &accountv1.AuditLogSinkSpec{Enabled: false}
	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{
			Name: "my-sink",
		}).
		Return(&cloudservice.GetAccountAuditLogSinkResponse{
			Sink: &accountv1.AuditLogSink{
				Name:            "my-sink",
				ResourceVersion: "rv-1",
				Spec:            sinkSpec,
			},
		}, nil)

	enabledSpec := &accountv1.AuditLogSinkSpec{Enabled: true}
	mockPrompter.EXPECT().
		PromptApply(sinkSpec, enabledSpec, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-1"}
	mockCloud.EXPECT().
		UpdateAccountAuditLogSink(context.Background(), &cloudservice.UpdateAccountAuditLogSinkRequest{
			Spec:            enabledSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateAccountAuditLogSinkResponse{
			AsyncOperation: op,
		}, nil)

	mockHandler.EXPECT().HandleOperation(op, "my-sink").Return(nil)

	err := temporalcloudcli.EnableAuditLogSink(context.Background(), temporalcloudcli.EnableAuditLogSinkParams{
		Name:             "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestEnableAuditLogSink_GetSinkError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	apiErr := errors.New("not found")

	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{
			Name: "my-sink",
		}).
		Return(nil, apiErr)

	err := temporalcloudcli.EnableAuditLogSink(context.Background(), temporalcloudcli.EnableAuditLogSinkParams{
		Name:             "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

func TestEnableAuditLogSink_UpdateError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	updateErr := errors.New("update error")

	sinkSpec := &accountv1.AuditLogSinkSpec{Enabled: false}
	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{
			Name: "my-sink",
		}).
		Return(&cloudservice.GetAccountAuditLogSinkResponse{
			Sink: &accountv1.AuditLogSink{
				Name:            "my-sink",
				ResourceVersion: "rv-1",
				Spec:            sinkSpec,
			},
		}, nil)

	enabledSpec := &accountv1.AuditLogSinkSpec{Enabled: true}
	mockPrompter.EXPECT().
		PromptApply(sinkSpec, enabledSpec, false).
		Return(nil)

	mockCloud.EXPECT().
		UpdateAccountAuditLogSink(context.Background(), &cloudservice.UpdateAccountAuditLogSinkRequest{
			Spec:            enabledSpec,
			ResourceVersion: "rv-1",
		}).
		Return(nil, updateErr)

	mockHandler.EXPECT().HandleUpdateErr(updateErr).Return(updateErr)

	err := temporalcloudcli.EnableAuditLogSink(context.Background(), temporalcloudcli.EnableAuditLogSinkParams{
		Name:             "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, updateErr)
}

func TestEnableAuditLogSink_CustomResourceVersion(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)

	sinkSpec := &accountv1.AuditLogSinkSpec{Enabled: false}
	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{
			Name: "my-sink",
		}).
		Return(&cloudservice.GetAccountAuditLogSinkResponse{
			Sink: &accountv1.AuditLogSink{
				Name:            "my-sink",
				ResourceVersion: "rv-1",
				Spec:            sinkSpec,
			},
		}, nil)

	enabledSpec := &accountv1.AuditLogSinkSpec{Enabled: true}
	mockPrompter.EXPECT().
		PromptApply(sinkSpec, enabledSpec, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-2"}
	mockCloud.EXPECT().
		UpdateAccountAuditLogSink(context.Background(), &cloudservice.UpdateAccountAuditLogSinkRequest{
			Spec:            enabledSpec,
			ResourceVersion: "rv-explicit",
		}).
		Return(&cloudservice.UpdateAccountAuditLogSinkResponse{
			AsyncOperation: op,
		}, nil)

	mockHandler.EXPECT().HandleOperation(op, "my-sink").Return(nil)

	err := temporalcloudcli.EnableAuditLogSink(context.Background(), temporalcloudcli.EnableAuditLogSinkParams{
		Name:             "my-sink",
		ResourceVersion:  "rv-explicit",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestDisableAuditLogSink_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)

	sinkSpec := &accountv1.AuditLogSinkSpec{Enabled: true}
	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{
			Name: "my-sink",
		}).
		Return(&cloudservice.GetAccountAuditLogSinkResponse{
			Sink: &accountv1.AuditLogSink{
				Name:            "my-sink",
				ResourceVersion: "rv-1",
				Spec:            sinkSpec,
			},
		}, nil)

	disabledSpec := &accountv1.AuditLogSinkSpec{Enabled: false}
	mockPrompter.EXPECT().
		PromptApply(sinkSpec, disabledSpec, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-1"}
	mockCloud.EXPECT().
		UpdateAccountAuditLogSink(context.Background(), &cloudservice.UpdateAccountAuditLogSinkRequest{
			Spec:            disabledSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateAccountAuditLogSinkResponse{
			AsyncOperation: op,
		}, nil)

	mockHandler.EXPECT().HandleOperation(op, "my-sink").Return(nil)

	err := temporalcloudcli.DisableAuditLogSink(context.Background(), temporalcloudcli.DisableAuditLogSinkParams{
		Name:             "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestDisableAuditLogSink_GetSinkError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	apiErr := errors.New("not found")

	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{
			Name: "my-sink",
		}).
		Return(nil, apiErr)

	err := temporalcloudcli.DisableAuditLogSink(context.Background(), temporalcloudcli.DisableAuditLogSinkParams{
		Name:             "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

func TestDeleteAuditLogSink_CustomResourceVersion(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	mockPrompter := cmdmock.NewMockPrompter(t)

	sinkSpec := &accountv1.AuditLogSinkSpec{}
	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{
			Name: "my-sink",
		}).
		Return(&cloudservice.GetAccountAuditLogSinkResponse{
			Sink: &accountv1.AuditLogSink{
				Name:            "my-sink",
				ResourceVersion: "rv-1",
				Spec:            sinkSpec,
			},
		}, nil)

	mockPrompter.EXPECT().
		PromptApply(sinkSpec, &accountv1.AuditLogSinkSpec{}, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-2"}
	mockCloud.EXPECT().
		DeleteAccountAuditLogSink(context.Background(), &cloudservice.DeleteAccountAuditLogSinkRequest{
			Name:            "my-sink",
			ResourceVersion: "rv-explicit",
		}).
		Return(&cloudservice.DeleteAccountAuditLogSinkResponse{
			AsyncOperation: op,
		}, nil)

	mockHandler.EXPECT().HandleOperation(op, "my-sink").Return(nil)

	err := temporalcloudcli.DeleteAuditLogSink(context.Background(), temporalcloudcli.DeleteAuditLogSinkParams{
		Name:             "my-sink",
		ResourceVersion:  "rv-explicit",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}
