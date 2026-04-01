package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	nexusv1 "go.temporal.io/cloud-sdk/api/nexus/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

var testEndpoint = &nexusv1.Endpoint{
	Id:              "ep-123",
	ResourceVersion: "v1",
	Spec: &nexusv1.EndpointSpec{
		Name: "my-endpoint",
		TargetSpec: &nexusv1.EndpointTargetSpec{
			Variant: &nexusv1.EndpointTargetSpec_WorkerTargetSpec{
				WorkerTargetSpec: &nexusv1.WorkerTargetSpec{
					NamespaceId: "ns-123",
					TaskQueue:   "my-task-queue",
				},
			},
		},
	},
}

// --- GetNexusEndpoint ---

func TestGetNexusEndpoint(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNexusEndpointGetCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "SuccessEndpointFound",
			cmd:  temporalcloudcli.CloudNexusEndpointGetCommand{Name: "my-endpoint"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						Name: "my-endpoint",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{
						Endpoints: []*nexusv1.Endpoint{testEndpoint},
					}, nil)
			},
			expectedJsonOutput: testEndpoint,
		},
		{
			name: "NotFound",
			cmd:  temporalcloudcli.CloudNexusEndpointGetCommand{Name: "missing"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						Name: "missing",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{}, nil)
			},
			expectedErr: `endpoint "missing" not found`,
		},
		{
			name: "APIError",
			cmd:  temporalcloudcli.CloudNexusEndpointGetCommand{Name: "my-endpoint"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("connection refused"))
			},
			expectedErr: "connection refused",
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
