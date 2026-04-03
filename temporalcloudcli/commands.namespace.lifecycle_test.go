package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

func TestGetLifecycle(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceLifecycleGetCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Success_WithDeleteProtection",
			cmd:  temporalcloudcli.CloudNamespaceLifecycleGetCommand{Namespace: "my-namespace"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace: "my-namespace",
							Spec: &namespacev1.NamespaceSpec{
								Lifecycle: &namespacev1.LifecycleSpec{EnableDeleteProtection: true},
							},
						},
					}, nil)
			},
			expectedJsonOutput: map[string]any{
				"namespace":              "my-namespace",
				"enableDeleteProtection": true,
			},
		},
		{
			name: "Success_NoLifecycleSpec",
			cmd:  temporalcloudcli.CloudNamespaceLifecycleGetCommand{Namespace: "my-namespace"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace: "my-namespace",
							Spec:      &namespacev1.NamespaceSpec{},
						},
					}, nil)
			},
			expectedJsonOutput: map[string]any{
				"namespace":              "my-namespace",
				"enableDeleteProtection": false,
			},
		},
		{
			name: "GetNamespaceError",
			cmd:  temporalcloudcli.CloudNamespaceLifecycleGetCommand{Namespace: "my-namespace"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("api error"))
			},
			expectedErr: "api error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedJsonOutput,
			})
		})
	}
}

func TestSetLifecycle_Success(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceLifecycleSetCommand{
		Namespace:              "my-namespace",
		EnableDeleteProtection: true,
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{
					Namespace: &namespacev1.Namespace{
						Namespace:       "my-namespace",
						ResourceVersion: "rv-1",
						Spec:            &namespacev1.NamespaceSpec{},
					},
				}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, &cloudservice.UpdateNamespaceRequest{
					Namespace: "my-namespace",
					Spec: &namespacev1.NamespaceSpec{
						Lifecycle: &namespacev1.LifecycleSpec{EnableDeleteProtection: true},
					},
					ResourceVersion: "rv-1",
				}, mock.Anything).
				Return(&cloudservice.UpdateNamespaceResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-123"},
				}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply: true,
			PromptResult:     true,
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-123"},
		JSONOutput:         true,
		ExpectedOutputJson: &cloudservice.UpdateNamespaceResponse{
			AsyncOperation: &operation.AsyncOperation{
				Id:    "op-123",
				State: operation.AsyncOperation_STATE_FULFILLED,
			},
		},
	})
}

func TestSetLifecycle_PromptDeclined(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceLifecycleSetCommand{
		Namespace:              "my-namespace",
		EnableDeleteProtection: true,
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{
					Namespace: &namespacev1.Namespace{
						Namespace:       "my-namespace",
						ResourceVersion: "rv-1",
						Spec:            &namespacev1.NamespaceSpec{},
					},
				}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply: true,
			PromptResult:     false,
		},
		ExpectedError: "Aborting update.",
	})
}

func TestSetLifecycle_GetNamespaceError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceLifecycleSetCommand{
		Namespace:              "my-namespace",
		EnableDeleteProtection: true,
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("api error"))
		},
		ExpectedError: "api error",
	})
}

func TestSetLifecycle_UpdateNamespaceError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceLifecycleSetCommand{
		Namespace:              "my-namespace",
		EnableDeleteProtection: true,
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-namespace"}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{
					Namespace: &namespacev1.Namespace{
						Namespace:       "my-namespace",
						ResourceVersion: "rv-1",
						Spec:            &namespacev1.NamespaceSpec{},
					},
				}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("update error"))
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply: true,
			PromptResult:     true,
		},
		ExpectedError: "update operation failed",
	})
}
