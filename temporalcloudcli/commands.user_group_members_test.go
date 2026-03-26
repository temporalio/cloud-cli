package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

// ---- ListUserGroupMembers ----

func TestListUserGroupMembers(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserGroupMembersListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudUserGroupMembersListCommand{
				GroupIdOptions: temporalcloudcli.GroupIdOptions{
					GroupId: "group-1",
				},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroupMembers(mock.Anything, &cloudservice.GetUserGroupMembersRequest{GroupId: "group-1"}, mock.Anything).
					Return(&cloudservice.GetUserGroupMembersResponse{
						Members: []*identityv1.UserGroupMember{{MemberId: &identityv1.UserGroupMemberId{
							MemberType: &identityv1.UserGroupMemberId_UserId{UserId: "user-1"},
						}}},
					}, nil)
			},
		},
		{
			name: "WithPagination",
			cmd: temporalcloudcli.CloudUserGroupMembersListCommand{
				GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"},
				PageSize:       10,
				PageToken:      "tok-abc",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroupMembers(mock.Anything, &cloudservice.GetUserGroupMembersRequest{
						GroupId:   "group-1",
						PageSize:  10,
						PageToken: "tok-abc",
					}, mock.Anything).
					Return(&cloudservice.GetUserGroupMembersResponse{
						Members: []*identityv1.UserGroupMember{{MemberId: &identityv1.UserGroupMemberId{
							MemberType: &identityv1.UserGroupMemberId_UserId{UserId: "user-1"},
						}}},
						NextPageToken: "tok-next",
					}, nil)
			},
		},
		{
			name: "GetError",
			cmd: temporalcloudcli.CloudUserGroupMembersListCommand{
				GroupIdOptions: temporalcloudcli.GroupIdOptions{
					GroupId: "group-1",
				},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroupMembers(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("api error"))
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
			})
		})
	}
}

// ---- AddUserGroupMember ----

func TestAddUserGroupMember(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserGroupMembersAddCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudUserGroupMembersAddCommand{
				GroupIdOptions: temporalcloudcli.GroupIdOptions{
					GroupId: "group-1",
				},
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{User: &identityv1.User{Id: "user-1"}}, nil)
				c.EXPECT().
					AddUserGroupMember(mock.Anything, &cloudservice.AddUserGroupMemberRequest{
						GroupId:  "group-1",
						MemberId: &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: "user-1"}},
					}, mock.Anything).
					Return(&cloudservice.AddUserGroupMemberResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-add"},
					}, nil)
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-add"},
		},
		{
			name: "SuccessByEmail",
			cmd: temporalcloudcli.CloudUserGroupMembersAddCommand{
				GroupIdOptions:            temporalcloudcli.GroupIdOptions{GroupId: "group-1"},
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserEmail: "alice@example.com"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "alice@example.com"}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{Users: []*identityv1.User{{Id: "user-1"}}}, nil)
				c.EXPECT().
					AddUserGroupMember(mock.Anything, &cloudservice.AddUserGroupMemberRequest{
						GroupId:  "group-1",
						MemberId: &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: "user-1"}},
					}, mock.Anything).
					Return(&cloudservice.AddUserGroupMemberResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-add"},
					}, nil)
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-add"},
		},
		{
			name:        "ValidationError_NoIdentification",
			cmd:         temporalcloudcli.CloudUserGroupMembersAddCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}},
			expectedErr: "must provide either --user-id or --user-email",
		},
		{
			name: "ValidationError_BothUserIDAndEmail",
			cmd: temporalcloudcli.CloudUserGroupMembersAddCommand{
				GroupIdOptions:            temporalcloudcli.GroupIdOptions{GroupId: "group-1"},
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1", UserEmail: "alice@example.com"},
			},
			expectedErr: "cannot provide both --user-id and --user-email",
		},
		{
			name: "GetUsersByEmailError",
			cmd: temporalcloudcli.CloudUserGroupMembersAddCommand{
				GroupIdOptions:            temporalcloudcli.GroupIdOptions{GroupId: "group-1"},
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserEmail: "alice@example.com"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUsers(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("api error"))
			},
			expectedErr: "api error",
		},
		{
			name: "UserNotFoundByEmail",
			cmd: temporalcloudcli.CloudUserGroupMembersAddCommand{
				GroupIdOptions:            temporalcloudcli.GroupIdOptions{GroupId: "group-1"},
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserEmail: "nobody@example.com"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUsers(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetUsersResponse{Users: nil}, nil)
			},
			expectedErr: `no user found with email "nobody@example.com"`,
		},
		{
			name: "GetUserError",
			cmd: temporalcloudcli.CloudUserGroupMembersAddCommand{
				GroupIdOptions: temporalcloudcli.GroupIdOptions{
					GroupId: "group-1",
				},
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("user not found"))
			},
			expectedErr: "user not found",
		},
		{
			name: "AddMemberError",
			cmd: temporalcloudcli.CloudUserGroupMembersAddCommand{
				GroupIdOptions: temporalcloudcli.GroupIdOptions{
					GroupId: "group-1",
				},
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetUserResponse{User: &identityv1.User{Id: "user-1"}}, nil)
				c.EXPECT().
					AddUserGroupMember(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("api error"))
			},
			expectedErr: "api error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// ---- RemoveUserGroupMember ----

func TestRemoveUserGroupMember(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserGroupMembersRemoveCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudUserGroupMembersRemoveCommand{
				GroupIdOptions: temporalcloudcli.GroupIdOptions{
					GroupId: "group-1",
				},
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{User: &identityv1.User{Id: "user-1"}}, nil)
				c.EXPECT().
					RemoveUserGroupMember(mock.Anything, &cloudservice.RemoveUserGroupMemberRequest{
						GroupId:  "group-1",
						MemberId: &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: "user-1"}},
					}, mock.Anything).
					Return(&cloudservice.RemoveUserGroupMemberResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-remove"},
					}, nil)
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-remove"},
		},
		{
			name: "SuccessByEmail",
			cmd: temporalcloudcli.CloudUserGroupMembersRemoveCommand{
				GroupIdOptions:            temporalcloudcli.GroupIdOptions{GroupId: "group-1"},
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserEmail: "alice@example.com"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "alice@example.com"}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{Users: []*identityv1.User{{Id: "user-1"}}}, nil)
				c.EXPECT().
					RemoveUserGroupMember(mock.Anything, &cloudservice.RemoveUserGroupMemberRequest{
						GroupId:  "group-1",
						MemberId: &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: "user-1"}},
					}, mock.Anything).
					Return(&cloudservice.RemoveUserGroupMemberResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-remove"},
					}, nil)
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-remove"},
		},
		{
			name:        "ValidationError_NoIdentification",
			cmd:         temporalcloudcli.CloudUserGroupMembersRemoveCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}},
			expectedErr: "must provide either --user-id or --user-email",
		},
		{
			name: "ValidationError_BothUserIDAndEmail",
			cmd: temporalcloudcli.CloudUserGroupMembersRemoveCommand{
				GroupIdOptions:            temporalcloudcli.GroupIdOptions{GroupId: "group-1"},
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1", UserEmail: "alice@example.com"},
			},
			expectedErr: "cannot provide both --user-id and --user-email",
		},
		{
			name: "GetUserError",
			cmd: temporalcloudcli.CloudUserGroupMembersRemoveCommand{
				GroupIdOptions: temporalcloudcli.GroupIdOptions{
					GroupId: "group-1",
				},
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("user not found"))
			},
			expectedErr: "user not found",
		},
		{
			name: "RemoveMemberError",
			cmd: temporalcloudcli.CloudUserGroupMembersRemoveCommand{
				GroupIdOptions: temporalcloudcli.GroupIdOptions{
					GroupId: "group-1",
				},
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetUserResponse{User: &identityv1.User{Id: "user-1"}}, nil)
				c.EXPECT().
					RemoveUserGroupMember(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("api error"))
			},
			expectedErr: "api error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}
