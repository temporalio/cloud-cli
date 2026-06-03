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

var testRoleSpec = &identityv1.CustomRoleSpec{
	Name:        "reader",
	Description: "read-only",
	Permissions: []*identityv1.CustomRoleSpec_Permission{{
		Resources: &identityv1.CustomRoleSpec_Resources{
			ResourceType: "namespace",
			AllowAll:     true,
		},
		Actions: []string{"read"},
	}},
}

const testRoleSpecJSON = `{
    "name": "reader",
    "description": "read-only",
    "permissions": [{
        "resources": {"resource_type": "namespace", "allow_all": true},
        "actions": ["read"]
    }]
}`

var testRole = &identityv1.CustomRole{
	Id:              "role-1",
	ResourceVersion: "rv-1",
	Spec:            testRoleSpec,
}

// ---- GetCustomRole ----

func TestGetCustomRole(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudCustomRoleGetCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudCustomRoleGetCommand{RoleIdOptions: temporalcloudcli.RoleIdOptions{RoleId: "role-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRole(mock.Anything, &cloudservice.GetCustomRoleRequest{RoleId: "role-1"}, mock.Anything).
					Return(&cloudservice.GetCustomRoleResponse{CustomRole: testRole}, nil)
			},
			expectedJsonOutput: testRole,
		},
		{
			name: "GetError",
			cmd:  temporalcloudcli.CloudCustomRoleGetCommand{RoleIdOptions: temporalcloudcli.RoleIdOptions{RoleId: "role-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRole(mock.Anything, mock.Anything, mock.Anything).
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

// ---- ListCustomRoles ----

func TestListCustomRoles(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudCustomRoleListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudCustomRoleListCommand{},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRoles(mock.Anything, &cloudservice.GetCustomRolesRequest{}, mock.Anything).
					Return(&cloudservice.GetCustomRolesResponse{
						CustomRoles: []*identityv1.CustomRole{testRole},
					}, nil)
			},
		},
		{
			name: "WithPagination",
			cmd:  temporalcloudcli.CloudCustomRoleListCommand{PageSize: 10, PageToken: "tok"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRoles(mock.Anything, &cloudservice.GetCustomRolesRequest{
						PageSize:  10,
						PageToken: "tok",
					}, mock.Anything).
					Return(&cloudservice.GetCustomRolesResponse{}, nil)
			},
		},
		{
			name: "Error",
			cmd:  temporalcloudcli.CloudCustomRoleListCommand{},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRoles(mock.Anything, mock.Anything, mock.Anything).
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

// ---- CreateCustomRole ----

func TestCreateCustomRole(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudCustomRoleCreateCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudCustomRoleCreateCommand{
				Spec: testRoleSpecJSON,
				AsyncOperationOptions: temporalcloudcli.AsyncOperationOptions{
					AsyncOperationId: "op-create",
				},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateCustomRole(mock.Anything, &cloudservice.CreateCustomRoleRequest{
						Spec:             testRoleSpec,
						AsyncOperationId: "op-create",
					}, mock.Anything).
					Return(&cloudservice.CreateCustomRoleResponse{
						RoleId:         "role-1",
						AsyncOperation: &operation.AsyncOperation{Id: "op-create"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-create"},
		},
		{
			name:          "PromptDeclined",
			cmd:           temporalcloudcli.CloudCustomRoleCreateCommand{Spec: testRoleSpecJSON},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting create.",
		},
		{
			name:        "InvalidJSON",
			cmd:         temporalcloudcli.CloudCustomRoleCreateCommand{Spec: "not json"},
			expectedErr: "failed to parse JSON spec",
		},
		{
			name: "CreateError",
			cmd:  temporalcloudcli.CloudCustomRoleCreateCommand{Spec: testRoleSpecJSON},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateCustomRole(mock.Anything, mock.Anything, mock.Anything).
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

// ---- UpdateCustomRole ----

func TestUpdateCustomRole(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudCustomRoleUpdateCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudCustomRoleUpdateCommand{
				RoleIdOptions: temporalcloudcli.RoleIdOptions{RoleId: "role-1"},
				Spec:          testRoleSpecJSON,
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRole(mock.Anything, &cloudservice.GetCustomRoleRequest{RoleId: "role-1"}, mock.Anything).
					Return(&cloudservice.GetCustomRoleResponse{CustomRole: testRole}, nil)
				c.EXPECT().
					UpdateCustomRole(mock.Anything, &cloudservice.UpdateCustomRoleRequest{
						RoleId:          "role-1",
						Spec:            testRoleSpec,
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateCustomRoleResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-update"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-update"},
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudCustomRoleUpdateCommand{
				RoleIdOptions:          temporalcloudcli.RoleIdOptions{RoleId: "role-1"},
				Spec:                   testRoleSpecJSON,
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRole(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetCustomRoleResponse{CustomRole: testRole}, nil)
				c.EXPECT().
					UpdateCustomRole(mock.Anything, &cloudservice.UpdateCustomRoleRequest{
						RoleId:          "role-1",
						Spec:            testRoleSpec,
						ResourceVersion: "rv-override",
					}, mock.Anything).
					Return(&cloudservice.UpdateCustomRoleResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-update"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-update"},
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudCustomRoleUpdateCommand{
				RoleIdOptions: temporalcloudcli.RoleIdOptions{RoleId: "role-1"},
				Spec:          testRoleSpecJSON,
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRole(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetCustomRoleResponse{CustomRole: testRole}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting update.",
		},
		{
			name: "InvalidJSON",
			cmd: temporalcloudcli.CloudCustomRoleUpdateCommand{
				RoleIdOptions: temporalcloudcli.RoleIdOptions{RoleId: "role-1"},
				Spec:          "not json",
			},
			expectedErr: "failed to parse JSON spec",
		},
		{
			name: "GetError",
			cmd: temporalcloudcli.CloudCustomRoleUpdateCommand{
				RoleIdOptions: temporalcloudcli.RoleIdOptions{RoleId: "role-1"},
				Spec:          testRoleSpecJSON,
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRole(mock.Anything, mock.Anything, mock.Anything).
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

// ---- DeleteCustomRole ----

func TestDeleteCustomRole(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudCustomRoleDeleteCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudCustomRoleDeleteCommand{RoleIdOptions: temporalcloudcli.RoleIdOptions{RoleId: "role-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRole(mock.Anything, &cloudservice.GetCustomRoleRequest{RoleId: "role-1"}, mock.Anything).
					Return(&cloudservice.GetCustomRoleResponse{CustomRole: testRole}, nil)
				c.EXPECT().
					DeleteCustomRole(mock.Anything, &cloudservice.DeleteCustomRoleRequest{
						RoleId:          "role-1",
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.DeleteCustomRoleResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-delete"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-delete"},
		},
		{
			name: "PromptDeclined",
			cmd:  temporalcloudcli.CloudCustomRoleDeleteCommand{RoleIdOptions: temporalcloudcli.RoleIdOptions{RoleId: "role-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRole(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetCustomRoleResponse{CustomRole: testRole}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting delete.",
		},
		{
			name: "GetError",
			cmd:  temporalcloudcli.CloudCustomRoleDeleteCommand{RoleIdOptions: temporalcloudcli.RoleIdOptions{RoleId: "role-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRole(mock.Anything, mock.Anything, mock.Anything).
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

// ---- ApplyCustomRole ----

func TestApplyCustomRole_Create(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudCustomRoleApplyCommand{Spec: testRoleSpecJSON},
		temporalcloudcli.TestCommandOptions{
			CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRoles(mock.Anything, &cloudservice.GetCustomRolesRequest{}, mock.Anything).
					Return(&cloudservice.GetCustomRolesResponse{}, nil)
				c.EXPECT().
					CreateCustomRole(mock.Anything, &cloudservice.CreateCustomRoleRequest{Spec: testRoleSpec}, mock.Anything).
					Return(&cloudservice.CreateCustomRoleResponse{
						RoleId:         "role-1",
						AsyncOperation: &operation.AsyncOperation{Id: "op-apply-create"},
					}, nil)
			},
			PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-apply-create"},
			JSONOutput:         true,
		})
}

func TestApplyCustomRole_Update(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudCustomRoleApplyCommand{Spec: testRoleSpecJSON},
		temporalcloudcli.TestCommandOptions{
			CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRoles(mock.Anything, &cloudservice.GetCustomRolesRequest{}, mock.Anything).
					Return(&cloudservice.GetCustomRolesResponse{
						CustomRoles: []*identityv1.CustomRole{testRole},
					}, nil)
				c.EXPECT().
					UpdateCustomRole(mock.Anything, &cloudservice.UpdateCustomRoleRequest{
						RoleId:          "role-1",
						Spec:            testRoleSpec,
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateCustomRoleResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-apply-update"},
					}, nil)
			},
			PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-apply-update"},
			JSONOutput:         true,
		})
}

// AIDEV-NOTE: Apply lookup pages until it finds a match by Name.
// This test exercises the multi-page path through findCustomRoleByName.
func TestApplyCustomRole_PagedLookup(t *testing.T) {
	otherRole := &identityv1.CustomRole{
		Id:              "role-other",
		ResourceVersion: "rv-other",
		Spec:            &identityv1.CustomRoleSpec{Name: "writer"},
	}
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudCustomRoleApplyCommand{Spec: testRoleSpecJSON},
		temporalcloudcli.TestCommandOptions{
			CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRoles(mock.Anything, &cloudservice.GetCustomRolesRequest{}, mock.Anything).
					Return(&cloudservice.GetCustomRolesResponse{
						CustomRoles:   []*identityv1.CustomRole{otherRole},
						NextPageToken: "page-2",
					}, nil)
				c.EXPECT().
					GetCustomRoles(mock.Anything, &cloudservice.GetCustomRolesRequest{PageToken: "page-2"}, mock.Anything).
					Return(&cloudservice.GetCustomRolesResponse{
						CustomRoles: []*identityv1.CustomRole{testRole},
					}, nil)
				c.EXPECT().
					UpdateCustomRole(mock.Anything, &cloudservice.UpdateCustomRoleRequest{
						RoleId:          "role-1",
						Spec:            testRoleSpec,
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateCustomRoleResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-apply-paged"},
					}, nil)
			},
			PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-apply-paged"},
			JSONOutput:         true,
		})
}

func TestApplyCustomRole_Declined(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudCustomRoleApplyCommand{Spec: testRoleSpecJSON},
		temporalcloudcli.TestCommandOptions{
			CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRoles(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetCustomRolesResponse{}, nil)
			},
			PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			JSONOutput:    true,
			ExpectedError: "Aborting apply.",
		})
}

func TestApplyCustomRole_InvalidJSON(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudCustomRoleApplyCommand{Spec: "not json"},
		temporalcloudcli.TestCommandOptions{
			JSONOutput:    true,
			ExpectedError: "failed to parse JSON spec",
		})
}

func TestApplyCustomRole_GetCustomRolesError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudCustomRoleApplyCommand{Spec: testRoleSpecJSON},
		temporalcloudcli.TestCommandOptions{
			CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRoles(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("api error"))
			},
			JSONOutput:    true,
			ExpectedError: "api error",
		})
}

func TestApplyCustomRole_DuplicateNameError(t *testing.T) {
	dup := &identityv1.CustomRole{
		Id:              "role-2",
		ResourceVersion: "rv-2",
		Spec:            &identityv1.CustomRoleSpec{Name: "reader"},
	}
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudCustomRoleApplyCommand{Spec: testRoleSpecJSON},
		temporalcloudcli.TestCommandOptions{
			CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRoles(mock.Anything, &cloudservice.GetCustomRolesRequest{}, mock.Anything).
					Return(&cloudservice.GetCustomRolesResponse{
						CustomRoles: []*identityv1.CustomRole{testRole, dup},
					}, nil)
			},
			JSONOutput:    true,
			ExpectedError: `multiple custom roles found with name "reader"`,
		})
}

// ---- EditCustomRole ----

func TestEditCustomRole(t *testing.T) {
	editedSpec := &identityv1.CustomRoleSpec{
		Name:        "reader",
		Description: "edited",
	}
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudCustomRoleEditCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		editorOptions           temporalcloudcli.TestEditorOptions
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudCustomRoleEditCommand{RoleIdOptions: temporalcloudcli.RoleIdOptions{RoleId: "role-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRole(mock.Anything, &cloudservice.GetCustomRoleRequest{RoleId: "role-1"}, mock.Anything).
					Return(&cloudservice.GetCustomRoleResponse{CustomRole: testRole}, nil)
				c.EXPECT().
					UpdateCustomRole(mock.Anything, &cloudservice.UpdateCustomRoleRequest{
						RoleId:          "role-1",
						Spec:            editedSpec,
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateCustomRoleResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-edit"},
					}, nil)
			},
			editorOptions:      temporalcloudcli.TestEditorOptions{Modified: editedSpec},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-edit"},
		},
		{
			name: "GetError",
			cmd:  temporalcloudcli.CloudCustomRoleEditCommand{RoleIdOptions: temporalcloudcli.RoleIdOptions{RoleId: "role-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRole(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
		},
		{
			name: "EditorError",
			cmd:  temporalcloudcli.CloudCustomRoleEditCommand{RoleIdOptions: temporalcloudcli.RoleIdOptions{RoleId: "role-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRole(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetCustomRoleResponse{CustomRole: testRole}, nil)
			},
			editorOptions: temporalcloudcli.TestEditorOptions{EditorError: errors.New("editor closed")},
			expectedErr:   "editor closed",
		},
		{
			name: "PromptDeclined",
			cmd:  temporalcloudcli.CloudCustomRoleEditCommand{RoleIdOptions: temporalcloudcli.RoleIdOptions{RoleId: "role-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetCustomRole(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetCustomRoleResponse{CustomRole: testRole}, nil)
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
