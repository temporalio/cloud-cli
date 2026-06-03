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

// ---- GetUserGroup ----

func TestGetUserGroup(t *testing.T) {
	testGroup := &identityv1.UserGroup{
		Id: "group-1",
		Spec: &identityv1.UserGroupSpec{
			DisplayName: "Engineering",
			GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
		},
	}
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserGroupGetCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudUserGroupGetCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, &cloudservice.GetUserGroupRequest{GroupId: "group-1"}, mock.Anything).
					Return(&cloudservice.GetUserGroupResponse{Group: testGroup}, nil)
			},
			expectedJsonOutput: testGroup,
		},
		{
			name: "GetError",
			cmd:  temporalcloudcli.CloudUserGroupGetCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, mock.Anything, mock.Anything).
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

// ---- ListUserGroups ----

func TestListUserGroups(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserGroupListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
	}{
		{
			name: "NoFilter",
			cmd:  temporalcloudcli.CloudUserGroupListCommand{},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroups(mock.Anything, &cloudservice.GetUserGroupsRequest{}, mock.Anything).
					Return(&cloudservice.GetUserGroupsResponse{
						Groups: []*identityv1.UserGroup{{Id: "group-1"}},
					}, nil)
			},
		},
		{
			name: "GoogleGroupFilter",
			cmd:  temporalcloudcli.CloudUserGroupListCommand{GoogleGroupEmailAddress: "eng@example.com"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroups(mock.Anything, &cloudservice.GetUserGroupsRequest{
						GoogleGroup: &cloudservice.GetUserGroupsRequest_GoogleGroupFilter{EmailAddress: "eng@example.com"},
					}, mock.Anything).
					Return(&cloudservice.GetUserGroupsResponse{}, nil)
			},
		},
		{
			name: "ScimGroupFilter",
			cmd:  temporalcloudcli.CloudUserGroupListCommand{ScimGroupIdpId: "idp-123"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroups(mock.Anything, &cloudservice.GetUserGroupsRequest{
						ScimGroup: &cloudservice.GetUserGroupsRequest_SCIMGroupFilter{IdpId: "idp-123"},
					}, mock.Anything).
					Return(&cloudservice.GetUserGroupsResponse{}, nil)
			},
		},
		{
			name: "GetError",
			cmd:  temporalcloudcli.CloudUserGroupListCommand{},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroups(mock.Anything, mock.Anything, mock.Anything).
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

// ---- DeleteUserGroup ----

func TestDeleteUserGroup(t *testing.T) {
	existingGroup := &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1"}
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserGroupDeleteCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudUserGroupDeleteCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, &cloudservice.GetUserGroupRequest{GroupId: "group-1"}, mock.Anything).
					Return(&cloudservice.GetUserGroupResponse{Group: existingGroup}, nil)
				c.EXPECT().
					DeleteUserGroup(mock.Anything, &cloudservice.DeleteUserGroupRequest{
						GroupId:         "group-1",
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.DeleteUserGroupResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name: "ResourceVersionOverride",
			cmd:  temporalcloudcli.CloudUserGroupDeleteCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}, ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, &cloudservice.GetUserGroupRequest{GroupId: "group-1"}, mock.Anything).
					Return(&cloudservice.GetUserGroupResponse{Group: existingGroup}, nil)
				c.EXPECT().
					DeleteUserGroup(mock.Anything, &cloudservice.DeleteUserGroupRequest{
						GroupId:         "group-1",
						ResourceVersion: "rv-override",
					}, mock.Anything).
					Return(&cloudservice.DeleteUserGroupResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name: "PromptDeclined",
			cmd:  temporalcloudcli.CloudUserGroupDeleteCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetUserGroupResponse{Group: existingGroup}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting delete.",
		},
		{
			name: "GetUserGroupError",
			cmd:  temporalcloudcli.CloudUserGroupDeleteCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
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
			})
		})
	}
}

// ---- CreateCloudGroup ----

func TestCreateCloudGroup(t *testing.T) {
	cloudGroupSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
	}
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserGroupCreateCloudGroupCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudUserGroupCreateCloudGroupCommand{DisplayName: "Engineering"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateUserGroup(mock.Anything, &cloudservice.CreateUserGroupRequest{Spec: cloudGroupSpec}, mock.Anything).
					Return(&cloudservice.CreateUserGroupResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-1"},
		},
		{
			name:          "PromptDeclined",
			cmd:           temporalcloudcli.CloudUserGroupCreateCloudGroupCommand{DisplayName: "Engineering"},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting create.",
		},
		{
			name:        "InvalidAccountRole",
			cmd:         temporalcloudcli.CloudUserGroupCreateCloudGroupCommand{DisplayName: "Engineering", AccountRole: "superadmin"},
			expectedErr: "invalid account role",
		},
		{
			name: "CreateError",
			cmd:  temporalcloudcli.CloudUserGroupCreateCloudGroupCommand{DisplayName: "Engineering"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateUserGroup(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("api error"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedErr:   "api error",
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
			})
		})
	}
}

// ---- CreateGoogleGroup ----

func TestCreateGoogleGroup(t *testing.T) {
	googleGroupSpec := &identityv1.UserGroupSpec{
		DisplayName: "Platform",
		GroupType: &identityv1.UserGroupSpec_GoogleGroup{
			GoogleGroup: &identityv1.GoogleGroupSpec{EmailAddress: "platform@example.com"},
		},
	}
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserGroupCreateGoogleGroupCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudUserGroupCreateGoogleGroupCommand{
				DisplayName:      "Platform",
				GoogleGroupEmail: "platform@example.com",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateUserGroup(mock.Anything, &cloudservice.CreateUserGroupRequest{Spec: googleGroupSpec}, mock.Anything).
					Return(&cloudservice.CreateUserGroupResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-2"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-2"},
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudUserGroupCreateGoogleGroupCommand{
				DisplayName:      "Platform",
				GoogleGroupEmail: "platform@example.com",
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting create.",
		},
		{
			name: "CreateError",
			cmd: temporalcloudcli.CloudUserGroupCreateGoogleGroupCommand{
				DisplayName:      "Platform",
				GoogleGroupEmail: "platform@example.com",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateUserGroup(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("api error"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedErr:   "api error",
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
			})
		})
	}
}

// ---- CreateScimGroup ----

func TestCreateScimGroup(t *testing.T) {
	scimGroupSpec := &identityv1.UserGroupSpec{
		DisplayName: "Security",
		GroupType: &identityv1.UserGroupSpec_ScimGroup{
			ScimGroup: &identityv1.SCIMGroupSpec{IdpId: "idp-abc"},
		},
	}
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserGroupCreateScimGroupCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudUserGroupCreateScimGroupCommand{DisplayName: "Security", ScimIdpId: "idp-abc"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateUserGroup(mock.Anything, &cloudservice.CreateUserGroupRequest{Spec: scimGroupSpec}, mock.Anything).
					Return(&cloudservice.CreateUserGroupResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-3"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-3"},
		},
		{
			name:          "PromptDeclined",
			cmd:           temporalcloudcli.CloudUserGroupCreateScimGroupCommand{DisplayName: "Security", ScimIdpId: "idp-abc"},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting create.",
		},
		{
			name: "CreateError",
			cmd:  temporalcloudcli.CloudUserGroupCreateScimGroupCommand{DisplayName: "Security", ScimIdpId: "idp-abc"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateUserGroup(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("api error"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedErr:   "api error",
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
			})
		})
	}
}

// ---- ApplyUserGroup ----
// Apply has distinctly different mock setups for create vs update, so individual functions are used.

func TestApplyUserGroup_Create(t *testing.T) {
	spec := &identityv1.UserGroupSpec{DisplayName: "Engineering"}
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudUserGroupApplyCommand{
		Spec: `{"displayName":"Engineering"}`,
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetUserGroups(mock.Anything, &cloudservice.GetUserGroupsRequest{DisplayName: "Engineering"}, mock.Anything).
				Return(&cloudservice.GetUserGroupsResponse{}, nil)
			c.EXPECT().
				CreateUserGroup(mock.Anything, &cloudservice.CreateUserGroupRequest{Spec: spec}, mock.Anything).
				Return(&cloudservice.CreateUserGroupResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-apply-1"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-apply-1"},
		JSONOutput:         true,
	})
}

func TestApplyUserGroup_Update(t *testing.T) {
	existingSpec := &identityv1.UserGroupSpec{DisplayName: "Engineering"}
	newSpec := &identityv1.UserGroupSpec{DisplayName: "Engineering"}
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudUserGroupApplyCommand{
		Spec: `{"displayName":"Engineering"}`,
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetUserGroups(mock.Anything, &cloudservice.GetUserGroupsRequest{DisplayName: "Engineering"}, mock.Anything).
				Return(&cloudservice.GetUserGroupsResponse{
					Groups: []*identityv1.UserGroup{
						{Id: "group-1", ResourceVersion: "rv-1", Spec: existingSpec},
					},
				}, nil)
			c.EXPECT().
				UpdateUserGroup(mock.Anything, &cloudservice.UpdateUserGroupRequest{
					GroupId:         "group-1",
					Spec:            newSpec,
					ResourceVersion: "rv-1",
				}, mock.Anything).
				Return(&cloudservice.UpdateUserGroupResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-apply-2"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-apply-2"},
		JSONOutput:         true,
	})
}

func TestApplyUserGroup_Declined(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudUserGroupApplyCommand{
		Spec: `{"displayName":"Engineering"}`,
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetUserGroups(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetUserGroupsResponse{}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
		JSONOutput:    true,
		ExpectedError: "Aborting apply.",
	})
}

func TestApplyUserGroup_GetUserGroupsError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudUserGroupApplyCommand{
		Spec: `{"displayName":"Engineering"}`,
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetUserGroups(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("api error"))
		},
		JSONOutput:    true,
		ExpectedError: "api error",
	})
}

// ---- EditUserGroup ----

func TestEditUserGroup(t *testing.T) {
	existingSpec := &identityv1.UserGroupSpec{DisplayName: "Engineering"}
	editedSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
	}
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserGroupEditCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		editorOptions           temporalcloudcli.TestEditorOptions
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudUserGroupEditCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, &cloudservice.GetUserGroupRequest{GroupId: "group-1"}, mock.Anything).
					Return(&cloudservice.GetUserGroupResponse{
						Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1", Spec: existingSpec},
					}, nil)
				c.EXPECT().
					UpdateUserGroup(mock.Anything, &cloudservice.UpdateUserGroupRequest{
						GroupId:         "group-1",
						Spec:            editedSpec,
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserGroupResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-edit"},
					}, nil)
			},
			editorOptions:      temporalcloudcli.TestEditorOptions{Modified: editedSpec},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-edit"},
		},
		{
			name: "GetUserGroupError",
			cmd:  temporalcloudcli.CloudUserGroupEditCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
		},
		{
			name: "EditorError",
			cmd:  temporalcloudcli.CloudUserGroupEditCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetUserGroupResponse{
						Group: &identityv1.UserGroup{Id: "group-1", Spec: existingSpec},
					}, nil)
			},
			editorOptions: temporalcloudcli.TestEditorOptions{EditorError: errors.New("editor closed")},
			expectedErr:   "editor closed",
		},
		{
			name: "PromptDeclined",
			cmd:  temporalcloudcli.CloudUserGroupEditCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetUserGroupResponse{
						Group: &identityv1.UserGroup{Id: "group-1", Spec: existingSpec},
					}, nil)
			},
			editorOptions: temporalcloudcli.TestEditorOptions{Modified: editedSpec},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting edit.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				EditorOptions:           tt.editorOptions,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// ---- UpdateUserGroup ----

func TestUpdateUserGroup(t *testing.T) {
	existingSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
	}
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserGroupUpdateCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "AccountRole",
			cmd:  temporalcloudcli.CloudUserGroupUpdateCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}, AccountRole: "developer"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, &cloudservice.GetUserGroupRequest{GroupId: "group-1"}, mock.Anything).
					Return(&cloudservice.GetUserGroupResponse{
						Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1", Spec: existingSpec},
					}, nil)
				c.EXPECT().
					UpdateUserGroup(mock.Anything, &cloudservice.UpdateUserGroupRequest{
						GroupId: "group-1",
						Spec: &identityv1.UserGroupSpec{
							DisplayName: "Engineering",
							GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
							Access: &identityv1.Access{
								AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
							},
						},
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserGroupResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-upd"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-upd"},
		},
		{
			name: "NamespaceAccess",
			cmd: temporalcloudcli.CloudUserGroupUpdateCommand{
				GroupIdOptions: temporalcloudcli.GroupIdOptions{
					GroupId: "group-1",
				},
				NamespaceAccess: []string{"my-ns.acct=write"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, &cloudservice.GetUserGroupRequest{GroupId: "group-1"}, mock.Anything).
					Return(&cloudservice.GetUserGroupResponse{
						Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1", Spec: existingSpec},
					}, nil)
				c.EXPECT().
					UpdateUserGroup(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.UpdateUserGroupResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-upd"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-upd"},
		},
		{
			name:        "NoFlags",
			cmd:         temporalcloudcli.CloudUserGroupUpdateCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}},
			expectedErr: "must provide at least one of --account-role, --namespace-access, or --custom-role",
		},
		{
			name:        "InvalidRole",
			cmd:         temporalcloudcli.CloudUserGroupUpdateCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}, AccountRole: "superadmin"},
			expectedErr: "invalid account role",
		},
		{
			name: "GetUserGroupError",
			cmd:  temporalcloudcli.CloudUserGroupUpdateCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}, AccountRole: "developer"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
		},
		{
			name: "PromptDeclined",
			cmd:  temporalcloudcli.CloudUserGroupUpdateCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}, AccountRole: "developer"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetUserGroupResponse{
						Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1", Spec: existingSpec},
					}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting update.",
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
			})
		})
	}
}

// ---- SetUserGroupAccountRole ----

func TestSetUserGroupAccountRole(t *testing.T) {
	oldSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		Access:      &identityv1.Access{},
	}
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserGroupSetAccountRoleCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudUserGroupSetAccountRoleCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}, AccountRole: "developer"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, &cloudservice.GetUserGroupRequest{GroupId: "group-1"}, mock.Anything).
					Return(&cloudservice.GetUserGroupResponse{
						Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateUserGroup(mock.Anything, &cloudservice.UpdateUserGroupRequest{
						GroupId: "group-1",
						Spec: &identityv1.UserGroupSpec{
							DisplayName: "Engineering",
							Access: &identityv1.Access{
								AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
							},
						},
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserGroupResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-role"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-role"},
		},
		{
			name:        "InvalidRole",
			cmd:         temporalcloudcli.CloudUserGroupSetAccountRoleCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}, AccountRole: "invalid"},
			expectedErr: "invalid account role",
		},
		{
			name: "GetUserGroupError",
			cmd:  temporalcloudcli.CloudUserGroupSetAccountRoleCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}, AccountRole: "developer"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
		},
		{
			name: "PromptDeclined",
			cmd:  temporalcloudcli.CloudUserGroupSetAccountRoleCommand{GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"}, AccountRole: "developer"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetUserGroupResponse{
						Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting set-account-role.",
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
			})
		})
	}
}

// ---- SetUserGroupNamespacePermissions ----

func TestSetUserGroupNamespacePermissions(t *testing.T) {
	oldSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		Access: &identityv1.Access{
			NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
				"old-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
			},
		},
	}
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserGroupSetNamespacePermissionsCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudUserGroupSetNamespacePermissionsCommand{
				GroupIdOptions: temporalcloudcli.GroupIdOptions{
					GroupId: "group-1",
				},
				NamespaceAccess: []string{"my-ns.acct=write"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, &cloudservice.GetUserGroupRequest{GroupId: "group-1"}, mock.Anything).
					Return(&cloudservice.GetUserGroupResponse{
						Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateUserGroup(mock.Anything, &cloudservice.UpdateUserGroupRequest{
						GroupId: "group-1",
						Spec: &identityv1.UserGroupSpec{
							DisplayName: "Engineering",
							Access: &identityv1.Access{
								NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
									"old-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
									"my-ns.acct":  {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
								},
							},
						},
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserGroupResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-ns"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-ns"},
		},
		{
			name: "InvalidFormat",
			cmd: temporalcloudcli.CloudUserGroupSetNamespacePermissionsCommand{
				GroupIdOptions: temporalcloudcli.GroupIdOptions{
					GroupId: "group-1",
				},
				NamespaceAccess: []string{"no-equals-sign"},
			},
			expectedErr: "invalid namespace-access",
		},
		{
			name: "GetUserGroupError",
			cmd: temporalcloudcli.CloudUserGroupSetNamespacePermissionsCommand{
				GroupIdOptions: temporalcloudcli.GroupIdOptions{
					GroupId: "group-1",
				},
				NamespaceAccess: []string{"my-ns.acct=write"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudUserGroupSetNamespacePermissionsCommand{
				GroupIdOptions: temporalcloudcli.GroupIdOptions{
					GroupId: "group-1",
				},
				NamespaceAccess: []string{"my-ns.acct=write"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUserGroup(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetUserGroupResponse{
						Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting set-namespace-permissions.",
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
			})
		})
	}
}
