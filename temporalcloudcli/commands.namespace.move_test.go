package temporalcloudcli_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

const (
	testMoveNamespace = "my-ns.my-acct"
	testMoveRV        = "rv-1"
	testMoveSource    = "proj-source"
	testMoveDest      = "proj-dest"
)

func expectGetNamespaceForMove(c *cloudmock.MockCloudServiceClient) {
	c.EXPECT().
		GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: testMoveNamespace}, mock.Anything).
		Return(&cloudservice.GetNamespaceResponse{
			Namespace: &namespacev1.Namespace{
				Namespace:       testMoveNamespace,
				ResourceVersion: testMoveRV,
				ProjectId:       testMoveSource,
				Spec:            &namespacev1.NamespaceSpec{},
			},
		}, nil)
}

func expectMove(
	c *cloudmock.MockCloudServiceClient,
	matches func(*cloudservice.MoveNamespaceToProjectRequest) bool,
) {
	c.EXPECT().
		MoveNamespaceToProject(mock.Anything, mock.MatchedBy(matches), mock.Anything).
		Return(&cloudservice.MoveNamespaceToProjectResponse{
			AsyncOperation: &operation.AsyncOperation{Id: "op-move"},
		}, nil)
}

func TestNamespaceMoveToProjectConnectivitySelection(t *testing.T) {
	tests := []struct {
		name        string
		cmd         temporalcloudcli.CloudNamespaceMoveToProjectCommand
		wantRequest func(*cloudservice.MoveNamespaceToProjectRequest) bool
	}{
		{
			name: "NamedRuleIds",
			cmd: temporalcloudcli.CloudNamespaceMoveToProjectCommand{
				DestinationProjectId: testMoveDest,
				SourceProjectId:      testMoveSource,
				ConnectivityRuleId:   []string{"rule-a", "rule-b"},
			},
			wantRequest: func(req *cloudservice.MoveNamespaceToProjectRequest) bool {
				return slices.Equal(req.GetRuleIds().GetConnectivityRuleIds(), []string{"rule-a", "rule-b"})
			},
		},
		{
			name: "ExplicitlyUnrestricted",
			cmd: temporalcloudcli.CloudNamespaceMoveToProjectCommand{
				DestinationProjectId: testMoveDest,
				SourceProjectId:      testMoveSource,
				NoConnectivityRules:  true,
			},
			wantRequest: func(req *cloudservice.MoveNamespaceToProjectRequest) bool {
				return req.GetUnrestricted() != nil && req.GetRuleIds() == nil
			},
		},
		{
			name: "SelectionOmitted",
			cmd: temporalcloudcli.CloudNamespaceMoveToProjectCommand{
				DestinationProjectId: testMoveDest,
				SourceProjectId:      testMoveSource,
			},
			wantRequest: func(req *cloudservice.MoveNamespaceToProjectRequest) bool {
				return req.GetDestinationConnectivityRules() == nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd
			cmd.Namespace = testMoveNamespace
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					expectGetNamespaceForMove(c)
					expectMove(c, tt.wantRequest)
				},
				PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
				AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-move"},
			})
		})
	}
}

func TestNamespaceMoveToProjectRequestFields(t *testing.T) {
	tests := []struct {
		name            string
		resourceVersion string
		wantRV          string
	}{
		{name: "ResourceVersionFetched", wantRV: testMoveRV},
		{name: "ResourceVersionOverride", resourceVersion: "rv-override", wantRV: "rv-override"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := temporalcloudcli.CloudNamespaceMoveToProjectCommand{
				DestinationProjectId: testMoveDest,
				SourceProjectId:      testMoveSource,
			}
			cmd.Namespace = testMoveNamespace
			cmd.ResourceVersion = tt.resourceVersion
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					expectGetNamespaceForMove(c)
					expectMove(c, func(req *cloudservice.MoveNamespaceToProjectRequest) bool {
						return req.GetNamespace() == testMoveNamespace &&
							req.GetDestinationProjectId() == testMoveDest &&
							req.GetExpectedSourceProjectId() == testMoveSource &&
							req.GetResourceVersion() == tt.wantRV
					})
				},
				PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
				AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-move"},
			})
		})
	}
}

func TestNamespaceMoveToProjectRejectsInvalidRuleSelection(t *testing.T) {
	tests := []struct {
		name        string
		cmd         temporalcloudcli.CloudNamespaceMoveToProjectCommand
		expectedErr string
	}{
		{
			name: "RuleIdsAndNoRulesTogether",
			cmd: temporalcloudcli.CloudNamespaceMoveToProjectCommand{
				DestinationProjectId: testMoveDest,
				SourceProjectId:      testMoveSource,
				ConnectivityRuleId:   []string{"rule-a"},
				NoConnectivityRules:  true,
			},
			expectedErr: "mutually exclusive",
		},
		{
			name: "DuplicateRuleID",
			cmd: temporalcloudcli.CloudNamespaceMoveToProjectCommand{
				DestinationProjectId: testMoveDest,
				SourceProjectId:      testMoveSource,
				ConnectivityRuleId:   []string{"rule-a", "rule-a"},
			},
			expectedErr: `connectivity rule ID "rule-a" specified more than once`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd
			cmd.Namespace = testMoveNamespace
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				ExpectedError: tt.expectedErr,
			})
		})
	}
}

func TestNamespaceMoveToProjectPromptDeclined(t *testing.T) {
	cmd := temporalcloudcli.CloudNamespaceMoveToProjectCommand{
		DestinationProjectId: testMoveDest,
		SourceProjectId:      testMoveSource,
	}
	cmd.Namespace = testMoveNamespace
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: expectGetNamespaceForMove,
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes:        true,
			ExpectPromptYesMessage: `Move namespace "my-ns.my-acct" from project "proj-source" to project "proj-dest"`,
			PromptResult:           false,
		},
		ExpectedError: "Aborting move.",
	})
}

func TestNamespaceMoveToProjectSurfacesRejection(t *testing.T) {
	cmd := temporalcloudcli.CloudNamespaceMoveToProjectCommand{
		DestinationProjectId: testMoveDest,
		SourceProjectId:      testMoveSource,
	}
	cmd.Namespace = testMoveNamespace
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			expectGetNamespaceForMove(c)
			c.EXPECT().
				MoveNamespaceToProject(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New(`namespace "my-ns.my-acct" has a migration in progress`))
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
		ExpectedError: "has a migration in progress",
	})
}
