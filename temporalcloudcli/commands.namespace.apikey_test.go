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

func TestNamespaceApiKeyGet(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceApiKeyGetCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Enabled",
			cmd:  temporalcloudcli.CloudNamespaceApiKeyGetCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-acct"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: &namespacev1.Namespace{
						Namespace: "my-ns.my-acct",
						Spec: &namespacev1.NamespaceSpec{
							ApiKeyAuth: &namespacev1.ApiKeyAuthSpec{Enabled: true},
						},
					}}, nil)
			},
			expectedJsonOutput: map[string]any{
				"namespace":         "my-ns.my-acct",
				"apiKeyAuthEnabled": true,
			},
		},
		{
			name: "Disabled",
			cmd:  temporalcloudcli.CloudNamespaceApiKeyGetCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-acct"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: &namespacev1.Namespace{
						Namespace: "my-ns.my-acct",
						Spec: &namespacev1.NamespaceSpec{
							ApiKeyAuth: &namespacev1.ApiKeyAuthSpec{Enabled: false},
						},
					}}, nil)
			},
			expectedJsonOutput: map[string]any{
				"namespace":         "my-ns.my-acct",
				"apiKeyAuthEnabled": false,
			},
		},
		{
			name: "NilApiKeyAuth",
			cmd:  temporalcloudcli.CloudNamespaceApiKeyGetCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-acct"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: &namespacev1.Namespace{
						Namespace: "my-ns.my-acct",
						Spec:      &namespacev1.NamespaceSpec{},
					}}, nil)
			},
			expectedJsonOutput: map[string]any{
				"namespace":         "my-ns.my-acct",
				"apiKeyAuthEnabled": false,
			},
		},
		{
			name: "GetNamespaceError",
			cmd:  temporalcloudcli.CloudNamespaceApiKeyGetCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
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

func TestNamespaceApiKeyEnable(t *testing.T) {
	existingNS := func(apiKeyAuth *namespacev1.ApiKeyAuthSpec) *namespacev1.Namespace {
		return &namespacev1.Namespace{
			Namespace:       "my-ns.my-acct",
			ResourceVersion: "rv-fetched",
			Spec: &namespacev1.NamespaceSpec{
				Name:       "my-ns",
				Regions:    []string{"aws-us-east-1"},
				ApiKeyAuth: apiKeyAuth,
			},
		}
	}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceApiKeyEnableCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "EnableFromNil",
			cmd:  temporalcloudcli.CloudNamespaceApiKeyEnableCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-acct"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNS(nil)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return req.Namespace == "my-ns.my-acct" &&
							req.ResourceVersion == "rv-fetched" &&
							proto.Equal(req.Spec.ApiKeyAuth, &namespacev1.ApiKeyAuthSpec{Enabled: true})
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-enable"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-enable"},
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudNamespaceApiKeyEnableCommand{
				NamespaceOptions:       temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-user"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNS(nil)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return req.ResourceVersion == "rv-user"
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-rv"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-rv"},
		},
		{
			name: "GetNamespaceError",
			cmd:  temporalcloudcli.CloudNamespaceApiKeyEnableCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("namespace not found"))
			},
			expectedErr: "namespace not found",
		},
		{
			name: "PromptDeclined",
			cmd:  temporalcloudcli.CloudNamespaceApiKeyEnableCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNS(nil)}, nil)
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

func TestNamespaceApiKeyDisable(t *testing.T) {
	existingNS := &namespacev1.Namespace{
		Namespace:       "my-ns.my-acct",
		ResourceVersion: "rv-fetched",
		Spec: &namespacev1.NamespaceSpec{
			Name:       "my-ns",
			Regions:    []string{"aws-us-east-1"},
			ApiKeyAuth: &namespacev1.ApiKeyAuthSpec{Enabled: true},
		},
	}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceApiKeyDisableCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "DisableFromEnabled",
			cmd:  temporalcloudcli.CloudNamespaceApiKeyDisableCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNS}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return proto.Equal(req.Spec.ApiKeyAuth, &namespacev1.ApiKeyAuthSpec{Enabled: false})
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-disable"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-disable"},
		},
		{
			name: "UpdateNamespaceError",
			cmd:  temporalcloudcli.CloudNamespaceApiKeyDisableCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNS}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("update failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			expectedErr:   "update failed",
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
