package temporalcloudcli_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	connectivityrulev1 "go.temporal.io/cloud-sdk/api/connectivityrule/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

const (
	testConnectivityNamespace = "my-ns.my-acct"
	testConnectivityRV        = "rv-1"
)

func ruleIdsMatch(want []string) func(*cloudservice.UpdateNamespaceRequest) bool {
	return func(req *cloudservice.UpdateNamespaceRequest) bool {
		return slices.Equal(req.GetSpec().GetConnectivityRuleIds(), want)
	}
}

// --- List ---

func TestNamespaceConnectivityList(t *testing.T) {
	rules := []*connectivityrulev1.ConnectivityRule{
		{
			Id: "rule-a",
			Spec: &connectivityrulev1.ConnectivityRuleSpec{
				ConnectionType: &connectivityrulev1.ConnectivityRuleSpec_PublicRule{
					PublicRule: &connectivityrulev1.PublicConnectivityRule{},
				},
			},
		},
		{
			Id: "rule-b",
			Spec: &connectivityrulev1.ConnectivityRuleSpec{
				ConnectionType: &connectivityrulev1.ConnectivityRuleSpec_PrivateRule{
					PrivateRule: &connectivityrulev1.PrivateConnectivityRule{
						ConnectionId: "vpce-123",
						Region:       "aws-us-east-1",
					},
				},
			},
		},
	}

	tests := []struct {
		name                    string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedRules           []*connectivityrulev1.ConnectivityRule
		expectedErr             string
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: testConnectivityNamespace}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace:         testConnectivityNamespace,
							ResourceVersion:   testConnectivityRV,
							Spec:              &namespacev1.NamespaceSpec{ConnectivityRuleIds: []string{"rule-a", "rule-b"}},
							ConnectivityRules: rules,
						},
					}, nil)
			},
			expectedRules: rules,
		},
		{
			name: "Empty",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace:       testConnectivityNamespace,
							ResourceVersion: testConnectivityRV,
							Spec:            &namespacev1.NamespaceSpec{},
						},
					}, nil)
			},
			expectedRules: nil,
		},
		{
			name: "GetNamespaceError",
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
			cmd := temporalcloudcli.CloudNamespaceConnectivityListCommand{}
			cmd.Namespace = testConnectivityNamespace
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson: struct {
					ConnectivityRules []*connectivityrulev1.ConnectivityRule
				}{
					ConnectivityRules: tt.expectedRules,
				},
			})
		})
	}
}

// --- Create ---

func TestNamespaceConnectivityCreate(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceConnectivityCreateCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "AppendsToExistingList",
			cmd: temporalcloudcli.CloudNamespaceConnectivityCreateCommand{
				ConnectivityRuleId: []string{"rule-b"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				existing := &namespacev1.NamespaceSpec{ConnectivityRuleIds: []string{"rule-a"}}
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: testConnectivityNamespace}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace:       testConnectivityNamespace,
							ResourceVersion: testConnectivityRV,
							Spec:            existing,
						},
					}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(ruleIdsMatch([]string{"rule-a", "rule-b"})), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-create"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-create"},
		},
		{
			name: "AppendsMultiple",
			cmd: temporalcloudcli.CloudNamespaceConnectivityCreateCommand{
				ConnectivityRuleId: []string{"rule-b", "rule-c"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				existing := &namespacev1.NamespaceSpec{ConnectivityRuleIds: []string{"rule-a"}}
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace:       testConnectivityNamespace,
							ResourceVersion: testConnectivityRV,
							Spec:            existing,
						},
					}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(ruleIdsMatch([]string{"rule-a", "rule-b", "rule-c"})), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-create"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-create"},
		},
		{
			name: "AppendsToEmptyList",
			cmd: temporalcloudcli.CloudNamespaceConnectivityCreateCommand{
				ConnectivityRuleId: []string{"rule-a"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: testConnectivityNamespace}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace:       testConnectivityNamespace,
							ResourceVersion: testConnectivityRV,
							Spec:            &namespacev1.NamespaceSpec{},
						},
					}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(ruleIdsMatch([]string{"rule-a"})), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-create"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-create"},
		},
		{
			name: "GetNamespaceError",
			cmd: temporalcloudcli.CloudNamespaceConnectivityCreateCommand{
				ConnectivityRuleId: []string{"rule-a"},
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
			cmd: temporalcloudcli.CloudNamespaceConnectivityCreateCommand{
				ConnectivityRuleId: []string{"rule-a"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace:       testConnectivityNamespace,
							ResourceVersion: testConnectivityRV,
							Spec:            &namespacev1.NamespaceSpec{},
						},
					}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting create.",
		},
		{
			name: "UpdateError",
			cmd: temporalcloudcli.CloudNamespaceConnectivityCreateCommand{
				ConnectivityRuleId: []string{"rule-a"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace:       testConnectivityNamespace,
							ResourceVersion: testConnectivityRV,
							Spec:            &namespacev1.NamespaceSpec{},
						},
					}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("update failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			expectedErr:   "update failed",
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudNamespaceConnectivityCreateCommand{
				ConnectivityRuleId: []string{"rule-a"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace:       testConnectivityNamespace,
							ResourceVersion: testConnectivityRV,
							Spec:            &namespacev1.NamespaceSpec{},
						},
					}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return req.GetResourceVersion() == "rv-override"
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-create"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-create"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd
			cmd.Namespace = testConnectivityNamespace
			if tt.name == "ResourceVersionOverride" {
				cmd.ResourceVersion = "rv-override"
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- Delete ---

func TestNamespaceConnectivityDelete(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceConnectivityDeleteCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "RemovesAttachedRule",
			cmd: temporalcloudcli.CloudNamespaceConnectivityDeleteCommand{
				ConnectivityRuleId: []string{"rule-b"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				existing := &namespacev1.NamespaceSpec{ConnectivityRuleIds: []string{"rule-a", "rule-b"}}
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: testConnectivityNamespace}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace:       testConnectivityNamespace,
							ResourceVersion: testConnectivityRV,
							Spec:            existing,
						},
					}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(ruleIdsMatch([]string{"rule-a"})), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name: "RemovesMultiple",
			cmd: temporalcloudcli.CloudNamespaceConnectivityDeleteCommand{
				ConnectivityRuleId: []string{"rule-a", "rule-b"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				existing := &namespacev1.NamespaceSpec{ConnectivityRuleIds: []string{"rule-a", "rule-b", "rule-c"}}
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace:       testConnectivityNamespace,
							ResourceVersion: testConnectivityRV,
							Spec:            existing,
						},
					}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(ruleIdsMatch([]string{"rule-c"})), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name: "NotAttachedPassesThrough",
			cmd: temporalcloudcli.CloudNamespaceConnectivityDeleteCommand{
				ConnectivityRuleId: []string{"rule-x"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				existing := &namespacev1.NamespaceSpec{ConnectivityRuleIds: []string{"rule-a"}}
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace:       testConnectivityNamespace,
							ResourceVersion: testConnectivityRV,
							Spec:            existing,
						},
					}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(ruleIdsMatch([]string{"rule-a"})), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name: "GetNamespaceError",
			cmd: temporalcloudcli.CloudNamespaceConnectivityDeleteCommand{
				ConnectivityRuleId: []string{"rule-a"},
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
			cmd: temporalcloudcli.CloudNamespaceConnectivityDeleteCommand{
				ConnectivityRuleId: []string{"rule-a"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace:       testConnectivityNamespace,
							ResourceVersion: testConnectivityRV,
							Spec:            &namespacev1.NamespaceSpec{ConnectivityRuleIds: []string{"rule-a"}},
						},
					}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting delete.",
		},
		{
			name: "UpdateError",
			cmd: temporalcloudcli.CloudNamespaceConnectivityDeleteCommand{
				ConnectivityRuleId: []string{"rule-a"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace:       testConnectivityNamespace,
							ResourceVersion: testConnectivityRV,
							Spec:            &namespacev1.NamespaceSpec{ConnectivityRuleIds: []string{"rule-a"}},
						},
					}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("update failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			expectedErr:   "update failed",
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudNamespaceConnectivityDeleteCommand{
				ConnectivityRuleId: []string{"rule-a"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{
						Namespace: &namespacev1.Namespace{
							Namespace:       testConnectivityNamespace,
							ResourceVersion: testConnectivityRV,
							Spec:            &namespacev1.NamespaceSpec{ConnectivityRuleIds: []string{"rule-a"}},
						},
					}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return req.GetResourceVersion() == "rv-override"
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd
			cmd.Namespace = testConnectivityNamespace
			if tt.name == "ResourceVersionOverride" {
				cmd.ResourceVersion = "rv-override"
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}
