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

// --- ListNamespaceTags ---

func TestListNamespaceTags(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceTagListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedOutputJson      any
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudNamespaceTagListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.acct"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Tags: map[string]string{"env": "prod", "team": "platform"},
						},
					}, nil)
			},
			// Tags are sorted by key: "env" < "team"
			expectedOutputJson: struct{ Tags []temporalcloudcli.Tag }{
				Tags: []temporalcloudcli.Tag{
					{Key: "env", Value: "prod"},
					{Key: "team", Value: "platform"},
				},
			},
		},
		{
			name: "EmptyTags",
			cmd:  temporalcloudcli.CloudNamespaceTagListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.acct"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{},
					}, nil)
			},
			expectedOutputJson: struct{ Tags []temporalcloudcli.Tag }{
				Tags: []temporalcloudcli.Tag{},
			},
		},
		{
			name: "GetNamespaceError",
			cmd:  temporalcloudcli.CloudNamespaceTagListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.acct"}, mock.Anything).
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
				ExpectedOutputJson:      tt.expectedOutputJson,
			})
		})
	}
}

// --- CreateNamespaceTag ---

func TestCreateNamespaceTag(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceTagCreateCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
		expectedOutputJson      any
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudNamespaceTagCreateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
				Key:              "env",
				Value:            "prod",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, &cloudservice.UpdateNamespaceTagsRequest{
						Namespace:    "my-ns.acct",
						TagsToUpsert: map[string]string{"env": "prod"},
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceTagsResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-1"},
			expectedOutputJson: &cloudservice.UpdateNamespaceTagsResponse{
				AsyncOperation: &operation.AsyncOperation{
					Id:    "op-1",
					State: operation.AsyncOperation_STATE_FULFILLED,
				},
			},
		},
		{
			name: "PromptDeclines",
			cmd: temporalcloudcli.CloudNamespaceTagCreateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
				Key:              "env",
				Value:            "prod",
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting create.",
		},
		{
			name: "PromptError",
			cmd: temporalcloudcli.CloudNamespaceTagCreateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
				Key:              "env",
				Value:            "prod",
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptError: errors.New("prompt failed")},
			expectedErr:   "prompt failed",
		},
		{
			name: "APIError",
			cmd: temporalcloudcli.CloudNamespaceTagCreateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
				Key:              "env",
				Value:            "prod",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("API error"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedErr:   "API error",
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
				ExpectedOutputJson:      tt.expectedOutputJson,
			})
		})
	}
}

// --- UpdateNamespaceTag ---

func TestUpdateNamespaceTag(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceTagUpdateCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
		expectedOutputJson      any
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudNamespaceTagUpdateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
				Key:              "env",
				Value:            "staging",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, &cloudservice.UpdateNamespaceTagsRequest{
						Namespace:    "my-ns.acct",
						TagsToUpsert: map[string]string{"env": "staging"},
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceTagsResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-1"},
			expectedOutputJson: &cloudservice.UpdateNamespaceTagsResponse{
				AsyncOperation: &operation.AsyncOperation{
					Id:    "op-1",
					State: operation.AsyncOperation_STATE_FULFILLED,
				},
			},
		},
		{
			name: "APIError",
			cmd: temporalcloudcli.CloudNamespaceTagUpdateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
				Key:              "env",
				Value:            "staging",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("API error"))
			},
			expectedErr: "API error",
		},
		{
			name: "PollingError",
			cmd: temporalcloudcli.CloudNamespaceTagUpdateCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
				Key:              "env",
				Value:            "staging",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.UpdateNamespaceTagsResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{
				AsyncOperationID: "op-1",
				ErrorToReturn:    errors.New("polling failed"),
			},
			expectedErr: "polling failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedOutputJson,
			})
		})
	}
}

// --- DeleteNamespaceTag ---

func TestDeleteNamespaceTag(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceTagDeleteCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
		expectedOutputJson      any
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudNamespaceTagDeleteCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
				Key:              "env",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, &cloudservice.UpdateNamespaceTagsRequest{
						Namespace:    "my-ns.acct",
						TagsToRemove: []string{"env"},
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceTagsResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-1"},
			expectedOutputJson: &cloudservice.UpdateNamespaceTagsResponse{
				AsyncOperation: &operation.AsyncOperation{
					Id:    "op-1",
					State: operation.AsyncOperation_STATE_FULFILLED,
				},
			},
		},
		{
			name: "PromptDeclines",
			cmd: temporalcloudcli.CloudNamespaceTagDeleteCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
				Key:              "env",
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting delete.",
		},
		{
			name: "PromptError",
			cmd: temporalcloudcli.CloudNamespaceTagDeleteCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
				Key:              "env",
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptError: errors.New("prompt failed")},
			expectedErr:   "prompt failed",
		},
		{
			name: "APIError",
			cmd: temporalcloudcli.CloudNamespaceTagDeleteCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
				Key:              "env",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					UpdateNamespaceTags(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("API error"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedErr:   "API error",
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
				ExpectedOutputJson:      tt.expectedOutputJson,
			})
		})
	}
}
