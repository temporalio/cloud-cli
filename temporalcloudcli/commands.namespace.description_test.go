package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/protobuf/proto"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

func TestNamespaceDescriptionGet(t *testing.T) {
	getNSReq := &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-acct"}
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceDescriptionGetCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Populated",
			cmd:  temporalcloudcli.CloudNamespaceDescriptionGetCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, getNSReq, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: &namespacev1.Namespace{
						Namespace: "my-ns.my-acct",
						Spec: &namespacev1.NamespaceSpec{
							Description: "Example namespace description",
						},
					}}, nil)
			},
			expectedJsonOutput: map[string]any{
				"namespace":   "my-ns.my-acct",
				"description": "Example namespace description",
			},
		},
		{
			name: "Empty",
			cmd:  temporalcloudcli.CloudNamespaceDescriptionGetCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, getNSReq, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: &namespacev1.Namespace{
						Namespace: "my-ns.my-acct",
						Spec:      &namespacev1.NamespaceSpec{},
					}}, nil)
			},
			expectedJsonOutput: map[string]any{
				"namespace":   "my-ns.my-acct",
				"description": "",
			},
		},
		{
			name: "GetNamespaceError",
			cmd:  temporalcloudcli.CloudNamespaceDescriptionGetCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, getNSReq, mock.Anything).
					Return(nil, errors.New("namespace not found"))
			},
			expectedErr: "namespace not found",
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

func TestNamespaceDescriptionSet(t *testing.T) {
	getNSReq := &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-acct"}
	existingNS := func(description string) *namespacev1.Namespace {
		return &namespacev1.Namespace{
			Namespace:       "my-ns.my-acct",
			ResourceVersion: "rv-fetched",
			Spec: &namespacev1.NamespaceSpec{
				Name:          "my-ns",
				Regions:       []string{"aws-us-east-1"},
				RetentionDays: 30,
				Description:   description,
			},
		}
	}
	matchUpdate := func(currentDescription, newDescription, rv, asyncID string) any {
		wantSpec := proto.Clone(existingNS(currentDescription).Spec).(*namespacev1.NamespaceSpec)
		wantSpec.Description = newDescription
		want := &cloudservice.UpdateNamespaceRequest{
			Namespace:        "my-ns.my-acct",
			Spec:             wantSpec,
			ResourceVersion:  rv,
			AsyncOperationId: asyncID,
		}
		return mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			return proto.Equal(req, want)
		})
	}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceDescriptionSetCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Set",
			cmd: temporalcloudcli.CloudNamespaceDescriptionSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				Value:            "Updated namespace description",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, getNSReq, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNS("")}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, matchUpdate("", "Updated namespace description", "rv-fetched", ""), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-set"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-set"},
		},
		{
			name: "Clear",
			cmd: temporalcloudcli.CloudNamespaceDescriptionSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				Value:            "",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, getNSReq, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNS("old")}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, matchUpdate("old", "", "rv-fetched", ""), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-clear"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-clear"},
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudNamespaceDescriptionSetCommand{
				NamespaceOptions:       temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				Value:                  "Updated namespace description",
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-user"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, getNSReq, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNS("")}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, matchUpdate("", "Updated namespace description", "rv-user", ""), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-rv"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-rv"},
		},
		{
			name: "AsyncOperationIdOverride",
			cmd: temporalcloudcli.CloudNamespaceDescriptionSetCommand{
				NamespaceOptions:      temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				Value:                 "Updated namespace description",
				AsyncOperationOptions: temporalcloudcli.AsyncOperationOptions{AsyncOperationId: "op-custom"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, getNSReq, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNS("")}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, matchUpdate("", "Updated namespace description", "rv-fetched", "op-custom"), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-custom"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-custom"},
		},
		{
			name: "GetNamespaceError",
			cmd: temporalcloudcli.CloudNamespaceDescriptionSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				Value:            "Updated namespace description",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, getNSReq, mock.Anything).
					Return(nil, errors.New("namespace not found"))
			},
			expectedErr: "namespace not found",
		},
		{
			name: "UpdateNamespaceError",
			cmd: temporalcloudcli.CloudNamespaceDescriptionSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				Value:            "Updated namespace description",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, getNSReq, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNS("")}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, matchUpdate("", "Updated namespace description", "rv-fetched", ""), mock.Anything).
					Return(nil, errors.New("update failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			expectedErr:   "update failed",
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudNamespaceDescriptionSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				Value:            "Updated namespace description",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, getNSReq, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNS("")}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}
