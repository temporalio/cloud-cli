package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/protobuf/proto"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

// --- SetNamespacePermissions ---

func TestUserSetNamespacePermissions(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-ns"}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserSetNamespacePermissionsCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectPrompt            bool
		promptResult            bool
		expectedErr             string
	}{
		{
			name: "ByIdSuccess",
			cmd: temporalcloudcli.CloudUserSetNamespacePermissionsCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				NamespaceAccess:           []string{"my-ns.acct=write", "other-ns.acct=read"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				oldSpec := &identityv1.UserSpec{
					Email: "alice@example.com",
					Access: &identityv1.Access{
						NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
							"old-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
						},
					},
				}
				newSpec := &identityv1.UserSpec{
					Email: "alice@example.com",
					Access: &identityv1.Access{
						NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
							"old-ns.acct":   {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
							"my-ns.acct":    {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
							"other-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
						},
					},
				}
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateUser(mock.Anything, &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
			},
			expectPrompt: true,
			promptResult: true,
		},
		{
			name: "ByEmailSuccess",
			cmd: temporalcloudcli.CloudUserSetNamespacePermissionsCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserEmail: "alice@example.com"},
				NamespaceAccess:           []string{"my-ns.acct=admin"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				oldSpec := &identityv1.UserSpec{Email: "alice@example.com"}
				newSpec := &identityv1.UserSpec{
					Email: "alice@example.com",
					Access: &identityv1.Access{
						NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
							"my-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_ADMIN},
						},
					},
				}
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "alice@example.com"}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec}},
					}, nil)
				c.EXPECT().
					UpdateUser(mock.Anything, &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
			},
			expectPrompt: true,
			promptResult: true,
		},
		{
			name: "RemoveAccess",
			cmd: temporalcloudcli.CloudUserSetNamespacePermissionsCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				NamespaceAccess:           []string{"ns1.acct="},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				oldSpec := &identityv1.UserSpec{
					Email: "alice@example.com",
					Access: &identityv1.Access{
						NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
							"ns1.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
							"ns2.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
						},
					},
				}
				newSpec := &identityv1.UserSpec{
					Email: "alice@example.com",
					Access: &identityv1.Access{
						NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
							"ns2.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
						},
					},
				}
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateUser(mock.Anything, &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
			},
			expectPrompt: true,
			promptResult: true,
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudUserSetNamespacePermissionsCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				NamespaceAccess:           []string{"my-ns.acct=write"},
				ResourceVersionOptions:    temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				oldSpec := &identityv1.UserSpec{Email: "alice@example.com"}
				newSpec := &identityv1.UserSpec{
					Email: "alice@example.com",
					Access: &identityv1.Access{
						NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
							"my-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
						},
					},
				}
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateUser(mock.Anything, &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-override",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
			},
			expectPrompt: true,
			promptResult: true,
		},
		{
			name: "InvalidPermission",
			cmd: temporalcloudcli.CloudUserSetNamespacePermissionsCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				NamespaceAccess:           []string{"my-ns.acct=superwrite"},
			},
			expectedErr: "invalid permission",
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudUserSetNamespacePermissionsCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				NamespaceAccess:           []string{"my-ns.acct=write"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: &identityv1.UserSpec{Email: "alice@example.com"}},
					}, nil)
			},
			expectPrompt: true,
			promptResult: false,
			expectedErr:  "Aborting set.",
		},
		{
			name: "NoIdentifier",
			cmd: temporalcloudcli.CloudUserSetNamespacePermissionsCommand{
				NamespaceAccess: []string{"my-ns.acct=write"},
			},
			expectedErr: "must provide either --user-id or --user-email",
		},
		{
			name: "BothIdentifiers",
			cmd: temporalcloudcli.CloudUserSetNamespacePermissionsCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1", UserEmail: "alice@example.com"},
				NamespaceAccess:           []string{"my-ns.acct=write"},
			},
			expectedErr: "cannot provide both --user-id and --user-email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptOpts := temporalcloudcli.TestPromptOptions{}
			asyncOpts := temporalcloudcli.TestAsyncPollerOptions{}
			if tt.expectPrompt {
				promptOpts.ExpectPromptYes = true
				promptOpts.PromptResult = tt.promptResult
			}
			if tt.cloudClientExpectations != nil && tt.expectedErr == "" {
				asyncOpts.AsyncOperationID = op.Id
			}
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           promptOpts,
				AsyncPollerOptions:      asyncOpts,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- SetAccountRole ---

func TestUserSetAccountRole(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-role"}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserSetAccountRoleCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectPrompt            bool
		promptResult            bool
		expectedErr             string
	}{
		{
			name: "ByIdSuccess",
			cmd: temporalcloudcli.CloudUserSetAccountRoleCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				AccountRole:               "admin",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				oldSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
				}
				newSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
				}
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateUser(mock.Anything, &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
			},
			expectPrompt: true,
			promptResult: true,
		},
		{
			name: "ByEmailSuccess",
			cmd: temporalcloudcli.CloudUserSetAccountRoleCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserEmail: "alice@example.com"},
				AccountRole:               "read",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				oldSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
				}
				newSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_READ}},
				}
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "alice@example.com"}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec}},
					}, nil)
				c.EXPECT().
					UpdateUser(mock.Anything, &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
			},
			expectPrompt: true,
			promptResult: true,
		},
		{
			name: "NilAccess",
			cmd: temporalcloudcli.CloudUserSetAccountRoleCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				AccountRole:               "developer",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				oldSpec := &identityv1.UserSpec{Email: "alice@example.com"}
				newSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
				}
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateUser(mock.Anything, &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
			},
			expectPrompt: true,
			promptResult: true,
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudUserSetAccountRoleCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				AccountRole:               "admin",
				ResourceVersionOptions:    temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				oldSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
				}
				newSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
				}
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateUser(mock.Anything, &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-override",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
			},
			expectPrompt: true,
			promptResult: true,
		},
		{
			name: "InvalidRole",
			cmd: temporalcloudcli.CloudUserSetAccountRoleCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				AccountRole:               "superuser",
			},
			expectedErr: "invalid account role",
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudUserSetAccountRoleCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				AccountRole:               "admin",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{
							Id:              "user-1",
							ResourceVersion: "rv-1",
							Spec: &identityv1.UserSpec{
								Email:  "alice@example.com",
								Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
							},
						},
					}, nil)
			},
			expectPrompt: true,
			promptResult: false,
			expectedErr:  "Aborting set.",
		},
		{
			name: "NoIdentifier",
			cmd: temporalcloudcli.CloudUserSetAccountRoleCommand{
				AccountRole: "admin",
			},
			expectedErr: "must provide either --user-id or --user-email",
		},
		{
			name: "BothIdentifiers",
			cmd: temporalcloudcli.CloudUserSetAccountRoleCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1", UserEmail: "alice@example.com"},
				AccountRole:               "admin",
			},
			expectedErr: "cannot provide both --user-id and --user-email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptOpts := temporalcloudcli.TestPromptOptions{}
			asyncOpts := temporalcloudcli.TestAsyncPollerOptions{}
			if tt.expectPrompt {
				promptOpts.ExpectPromptYes = true
				promptOpts.PromptResult = tt.promptResult
			}
			if tt.cloudClientExpectations != nil && tt.expectedErr == "" {
				asyncOpts.AsyncOperationID = op.Id
			}
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           promptOpts,
				AsyncPollerOptions:      asyncOpts,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- DeleteUser ---

func TestUserDelete(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-delete"}
	apiErr := errors.New("delete error")

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserDeleteCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptResult            bool
		expectedErr             string
	}{
		{
			name: "ByIdSuccess",
			cmd: temporalcloudcli.CloudUserDeleteCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1"},
					}, nil)
				c.EXPECT().
					DeleteUser(mock.Anything, &cloudservice.DeleteUserRequest{
						UserId:          "user-1",
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.DeleteUserResponse{AsyncOperation: op}, nil)
			},
			promptResult: true,
		},
		{
			name: "ByEmailSuccess",
			cmd: temporalcloudcli.CloudUserDeleteCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserEmail: "alice@example.com"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "alice@example.com"}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1"}},
					}, nil)
				c.EXPECT().
					DeleteUser(mock.Anything, &cloudservice.DeleteUserRequest{
						UserId:          "user-1",
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.DeleteUserResponse{AsyncOperation: op}, nil)
			},
			promptResult: true,
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudUserDeleteCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				ResourceVersionOptions:    temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1"},
					}, nil)
				c.EXPECT().
					DeleteUser(mock.Anything, &cloudservice.DeleteUserRequest{
						UserId:          "user-1",
						ResourceVersion: "rv-override",
					}, mock.Anything).
					Return(&cloudservice.DeleteUserResponse{AsyncOperation: op}, nil)
			},
			promptResult: true,
		},
		{
			name: "ByEmailNotFound",
			cmd: temporalcloudcli.CloudUserDeleteCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserEmail: "nobody@example.com"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "nobody@example.com"}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{Users: nil}, nil)
			},
			expectedErr: "no user found with email",
		},
		{
			name: "NoIdentifier",
			cmd:  temporalcloudcli.CloudUserDeleteCommand{},
			expectedErr: "must provide either --user-id or --user-email",
		},
		{
			name: "BothIdentifiers",
			cmd: temporalcloudcli.CloudUserDeleteCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1", UserEmail: "alice@example.com"},
			},
			expectedErr: "cannot provide both --user-id and --user-email",
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudUserDeleteCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1"},
					}, nil)
			},
			promptResult: false,
			expectedErr:  "Aborting delete.",
		},
		{
			name: "ApiError",
			cmd: temporalcloudcli.CloudUserDeleteCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1"},
					}, nil)
				c.EXPECT().
					DeleteUser(mock.Anything, &cloudservice.DeleteUserRequest{
						UserId:          "user-1",
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(nil, apiErr)
			},
			promptResult: true,
			expectedErr:  "delete error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptOpts := temporalcloudcli.TestPromptOptions{}
			asyncOpts := temporalcloudcli.TestAsyncPollerOptions{}
			// PromptYes is called after resolveUser succeeds, before DeleteUser.
			// It is NOT called when validation fails or resolveUser fails.
			if tt.cloudClientExpectations != nil && tt.expectedErr != "no user found with email" {
				promptOpts.ExpectPromptYes = true
				promptOpts.PromptResult = tt.promptResult
			}
			if tt.expectedErr == "" {
				asyncOpts.AsyncOperationID = op.Id
			}
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           promptOpts,
				AsyncPollerOptions:      asyncOpts,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- EditUser ---

func TestUserEdit(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-edit"}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserEditCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		editedSpec              proto.Message
		editorError             error
		expectPrompt            bool
		promptResult            bool
		asyncOpID               string
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudUserEditCommand{UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				existingSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
				}
				editedSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
				}
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec},
					}, nil)
				c.EXPECT().
					UpdateUser(mock.Anything, &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            editedSpec,
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
			},
			editedSpec: &identityv1.UserSpec{
				Email:  "alice@example.com",
				Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
			},
			expectPrompt: true,
			promptResult: true,
			asyncOpID:    op.Id,
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudUserEditCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				ResourceVersionOptions:    temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				existingSpec := &identityv1.UserSpec{Email: "alice@example.com"}
				editedSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_READ}},
				}
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec},
					}, nil)
				c.EXPECT().
					UpdateUser(mock.Anything, &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            editedSpec,
						ResourceVersion: "rv-override",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
			},
			editedSpec: &identityv1.UserSpec{
				Email:  "alice@example.com",
				Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_READ}},
			},
			expectPrompt: true,
			promptResult: true,
			asyncOpID:    op.Id,
		},
		{
			name: "ByEmail",
			cmd:  temporalcloudcli.CloudUserEditCommand{UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserEmail: "alice@example.com"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				existingSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
				}
				editedSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
				}
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "alice@example.com"}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec}},
					}, nil)
				c.EXPECT().
					UpdateUser(mock.Anything, &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            editedSpec,
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
			},
			editedSpec: &identityv1.UserSpec{
				Email:  "alice@example.com",
				Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
			},
			expectPrompt: true,
			promptResult: true,
			asyncOpID:    op.Id,
		},
		{
			name: "PromptDeclined",
			cmd:  temporalcloudcli.CloudUserEditCommand{UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: &identityv1.UserSpec{Email: "alice@example.com"}},
					}, nil)
			},
			editedSpec: &identityv1.UserSpec{
				Email:  "alice@example.com",
				Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
			},
			expectPrompt: true,
			promptResult: false,
			expectedErr:  "Aborting edit.",
		},
		{
			name: "GetUserError",
			cmd:  temporalcloudcli.CloudUserEditCommand{UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			// editedSpec nil: editor must NOT be called
			expectedErr: "not found",
		},
		{
			name: "EditorError",
			cmd:  temporalcloudcli.CloudUserEditCommand{UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: &identityv1.UserSpec{Email: "alice@example.com"}},
					}, nil)
			},
			editorError: errors.New("editor failed"),
			expectedErr: "editor failed",
		},
		{
			name: "ByEmailNotFound",
			cmd:  temporalcloudcli.CloudUserEditCommand{UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserEmail: "nobody@example.com"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "nobody@example.com"}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{Users: nil}, nil)
			},
			// editedSpec nil: editor must NOT be called
			expectedErr: "no user found with email",
		},
		{
			name:        "NoIdentifier",
			cmd:         temporalcloudcli.CloudUserEditCommand{},
			expectedErr: "must provide either --user-id or --user-email",
		},
		{
			name: "BothIdentifiers",
			cmd: temporalcloudcli.CloudUserEditCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1", UserEmail: "alice@example.com"},
			},
			expectedErr: "cannot provide both --user-id and --user-email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editorOpts := temporalcloudcli.TestEditorOptions{}
			if tt.editedSpec != nil {
				editorOpts.Modified = tt.editedSpec
			} else if tt.editorError != nil {
				editorOpts.EditorError = tt.editorError
			}
			promptOpts := temporalcloudcli.TestPromptOptions{}
			if tt.expectPrompt {
				promptOpts.ExpectPrompApply = true
				promptOpts.PromptResult = tt.promptResult
			}
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				EditorOptions:           editorOpts,
				PromptOptions:           promptOpts,
				AsyncPollerOptions:      temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: tt.asyncOpID},
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- ApplyUser ---

func TestUserApply(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-apply"}
	createErr := errors.New("create error")

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserApplyCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptResult            bool
		asyncOpID               string
		expectedErr             string
	}{
		{
			name:        "InvalidSpec_EmptyEmail",
			cmd:         temporalcloudcli.CloudUserApplyCommand{Spec: `{}`},
			expectedErr: "spec must include an email address",
		},
		{
			name: "CreateSuccess",
			cmd: temporalcloudcli.CloudUserApplyCommand{
				Spec: `{"email": "alice@example.com"}`,
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				spec := &identityv1.UserSpec{Email: "alice@example.com"}
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "alice@example.com"}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{Users: nil}, nil)
				c.EXPECT().
					CreateUser(mock.Anything, &cloudservice.CreateUserRequest{Spec: spec}, mock.Anything).
					Return(&cloudservice.CreateUserResponse{UserId: "user-new", AsyncOperation: op}, nil)
			},
			promptResult: true,
			asyncOpID:    op.Id,
		},
		{
			name: "UpdateSuccess",
			cmd: temporalcloudcli.CloudUserApplyCommand{
				Spec: `{"email": "alice@example.com", "access": {"account_access": {"role": "ROLE_ADMIN"}}}`,
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				newSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
				}
				existingSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
				}
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "alice@example.com"}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec}},
					}, nil)
				c.EXPECT().
					UpdateUser(mock.Anything, &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
			},
			promptResult: true,
			asyncOpID:    op.Id,
		},
		{
			name: "UpdateResourceVersionOverride",
			cmd: temporalcloudcli.CloudUserApplyCommand{
				Spec:                   `{"email": "alice@example.com"}`,
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				spec := &identityv1.UserSpec{Email: "alice@example.com"}
				existingSpec := &identityv1.UserSpec{Email: "alice@example.com"}
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "alice@example.com"}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec}},
					}, nil)
				c.EXPECT().
					UpdateUser(mock.Anything, &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            spec,
						ResourceVersion: "rv-override",
					}, mock.Anything).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
			},
			promptResult: true,
			asyncOpID:    op.Id,
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudUserApplyCommand{
				Spec: `{"email": "alice@example.com"}`,
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "alice@example.com"}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{Users: nil}, nil)
			},
			promptResult: false,
			expectedErr:  "Aborting apply.",
		},
		{
			name: "GetUsersError",
			cmd: temporalcloudcli.CloudUserApplyCommand{
				Spec: `{"email": "alice@example.com"}`,
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "alice@example.com"}, mock.Anything).
					Return(nil, errors.New("lookup error"))
			},
			expectedErr: "lookup error",
		},
		{
			name: "CreateApiError",
			cmd: temporalcloudcli.CloudUserApplyCommand{
				Spec: `{"email": "alice@example.com"}`,
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				spec := &identityv1.UserSpec{Email: "alice@example.com"}
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "alice@example.com"}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{Users: nil}, nil)
				c.EXPECT().
					CreateUser(mock.Anything, &cloudservice.CreateUserRequest{Spec: spec}, mock.Anything).
					Return(nil, createErr)
			},
			promptResult: true,
			expectedErr:  "create error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptOpts := temporalcloudcli.TestPromptOptions{}
			if tt.cloudClientExpectations != nil {
				// Prompt is called after GetUsers, before Create/Update
				if tt.expectedErr != "lookup error" {
					promptOpts.ExpectPrompApply = true
					promptOpts.PromptResult = tt.promptResult
				}
			}
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           promptOpts,
				AsyncPollerOptions:      temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: tt.asyncOpID},
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- GetUser ---

func TestUserGet(t *testing.T) {
	apiErr := errors.New("api error")
	testUser := &identityv1.User{
		Id:   "user-1",
		Spec: &identityv1.UserSpec{Email: "alice@example.com"},
	}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserGetCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedOutputJson      interface{}
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudUserGetCommand{UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
					Return(&cloudservice.GetUserResponse{User: testUser}, nil)
			},
			expectedOutputJson: testUser,
		},
		{
			name: "ApiError",
			cmd:  temporalcloudcli.CloudUserGetCommand{UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"}},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
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

// --- InviteUser ---

func TestUserInvite(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-invite"}
	apiErr := errors.New("api error")

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserInviteCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectPrompt            bool
		promptResult            bool
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudUserInviteCommand{
				Email:           "alice@example.com",
				AccountRole:     "developer",
				NamespaceAccess: []string{"my-ns.my-account=write"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				expectedSpec := &identityv1.UserSpec{
					Email: "alice@example.com",
					Access: &identityv1.Access{
						AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
						NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
							"my-ns.my-account": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
						},
					},
				}
				c.EXPECT().
					CreateUser(mock.Anything, &cloudservice.CreateUserRequest{Spec: expectedSpec}, mock.Anything).
					Return(&cloudservice.CreateUserResponse{UserId: "user-1", AsyncOperation: op}, nil)
			},
			expectPrompt: true,
			promptResult: true,
		},
		{
			name: "NoAccess",
			cmd: temporalcloudcli.CloudUserInviteCommand{
				Email: "bob@example.com",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateUser(mock.Anything, &cloudservice.CreateUserRequest{
						Spec: &identityv1.UserSpec{
							Email:  "bob@example.com",
							Access: &identityv1.Access{},
						},
					}, mock.Anything).
					Return(&cloudservice.CreateUserResponse{UserId: "user-2", AsyncOperation: op}, nil)
			},
			expectPrompt: true,
			promptResult: true,
		},
		{
			name:         "PromptDeclined",
			cmd:          temporalcloudcli.CloudUserInviteCommand{Email: "alice@example.com"},
			expectPrompt: true,
			promptResult: false,
			expectedErr:  "Aborting invite.",
		},
		{
			name: "ApiError",
			cmd: temporalcloudcli.CloudUserInviteCommand{
				Email: "alice@example.com",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateUser(mock.Anything, &cloudservice.CreateUserRequest{
						Spec: &identityv1.UserSpec{
							Email:  "alice@example.com",
							Access: &identityv1.Access{},
						},
					}, mock.Anything).
					Return(nil, apiErr)
			},
			expectPrompt: true,
			promptResult: true,
			expectedErr:  "api error",
		},
		{
			name: "InvalidAccountRole",
			cmd: temporalcloudcli.CloudUserInviteCommand{
				Email:       "alice@example.com",
				AccountRole: "superadmin",
			},
			// Validation fails before prompt
			expectedErr: "invalid account role",
		},
		{
			name: "InvalidNamespaceAccess",
			cmd: temporalcloudcli.CloudUserInviteCommand{
				Email:           "alice@example.com",
				NamespaceAccess: []string{"my-ns.acct=badperm"},
			},
			// Validation fails before prompt
			expectedErr: "invalid permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asyncOpts := temporalcloudcli.TestAsyncPollerOptions{}
			if tt.expectedErr == "" {
				asyncOpts.AsyncOperationID = op.Id
			}
			promptOpts := temporalcloudcli.TestPromptOptions{}
			if tt.expectPrompt {
				promptOpts.ExpectPromptYes = true
				promptOpts.PromptResult = tt.promptResult
			}
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           promptOpts,
				AsyncPollerOptions:      asyncOpts,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- ListUsers ---

func TestUserList(t *testing.T) {
	type listOutput struct {
		Users         []*identityv1.User
		NextPageToken string
	}
	apiErr := errors.New("api error")

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudUserListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedOutputJson      any
		expectedErr             string
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{
							{Id: "user-1", Spec: &identityv1.UserSpec{Email: "alice@example.com"}},
							{Id: "user-2", Spec: &identityv1.UserSpec{Email: "bob@example.com"}},
						},
					}, nil)
			},
			expectedOutputJson: listOutput{
				Users: []*identityv1.User{
					{Id: "user-1", Spec: &identityv1.UserSpec{Email: "alice@example.com"}},
					{Id: "user-2", Spec: &identityv1.UserSpec{Email: "bob@example.com"}},
				},
			},
		},
		{
			name: "WithFilters",
			cmd: temporalcloudcli.CloudUserListCommand{
				Email:     "alice@example.com",
				Namespace: "my-namespace.my-account",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{
						Email:     "alice@example.com",
						Namespace: "my-namespace.my-account",
					}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{
							{Id: "user-1", Spec: &identityv1.UserSpec{Email: "alice@example.com"}},
						},
					}, nil)
			},
			expectedOutputJson: listOutput{
				Users: []*identityv1.User{
					{Id: "user-1", Spec: &identityv1.UserSpec{Email: "alice@example.com"}},
				},
			},
		},
		{
			name: "WithPagination",
			cmd: temporalcloudcli.CloudUserListCommand{
				PageSize:  10,
				PageToken: "tok-abc",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{
						PageSize:  10,
						PageToken: "tok-abc",
					}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{
						Users:         []*identityv1.User{{Id: "user-3", Spec: &identityv1.UserSpec{Email: "carol@example.com"}}},
						NextPageToken: "tok-def",
					}, nil)
			},
			expectedOutputJson: listOutput{
				Users:         []*identityv1.User{{Id: "user-3", Spec: &identityv1.UserSpec{Email: "carol@example.com"}}},
				NextPageToken: "tok-def",
			},
		},
		{
			name: "ApiError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{}, mock.Anything).
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
