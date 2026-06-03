package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

// --- ListTags ---

func TestListTags(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceTagListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudNamespaceTagListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace.test-account"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace: "test-namespace.test-account",
							Tags: map[string]string{
								"team":        "platform",
								"environment": "production",
							},
						},
					}, nil)
			},
			expectedJsonOutput: struct {
				Tags []temporalcloudcli.Tag
			}{
				Tags: []temporalcloudcli.Tag{
					{Key: "environment", Value: "production"},
					{Key: "team", Value: "platform"},
				},
			},
		},
		{
			name: "EmptyList",
			cmd:  temporalcloudcli.CloudNamespaceTagListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace.test-account"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{Namespace: "test-namespace.test-account"},
					}, nil)
			},
			expectedJsonOutput: struct {
				Tags []temporalcloudcli.Tag
			}{Tags: []temporalcloudcli.Tag{}},
		},
		{
			name: "GetNamespaceError",
			cmd:  temporalcloudcli.CloudNamespaceTagListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("API error"))
			},
			expectedErr: "API error",
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

// --- CreateTag ---

func TestCreateTag(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceTagCreateCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudNamespaceTagCreateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				Key:              "environment",
				Value:            "production",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, &cloudservice.UpdateNamespaceTagsRequest{
						Namespace:    "test-namespace.test-account",
						TagsToUpsert: map[string]string{"environment": "production"},
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceTagsResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-create"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, ExpectPromptYesMessage: "Create", PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-create"},
		},
		{
			name: "AsyncOperationIdOverride",
			cmd: temporalcloudcli.CloudNamespaceTagCreateCommand{
				NamespaceOptions:      temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				AsyncOperationOptions: temporalcloudcli.AsyncOperationOptions{AsyncOperationId: "custom-op-id"},
				Key:                   "environment",
				Value:                 "production",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, &cloudservice.UpdateNamespaceTagsRequest{
						Namespace:        "test-namespace.test-account",
						TagsToUpsert:     map[string]string{"environment": "production"},
						AsyncOperationId: "custom-op-id",
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceTagsResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "custom-op-id"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "custom-op-id"},
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudNamespaceTagCreateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				Key:              "environment",
				Value:            "production",
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting create.",
		},
		{
			name: "ApiError",
			cmd: temporalcloudcli.CloudNamespaceTagCreateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				Key:              "environment",
				Value:            "production",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("API error"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedErr:   "create operation failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- UpdateTag ---

func TestUpdateTag(t *testing.T) {
	existingNamespace := &namespacev1.Namespace{
		Namespace: "test-namespace.test-account",
		Tags:      map[string]string{"environment": "production"},
	}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceTagUpdateCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudNamespaceTagUpdateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				Key:              "environment",
				Value:            "staging",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace.test-account"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, &cloudservice.UpdateNamespaceTagsRequest{
						Namespace:    "test-namespace.test-account",
						TagsToUpsert: map[string]string{"environment": "staging"},
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceTagsResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-update"},
					}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{
				ExpectPrompApply:          true,
				ExpectPromptApplyExisting: &namespacev1.Namespace{Tags: map[string]string{"environment": "production"}},
				ExpectPromptApplyModified: &namespacev1.Namespace{Tags: map[string]string{"environment": "staging"}},
				PromptResult:              true,
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-update"},
		},
		{
			name: "AsyncOperationIdOverride",
			cmd: temporalcloudcli.CloudNamespaceTagUpdateCommand{
				NamespaceOptions:      temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				AsyncOperationOptions: temporalcloudcli.AsyncOperationOptions{AsyncOperationId: "custom-op-id"},
				Key:                   "environment",
				Value:                 "staging",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, &cloudservice.UpdateNamespaceTagsRequest{
						Namespace:        "test-namespace.test-account",
						TagsToUpsert:     map[string]string{"environment": "staging"},
						AsyncOperationId: "custom-op-id",
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceTagsResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "custom-op-id"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "custom-op-id"},
		},
		{
			name: "GetNamespaceError",
			cmd: temporalcloudcli.CloudNamespaceTagUpdateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				Key:              "environment",
				Value:            "staging",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("namespace not found"))
			},
			expectedErr: "namespace not found",
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudNamespaceTagUpdateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				Key:              "environment",
				Value:            "staging",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting update.",
		},
		{
			name: "ApiError",
			cmd: temporalcloudcli.CloudNamespaceTagUpdateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				Key:              "environment",
				Value:            "staging",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("API error"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			expectedErr:   "update operation failed",
		},
		{
			name: "NothingToChange",
			cmd: temporalcloudcli.CloudNamespaceTagUpdateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				Key:              "environment",
				Value:            "staging",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, status.Error(codes.InvalidArgument, "nothing to change"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			expectedErr:   "update operation failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- DeleteTag ---

func TestDeleteTag(t *testing.T) {
	existingNamespace := &namespacev1.Namespace{
		Namespace: "test-namespace.test-account",
		Tags: map[string]string{
			"environment": "production",
			"team":        "platform",
		},
	}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceTagDeleteCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudNamespaceTagDeleteCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				Key:              "environment",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace.test-account"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, &cloudservice.UpdateNamespaceTagsRequest{
						Namespace:    "test-namespace.test-account",
						TagsToRemove: []string{"environment"},
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceTagsResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-delete"},
					}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{
				ExpectPrompApply: true,
				ExpectPromptApplyExisting: &namespacev1.Namespace{Tags: map[string]string{
					"environment": "production",
					"team":        "platform",
				}},
				ExpectPromptApplyModified: &namespacev1.Namespace{Tags: map[string]string{"team": "platform"}},
				PromptResult:              true,
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-delete"},
		},
		{
			name: "AsyncOperationIdOverride",
			cmd: temporalcloudcli.CloudNamespaceTagDeleteCommand{
				NamespaceOptions:      temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				AsyncOperationOptions: temporalcloudcli.AsyncOperationOptions{AsyncOperationId: "custom-op-id"},
				Key:                   "environment",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, &cloudservice.UpdateNamespaceTagsRequest{
						Namespace:        "test-namespace.test-account",
						TagsToRemove:     []string{"environment"},
						AsyncOperationId: "custom-op-id",
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceTagsResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "custom-op-id"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "custom-op-id"},
		},
		{
			name: "GetNamespaceError",
			cmd: temporalcloudcli.CloudNamespaceTagDeleteCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				Key:              "environment",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("namespace not found"))
			},
			expectedErr: "namespace not found",
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudNamespaceTagDeleteCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				Key:              "environment",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting delete.",
		},
		{
			name: "ApiError",
			cmd: temporalcloudcli.CloudNamespaceTagDeleteCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "test-namespace.test-account"},
				Key:              "environment",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("API error"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			expectedErr:   "delete operation failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}
