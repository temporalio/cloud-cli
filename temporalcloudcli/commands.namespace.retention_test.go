package temporalcloudcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
)

func newTestPrinter(buf *bytes.Buffer) *printer.Printer {
	return &printer.Printer{Output: buf, JSON: true}
}

// TestGetRetention_Success verifies that getRetention prints namespace and retentionDays.
func TestGetRetention_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	mockCloud.EXPECT().
		GetNamespace(context.Background(), &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{
			Namespace: &namespacev1.Namespace{
				Namespace: "my-namespace",
				Spec: &namespacev1.NamespaceSpec{
					RetentionDays: 14,
				},
			},
		}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.GetRetention(context.Background(), temporalcloudcli.GetRetentionParams{
		Namespace: "my-namespace",
		Cloud:     mockCloud,
		Printer:   newTestPrinter(&buf),
	})
	require.NoError(t, err)

	type retentionOutput struct {
		Namespace     string `json:"namespace"`
		RetentionDays int32  `json:"retentionDays"`
	}
	var out retentionOutput
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, retentionOutput{Namespace: "my-namespace", RetentionDays: 14}, out)
}

// TestGetRetention_Error verifies that a GetNamespace error propagates.
func TestGetRetention_Error(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	var buf bytes.Buffer
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetNamespace(context.Background(), &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}).
		Return(nil, apiErr)

	err := temporalcloudcli.GetRetention(context.Background(), temporalcloudcli.GetRetentionParams{
		Namespace: "my-namespace",
		Cloud:     mockCloud,
		Printer:   newTestPrinter(&buf),
	})
	require.ErrorIs(t, err, apiErr)
	assert.Empty(t, buf.String())
}

// TestSetRetention_Success verifies UpdateNamespace is called with the correct RetentionDays.
func TestSetRetention_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)

	mockCloud.EXPECT().
		GetNamespace(context.Background(), &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{
			Namespace: &namespacev1.Namespace{
				Namespace:       "my-namespace",
				ResourceVersion: "rv-1",
				Spec:            &namespacev1.NamespaceSpec{RetentionDays: 14},
			},
		}, nil)

	oldSpec := &namespacev1.NamespaceSpec{RetentionDays: 14}
	newSpec := &namespacev1.NamespaceSpec{RetentionDays: 30}

	mockPrompter.EXPECT().
		PromptApply(oldSpec, newSpec, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-123"}
	mockCloud.EXPECT().
		UpdateNamespace(context.Background(), &cloudservice.UpdateNamespaceRequest{
			Namespace:       "my-namespace",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: op,
		}, nil)

	mockRunner.EXPECT().
		Handle(op).
		Return(nil)

	err := temporalcloudcli.SetRetention(context.Background(), temporalcloudcli.SetRetentionParams{
		Namespace:        "my-namespace",
		RetentionDays:    30,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.NoError(t, err)
}

// TestSetRetention_PromptDeclined verifies UpdateNamespace is never called when prompt is declined.
func TestSetRetention_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	mockCloud.EXPECT().
		GetNamespace(context.Background(), &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{
			Namespace: &namespacev1.Namespace{
				Namespace:       "my-namespace",
				ResourceVersion: "rv-1",
				Spec:            &namespacev1.NamespaceSpec{RetentionDays: 14},
			},
		}, nil)

	oldSpec := &namespacev1.NamespaceSpec{RetentionDays: 14}
	newSpec := &namespacev1.NamespaceSpec{RetentionDays: 30}

	mockPrompter.EXPECT().
		PromptApply(oldSpec, newSpec, false).
		Return(promptErr)

	err := temporalcloudcli.SetRetention(context.Background(), temporalcloudcli.SetRetentionParams{
		Namespace:        "my-namespace",
		RetentionDays:    30,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.ErrorIs(t, err, promptErr)
}

// TestSetRetention_GetNamespaceError verifies that a GetNamespace error propagates and UpdateNamespace is never called.
func TestSetRetention_GetNamespaceError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetNamespace(context.Background(), &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}).
		Return(nil, apiErr)

	err := temporalcloudcli.SetRetention(context.Background(), temporalcloudcli.SetRetentionParams{
		Namespace:        "my-namespace",
		RetentionDays:    30,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.ErrorIs(t, err, apiErr)
}

// TestSetRetention_UpdateNamespaceError verifies that Runner.HandleErr receives the UpdateNamespace error.
func TestSetRetention_UpdateNamespaceError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	updateErr := errors.New("update error")

	mockCloud.EXPECT().
		GetNamespace(context.Background(), &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{
			Namespace: &namespacev1.Namespace{
				Namespace:       "my-namespace",
				ResourceVersion: "rv-1",
				Spec:            &namespacev1.NamespaceSpec{RetentionDays: 14},
			},
		}, nil)

	oldSpec := &namespacev1.NamespaceSpec{RetentionDays: 14}
	newSpec := &namespacev1.NamespaceSpec{RetentionDays: 30}

	mockPrompter.EXPECT().
		PromptApply(oldSpec, newSpec, false).
		Return(nil)

	mockCloud.EXPECT().
		UpdateNamespace(context.Background(), &cloudservice.UpdateNamespaceRequest{
			Namespace:       "my-namespace",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(nil, updateErr)

	mockRunner.EXPECT().
		HandleErr(updateErr).
		Return(updateErr)

	err := temporalcloudcli.SetRetention(context.Background(), temporalcloudcli.SetRetentionParams{
		Namespace:        "my-namespace",
		RetentionDays:    30,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.ErrorIs(t, err, updateErr)
}
