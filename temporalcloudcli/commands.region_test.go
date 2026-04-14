package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	regionv1 "go.temporal.io/cloud-sdk/api/region/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

// --- GetRegion ---

func TestGetRegion(t *testing.T) {
	testRegion := &regionv1.Region{
		Id:                  "aws-us-east-1",
		CloudProviderRegion: "us-east-1",
		Location:            "US East (N. Virginia)",
	}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudRegionGetCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudRegionGetCommand{Region: "aws-us-east-1"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetRegion(mock.Anything, &cloudservice.GetRegionRequest{Region: "aws-us-east-1"}, mock.Anything).
					Return(&cloudservice.GetRegionResponse{Region: testRegion}, nil)
			},
			expectedJsonOutput: testRegion,
		},
		{
			name: "GetRegionError",
			cmd:  temporalcloudcli.CloudRegionGetCommand{Region: "bad-region"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetRegion(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
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

// --- ListRegions ---

func TestListRegions(t *testing.T) {
	testRegions := []*regionv1.Region{
		{Id: "aws-us-east-1", CloudProviderRegion: "us-east-1", Location: "US East (N. Virginia)"},
		{Id: "aws-us-west-2", CloudProviderRegion: "us-west-2", Location: "US West (Oregon)"},
	}

	tests := []struct {
		name                    string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetRegions(mock.Anything, &cloudservice.GetRegionsRequest{}, mock.Anything).
					Return(&cloudservice.GetRegionsResponse{Regions: testRegions}, nil)
			},
			expectedJsonOutput: struct {
				Regions []*regionv1.Region
			}{Regions: testRegions},
		},
		{
			name: "Empty",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetRegions(mock.Anything, &cloudservice.GetRegionsRequest{}, mock.Anything).
					Return(&cloudservice.GetRegionsResponse{}, nil)
			},
			expectedJsonOutput: struct {
				Regions []*regionv1.Region
			}{},
		},
		{
			name: "GetRegionsError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetRegions(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("api error"))
			},
			expectedErr: "api error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudRegionListCommand{}, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedJsonOutput,
			})
		})
	}
}
