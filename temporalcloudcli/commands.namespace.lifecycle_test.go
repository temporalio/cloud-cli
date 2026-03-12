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

// TestGetLifecycle_Success verifies that GetLifecycle prints namespace and enableDeleteProtection.
func TestGetLifecycle_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	mockCloud.EXPECT().
		GetNamespace(context.Background(), &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{
			Namespace: &namespacev1.Namespace{
				Namespace: "my-namespace",
				Spec: &namespacev1.NamespaceSpec{
					Lifecycle: &namespacev1.LifecycleSpec{
						EnableDeleteProtection: true,
					},
				},
			},
		}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.GetLifecycle(context.Background(), temporalcloudcli.GetLifecycleParams{
		Namespace: "my-namespace",
		Cloud:     mockCloud,
		Printer:   &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)

	var out temporalcloudcli.GetLifecycleOutput
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, temporalcloudcli.GetLifecycleOutput{Namespace: "my-namespace", EnableDeleteProtection: true}, out)
}

// TestGetLifecycle_NilLifecycle verifies that GetLifecycle handles a nil lifecycle spec gracefully.
func TestGetLifecycle_NilLifecycle(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	mockCloud.EXPECT().
		GetNamespace(context.Background(), &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{
			Namespace: &namespacev1.Namespace{
				Namespace: "my-namespace",
				Spec:      &namespacev1.NamespaceSpec{},
			},
		}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.GetLifecycle(context.Background(), temporalcloudcli.GetLifecycleParams{
		Namespace: "my-namespace",
		Cloud:     mockCloud,
		Printer:   &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)

	type lifecycleOutput struct {
		Namespace              string `json:"namespace"`
		EnableDeleteProtection bool   `json:"enableDeleteProtection"`
	}
	var out lifecycleOutput
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, lifecycleOutput{Namespace: "my-namespace", EnableDeleteProtection: false}, out)
}

// TestGetLifecycle_Error verifies that a GetNamespace error propagates.
func TestGetLifecycle_Error(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	var buf bytes.Buffer
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetNamespace(context.Background(), &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}).
		Return(nil, apiErr)

	err := temporalcloudcli.GetLifecycle(context.Background(), temporalcloudcli.GetLifecycleParams{
		Namespace: "my-namespace",
		Cloud:     mockCloud,
		Printer:   &printer.Printer{Output: &buf, JSON: true},
	})
	require.ErrorIs(t, err, apiErr)
	assert.Empty(t, buf.String())
}

// TestSetLifecycle_Success verifies that UpdateNamespace is called with the correct EnableDeleteProtection.
func TestSetLifecycle_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)

	mockCloud.EXPECT().
		GetNamespace(context.Background(), &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{
			Namespace: &namespacev1.Namespace{
				Namespace:       "my-namespace",
				ResourceVersion: "rv-1",
				Spec: &namespacev1.NamespaceSpec{
					Lifecycle: &namespacev1.LifecycleSpec{
						EnableDeleteProtection: false,
					},
				},
			},
		}, nil)

	oldSpec := &namespacev1.NamespaceSpec{
		Lifecycle: &namespacev1.LifecycleSpec{EnableDeleteProtection: false},
	}
	newSpec := &namespacev1.NamespaceSpec{
		Lifecycle: &namespacev1.LifecycleSpec{EnableDeleteProtection: true},
	}

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

	err := temporalcloudcli.SetLifecycle(context.Background(), temporalcloudcli.SetLifecycleParams{
		Namespace:              "my-namespace",
		EnableDeleteProtection: true,
		Cloud:                  mockCloud,
		Prompter:               mockPrompter,
		OperationHandler:       mockRunner,
	})
	require.NoError(t, err)
}

// TestSetLifecycle_NilLifecycle verifies that SetLifecycle initializes a nil lifecycle spec.
func TestSetLifecycle_NilLifecycle(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)

	mockCloud.EXPECT().
		GetNamespace(context.Background(), &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{
			Namespace: &namespacev1.Namespace{
				Namespace:       "my-namespace",
				ResourceVersion: "rv-1",
				Spec:            &namespacev1.NamespaceSpec{},
			},
		}, nil)

	oldSpec := &namespacev1.NamespaceSpec{}
	newSpec := &namespacev1.NamespaceSpec{
		Lifecycle: &namespacev1.LifecycleSpec{EnableDeleteProtection: true},
	}

	mockPrompter.EXPECT().
		PromptApply(oldSpec, newSpec, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-456"}
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

	err := temporalcloudcli.SetLifecycle(context.Background(), temporalcloudcli.SetLifecycleParams{
		Namespace:              "my-namespace",
		EnableDeleteProtection: true,
		Cloud:                  mockCloud,
		Prompter:               mockPrompter,
		OperationHandler:       mockRunner,
	})
	require.NoError(t, err)
}

// TestSetLifecycle_PromptDeclined verifies UpdateNamespace is never called when prompt is declined.
func TestSetLifecycle_PromptDeclined(t *testing.T) {
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
				Spec: &namespacev1.NamespaceSpec{
					Lifecycle: &namespacev1.LifecycleSpec{EnableDeleteProtection: false},
				},
			},
		}, nil)

	oldSpec := &namespacev1.NamespaceSpec{
		Lifecycle: &namespacev1.LifecycleSpec{EnableDeleteProtection: false},
	}
	newSpec := &namespacev1.NamespaceSpec{
		Lifecycle: &namespacev1.LifecycleSpec{EnableDeleteProtection: true},
	}

	mockPrompter.EXPECT().
		PromptApply(oldSpec, newSpec, false).
		Return(promptErr)

	err := temporalcloudcli.SetLifecycle(context.Background(), temporalcloudcli.SetLifecycleParams{
		Namespace:              "my-namespace",
		EnableDeleteProtection: true,
		Cloud:                  mockCloud,
		Prompter:               mockPrompter,
		OperationHandler:       mockRunner,
	})
	require.ErrorIs(t, err, promptErr)
}

// TestSetLifecycle_GetNamespaceError verifies that a GetNamespace error propagates.
func TestSetLifecycle_GetNamespaceError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetNamespace(context.Background(), &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}).
		Return(nil, apiErr)

	err := temporalcloudcli.SetLifecycle(context.Background(), temporalcloudcli.SetLifecycleParams{
		Namespace:              "my-namespace",
		EnableDeleteProtection: true,
		Cloud:                  mockCloud,
		Prompter:               mockPrompter,
		OperationHandler:       mockRunner,
	})
	require.ErrorIs(t, err, apiErr)
}

// TestSetLifecycle_UpdateNamespaceError verifies that HandleErr receives the UpdateNamespace error.
func TestSetLifecycle_UpdateNamespaceError(t *testing.T) {
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
				Spec: &namespacev1.NamespaceSpec{
					Lifecycle: &namespacev1.LifecycleSpec{EnableDeleteProtection: false},
				},
			},
		}, nil)

	oldSpec := &namespacev1.NamespaceSpec{
		Lifecycle: &namespacev1.LifecycleSpec{EnableDeleteProtection: false},
	}
	newSpec := &namespacev1.NamespaceSpec{
		Lifecycle: &namespacev1.LifecycleSpec{EnableDeleteProtection: true},
	}

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

	err := temporalcloudcli.SetLifecycle(context.Background(), temporalcloudcli.SetLifecycleParams{
		Namespace:              "my-namespace",
		EnableDeleteProtection: true,
		Cloud:                  mockCloud,
		Prompter:               mockPrompter,
		OperationHandler:       mockRunner,
	})
	require.ErrorIs(t, err, updateErr)
}
