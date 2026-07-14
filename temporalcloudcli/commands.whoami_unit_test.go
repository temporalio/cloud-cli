package temporalcloudcli_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
)

func TestWhoami(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudWhoamiCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
	}{
		{
			name: "WhoamiSuccess",
			cmd:  temporalcloudcli.CloudWhoamiCommand{},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCurrentIdentity(mock.Anything, &cloudservice.GetCurrentIdentityRequest{}, mock.Anything).
					Return(&cloudservice.GetCurrentIdentityResponse{}, nil)
			},
		},
		{
			name: "WhoamiNotLoggedIn",
			cmd:  temporalcloudcli.CloudWhoamiCommand{},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCurrentIdentity(mock.Anything, &cloudservice.GetCurrentIdentityRequest{}, mock.Anything).
					Return(nil, fmt.Errorf("no login session found, please run `temporal cloud login`"))
			},
			expectedErr: "no login session found, please run `temporal cloud login`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}
