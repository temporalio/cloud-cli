package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	nexusv1 "go.temporal.io/cloud-sdk/api/nexus/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

var testMoveEndpoint = &nexusv1.Endpoint{
	Id:              "ep-123",
	ResourceVersion: "v1",
	ProjectId:       testMoveSource,
	Spec:            &nexusv1.EndpointSpec{Name: "my-endpoint"},
}

func expectEndpointByName(c *cloudmock.MockCloudServiceClient) {
	c.EXPECT().
		GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{Name: "my-endpoint"}, mock.Anything).
		Return(&cloudservice.GetNexusEndpointsResponse{Endpoints: []*nexusv1.Endpoint{testMoveEndpoint}}, nil)
}

func expectEndpointByID(c *cloudmock.MockCloudServiceClient) {
	c.EXPECT().
		GetNexusEndpoint(mock.Anything, &cloudservice.GetNexusEndpointRequest{EndpointId: "ep-123"}, mock.Anything).
		Return(&cloudservice.GetNexusEndpointResponse{Endpoint: testMoveEndpoint}, nil)
}

func expectBothProjectsFound(c *cloudmock.MockCloudServiceClient) {
	expectProjectLookup(c, testMoveProject(testMoveSource, "source"), testMoveProject(testMoveDest, "destination"))
}

func expectEndpointMove(c *cloudmock.MockCloudServiceClient, want *cloudservice.MoveNexusEndpointToProjectRequest) {
	c.EXPECT().
		MoveNexusEndpointToProject(mock.Anything, want, mock.Anything).
		Return(&cloudservice.MoveNexusEndpointToProjectResponse{
			AsyncOperation: &operation.AsyncOperation{Id: "op-move"},
		}, nil)
}

func TestNexusEndpointMoveToProject(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNexusEndpointMoveToProjectCommand
		resourceVersion         string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
	}{
		{
			name: "ByNameWithFetchedResourceVersion",
			cmd:  temporalcloudcli.CloudNexusEndpointMoveToProjectCommand{Name: "my-endpoint"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				expectEndpointByName(c)
				expectBothProjectsFound(c)
				expectEndpointMove(c, &cloudservice.MoveNexusEndpointToProjectRequest{
					EndpointId:              "ep-123",
					DestinationProjectId:    testMoveDest,
					ExpectedSourceProjectId: testMoveSource,
					ResourceVersion:         "v1",
				})
			},
		},
		{
			name:            "ByIDWithResourceVersionOverride",
			cmd:             temporalcloudcli.CloudNexusEndpointMoveToProjectCommand{Id: "ep-123"},
			resourceVersion: "v-override",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				expectEndpointByID(c)
				expectBothProjectsFound(c)
				expectEndpointMove(c, &cloudservice.MoveNexusEndpointToProjectRequest{
					EndpointId:              "ep-123",
					DestinationProjectId:    testMoveDest,
					ExpectedSourceProjectId: testMoveSource,
					ResourceVersion:         "v-override",
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmd
			cmd.DestinationProjectId = testMoveDest
			cmd.SourceProjectId = testMoveSource
			cmd.ResourceVersion = tt.resourceVersion
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions: temporalcloudcli.TestPromptOptions{
					ExpectPromptYes:        true,
					ExpectPromptYesMessage: `Move Nexus Endpoint "my-endpoint" from project "source" (proj-source) to project "destination" (proj-dest)`,
					PromptResult:           true,
				},
				AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-move"},
			})
		})
	}
}

func TestNexusEndpointMoveToProjectUnknownDestination(t *testing.T) {
	cmd := temporalcloudcli.CloudNexusEndpointMoveToProjectCommand{
		Name:                 "my-endpoint",
		DestinationProjectId: testMoveDest,
		SourceProjectId:      testMoveSource,
	}
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			expectEndpointByName(c)
			expectProjectLookup(c, testMoveProject(testMoveSource, "source"))
		},
		ExpectedError: `destination project "proj-dest" not found`,
	})
}

func TestNexusEndpointMoveToProjectPromptDeclined(t *testing.T) {
	cmd := temporalcloudcli.CloudNexusEndpointMoveToProjectCommand{
		Name:                 "my-endpoint",
		DestinationProjectId: testMoveDest,
		SourceProjectId:      testMoveSource,
	}
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			expectEndpointByName(c)
			expectBothProjectsFound(c)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
		ExpectedError: "Aborting move.",
	})
}

func TestNexusEndpointMoveToProjectSurfacesRejection(t *testing.T) {
	cmd := temporalcloudcli.CloudNexusEndpointMoveToProjectCommand{
		Name:                 "my-endpoint",
		DestinationProjectId: testMoveDest,
		SourceProjectId:      testMoveSource,
	}
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			expectEndpointByName(c)
			expectBothProjectsFound(c)
			c.EXPECT().
				MoveNexusEndpointToProject(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("resource limit reached: 5"))
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
		ExpectedError: "resource limit reached: 5",
	})
}
