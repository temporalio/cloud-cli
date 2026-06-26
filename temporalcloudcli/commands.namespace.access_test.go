package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

func TestNamespaceUserList(t *testing.T) {
	type listOutput struct {
		Users         []*identityv1.UserNamespaceAssignment
		NextPageToken string
	}
	apiErr := errors.New("api error")

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceUserListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedOutputJson      any
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudNamespaceUserListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace.my-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserNamespaceAssignments(mock.Anything, &cloudservice.GetUserNamespaceAssignmentsRequest{
						Namespace: "my-namespace.my-account",
					}, mock.Anything).
					Return(&cloudservice.GetUserNamespaceAssignmentsResponse{
						Users: []*identityv1.UserNamespaceAssignment{
							{Id: "user-1", Email: "alice@example.com", NamespaceAccess: &identityv1.NamespaceAccess{Permission: identityv1.NamespaceAccess_PERMISSION_WRITE}},
							{Id: "user-2", Email: "bob@example.com", InheritedAccess: true},
						},
					}, nil)
			},
			expectedOutputJson: listOutput{
				Users: []*identityv1.UserNamespaceAssignment{
					{Id: "user-1", Email: "alice@example.com", NamespaceAccess: &identityv1.NamespaceAccess{Permission: identityv1.NamespaceAccess_PERMISSION_WRITE}},
					{Id: "user-2", Email: "bob@example.com", InheritedAccess: true},
				},
			},
		},
		{
			name: "WithPagination",
			cmd: temporalcloudcli.CloudNamespaceUserListCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace.my-account"},
				PageSize:         10,
				PageToken:        "tok-abc",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserNamespaceAssignments(mock.Anything, &cloudservice.GetUserNamespaceAssignmentsRequest{
						Namespace: "my-namespace.my-account",
						PageSize:  10,
						PageToken: "tok-abc",
					}, mock.Anything).
					Return(&cloudservice.GetUserNamespaceAssignmentsResponse{
						Users:         []*identityv1.UserNamespaceAssignment{{Id: "user-3", Email: "carol@example.com"}},
						NextPageToken: "tok-def",
					}, nil)
			},
			expectedOutputJson: listOutput{
				Users:         []*identityv1.UserNamespaceAssignment{{Id: "user-3", Email: "carol@example.com"}},
				NextPageToken: "tok-def",
			},
		},
		{
			name: "ApiError",
			cmd:  temporalcloudcli.CloudNamespaceUserListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace.my-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserNamespaceAssignments(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, apiErr)
			},
			expectedErr: "api error",
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

func TestNamespaceServiceAccountList(t *testing.T) {
	type listOutput struct {
		ServiceAccounts []*identityv1.ServiceAccountNamespaceAssignment
		NextPageToken   string
	}
	apiErr := errors.New("api error")

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceServiceAccountListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedOutputJson      any
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudNamespaceServiceAccountListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace.my-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccountNamespaceAssignments(mock.Anything, &cloudservice.GetServiceAccountNamespaceAssignmentsRequest{
						Namespace: "my-namespace.my-account",
					}, mock.Anything).
					Return(&cloudservice.GetServiceAccountNamespaceAssignmentsResponse{
						ServiceAccounts: []*identityv1.ServiceAccountNamespaceAssignment{
							{Id: "sa-1", Name: "ci-runner", NamespaceAccess: &identityv1.NamespaceAccess{Permission: identityv1.NamespaceAccess_PERMISSION_READ}},
						},
					}, nil)
			},
			expectedOutputJson: listOutput{
				ServiceAccounts: []*identityv1.ServiceAccountNamespaceAssignment{
					{Id: "sa-1", Name: "ci-runner", NamespaceAccess: &identityv1.NamespaceAccess{Permission: identityv1.NamespaceAccess_PERMISSION_READ}},
				},
			},
		},
		{
			name: "WithPagination",
			cmd: temporalcloudcli.CloudNamespaceServiceAccountListCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace.my-account"},
				PageSize:         5,
				PageToken:        "tok-abc",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccountNamespaceAssignments(mock.Anything, &cloudservice.GetServiceAccountNamespaceAssignmentsRequest{
						Namespace: "my-namespace.my-account",
						PageSize:  5,
						PageToken: "tok-abc",
					}, mock.Anything).
					Return(&cloudservice.GetServiceAccountNamespaceAssignmentsResponse{
						ServiceAccounts: []*identityv1.ServiceAccountNamespaceAssignment{{Id: "sa-2", Name: "deploy-bot"}},
						NextPageToken:   "tok-def",
					}, nil)
			},
			expectedOutputJson: listOutput{
				ServiceAccounts: []*identityv1.ServiceAccountNamespaceAssignment{{Id: "sa-2", Name: "deploy-bot"}},
				NextPageToken:   "tok-def",
			},
		},
		{
			name: "ApiError",
			cmd:  temporalcloudcli.CloudNamespaceServiceAccountListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace.my-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccountNamespaceAssignments(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, apiErr)
			},
			expectedErr: "api error",
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

func TestNamespaceUserGroupList(t *testing.T) {
	type listOutput struct {
		Groups        []*identityv1.UserGroupNamespaceAssignment
		NextPageToken string
	}
	apiErr := errors.New("api error")

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceUserGroupListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedOutputJson      any
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudNamespaceUserGroupListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace.my-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroupNamespaceAssignments(mock.Anything, &cloudservice.GetUserGroupNamespaceAssignmentsRequest{
						Namespace: "my-namespace.my-account",
					}, mock.Anything).
					Return(&cloudservice.GetUserGroupNamespaceAssignmentsResponse{
						Groups: []*identityv1.UserGroupNamespaceAssignment{
							{Id: "grp-1", DisplayName: "Platform", NamespaceAccess: &identityv1.NamespaceAccess{Permission: identityv1.NamespaceAccess_PERMISSION_ADMIN}},
						},
					}, nil)
			},
			expectedOutputJson: listOutput{
				Groups: []*identityv1.UserGroupNamespaceAssignment{
					{Id: "grp-1", DisplayName: "Platform", NamespaceAccess: &identityv1.NamespaceAccess{Permission: identityv1.NamespaceAccess_PERMISSION_ADMIN}},
				},
			},
		},
		{
			name: "WithPagination",
			cmd: temporalcloudcli.CloudNamespaceUserGroupListCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace.my-account"},
				PageSize:         5,
				PageToken:        "tok-abc",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroupNamespaceAssignments(mock.Anything, &cloudservice.GetUserGroupNamespaceAssignmentsRequest{
						Namespace: "my-namespace.my-account",
						PageSize:  5,
						PageToken: "tok-abc",
					}, mock.Anything).
					Return(&cloudservice.GetUserGroupNamespaceAssignmentsResponse{
						Groups:        []*identityv1.UserGroupNamespaceAssignment{{Id: "grp-2", DisplayName: "SRE"}},
						NextPageToken: "tok-def",
					}, nil)
			},
			expectedOutputJson: listOutput{
				Groups:        []*identityv1.UserGroupNamespaceAssignment{{Id: "grp-2", DisplayName: "SRE"}},
				NextPageToken: "tok-def",
			},
		},
		{
			name: "ApiError",
			cmd:  temporalcloudcli.CloudNamespaceUserGroupListCommand{NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace.my-account"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroupNamespaceAssignments(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, apiErr)
			},
			expectedErr: "api error",
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
