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

var testEndpoint2 = &nexusv1.Endpoint{
	Id:              "ep-456",
	ResourceVersion: "v1",
	Spec: &nexusv1.EndpointSpec{
		Name: "other-endpoint",
		TargetSpec: &nexusv1.EndpointTargetSpec{
			Variant: &nexusv1.EndpointTargetSpec_WorkerTargetSpec{
				WorkerTargetSpec: &nexusv1.WorkerTargetSpec{
					NamespaceId: "ns-456",
					TaskQueue:   "other-task-queue",
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

// --- ListNexusEndpoints ---

func TestListNexusEndpoints(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNexusEndpointListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "SuccessSingleEndpoint",
			cmd:  temporalcloudcli.CloudNexusEndpointListCommand{},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{
						Endpoints: []*nexusv1.Endpoint{testEndpoint},
					}, nil)
			},
			expectedJsonOutput: struct {
				Endpoints []*nexusv1.Endpoint
			}{
				Endpoints: []*nexusv1.Endpoint{testEndpoint},
			},
		},
		{
			name: "MultiplePages",
			cmd:  temporalcloudcli.CloudNexusEndpointListCommand{},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{
						Endpoints:     []*nexusv1.Endpoint{testEndpoint},
						NextPageToken: "page-2",
					}, nil)
				c.EXPECT().
					GetNexusEndpoints(mock.Anything, &cloudservice.GetNexusEndpointsRequest{
						PageToken: "page-2",
					}, mock.Anything).
					Return(&cloudservice.GetNexusEndpointsResponse{
						Endpoints: []*nexusv1.Endpoint{testEndpoint2},
					}, nil)
			},
			expectedJsonOutput: struct {
				Endpoints []*nexusv1.Endpoint
			}{
				Endpoints: []*nexusv1.Endpoint{testEndpoint, testEndpoint2},
			},
		},
		{
			name: "APIError",
			cmd:  temporalcloudcli.CloudNexusEndpointListCommand{},
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
