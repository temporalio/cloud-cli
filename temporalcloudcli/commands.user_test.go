package temporalcloudcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/protobuf/proto"
)

func TestSetNamespacePermissions(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-ns"}
	promptErr := errors.New("Aborting apply.")

	tests := []struct {
		name            string
		params          temporalcloudcli.SetNamespacePermissionsParams
		setup           func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler)
		wantErr         error
		wantErrContains string
	}{
		{
			name: "by_id_success",
			params: temporalcloudcli.SetNamespacePermissionsParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				NamespaceAccesses:  []string{"my-ns.acct=write", "other-ns.acct=read"},
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
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
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				prompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)
				cloud.EXPECT().
					UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-1",
					}).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
		},
		{
			name: "by_email_success",
			params: temporalcloudcli.SetNamespacePermissionsParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserEmail: "alice@example.com"},
				NamespaceAccesses:  []string{"my-ns.acct=admin"},
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				oldSpec := &identityv1.UserSpec{Email: "alice@example.com"}
				newSpec := &identityv1.UserSpec{
					Email: "alice@example.com",
					Access: &identityv1.Access{
						NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
							"my-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_ADMIN},
						},
					},
				}
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec}},
					}, nil)
				prompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)
				cloud.EXPECT().
					UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-1",
					}).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
		},
		{
			name: "remove_access",
			params: temporalcloudcli.SetNamespacePermissionsParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				NamespaceAccesses:  []string{"ns1.acct="},
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
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
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				prompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)
				cloud.EXPECT().
					UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-1",
					}).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
		},
		{
			name: "invalid_permission",
			params: temporalcloudcli.SetNamespacePermissionsParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				NamespaceAccesses:  []string{"my-ns.acct=superwrite"},
			},
			wantErrContains: "invalid permission",
		},
		{
			name: "prompt_declined",
			params: temporalcloudcli.SetNamespacePermissionsParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				NamespaceAccesses:  []string{"my-ns.acct=write"},
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				oldSpec := &identityv1.UserSpec{Email: "alice@example.com"}
				newSpec := &identityv1.UserSpec{
					Email: "alice@example.com",
					Access: &identityv1.Access{
						NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
							"my-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
						},
					},
				}
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				prompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(promptErr)
			},
			wantErr: promptErr,
		},
		{
			name: "no_identifier",
			params: temporalcloudcli.SetNamespacePermissionsParams{
				NamespaceAccesses: []string{"my-ns.acct=write"},
			},
			wantErrContains: "must provide either --user-id or --user-email",
		},
		{
			name: "both_identifiers",
			params: temporalcloudcli.SetNamespacePermissionsParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1", UserEmail: "alice@example.com"},
				NamespaceAccesses:  []string{"my-ns.acct=write"},
			},
			wantErrContains: "cannot provide both --user-id and --user-email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCloud := cloudmock.NewMockCloudServiceClient(t)
			mockPrompter := cmdmock.NewMockPrompter(t)
			mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
			if tt.setup != nil {
				tt.setup(mockCloud, mockPrompter, mockHandler)
			}
			params := tt.params
			params.Cloud = mockCloud
			params.Prompter = mockPrompter
			params.OperationHandler = mockHandler

			err := temporalcloudcli.SetNamespacePermissions(context.Background(), params)
			switch {
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			case tt.wantErrContains != "":
				require.ErrorContains(t, err, tt.wantErrContains)
			default:
				require.NoError(t, err)
			}
		})
	}
}

func TestSetAccountRole(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-role"}
	promptErr := errors.New("Aborting apply.")

	tests := []struct {
		name            string
		params          temporalcloudcli.SetAccountRoleParams
		setup           func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler)
		wantErr         error
		wantErrContains string
	}{
		{
			name: "by_id_success",
			params: temporalcloudcli.SetAccountRoleParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				AccountRole:        "admin",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				oldSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
				}
				newSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
				}
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				prompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)
				cloud.EXPECT().
					UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-1",
					}).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
		},
		{
			name: "by_email_success",
			params: temporalcloudcli.SetAccountRoleParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserEmail: "alice@example.com"},
				AccountRole:        "read",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				oldSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
				}
				newSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_READ}},
				}
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec}},
					}, nil)
				prompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)
				cloud.EXPECT().
					UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-1",
					}).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
		},
		{
			name: "nil_access",
			params: temporalcloudcli.SetAccountRoleParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				AccountRole:        "developer",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				oldSpec := &identityv1.UserSpec{Email: "alice@example.com"}
				newSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
				}
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				prompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)
				cloud.EXPECT().
					UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-1",
					}).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
		},
		{
			name: "invalid_role",
			params: temporalcloudcli.SetAccountRoleParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				AccountRole:        "superuser",
			},
			wantErrContains: "invalid account role",
		},
		{
			name: "prompt_declined",
			params: temporalcloudcli.SetAccountRoleParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				AccountRole:        "admin",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				oldSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
				}
				newSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
				}
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				prompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(promptErr)
			},
			wantErr: promptErr,
		},
		{
			name: "no_identifier",
			params: temporalcloudcli.SetAccountRoleParams{
				AccountRole: "admin",
			},
			wantErrContains: "must provide either --user-id or --user-email",
		},
		{
			name: "both_identifiers",
			params: temporalcloudcli.SetAccountRoleParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1", UserEmail: "alice@example.com"},
				AccountRole:        "admin",
			},
			wantErrContains: "cannot provide both --user-id and --user-email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCloud := cloudmock.NewMockCloudServiceClient(t)
			mockPrompter := cmdmock.NewMockPrompter(t)
			mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
			if tt.setup != nil {
				tt.setup(mockCloud, mockPrompter, mockHandler)
			}
			params := tt.params
			params.Cloud = mockCloud
			params.Prompter = mockPrompter
			params.OperationHandler = mockHandler

			err := temporalcloudcli.SetAccountRole(context.Background(), params)
			switch {
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			case tt.wantErrContains != "":
				require.ErrorContains(t, err, tt.wantErrContains)
			default:
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-delete"}
	apiErr := errors.New("delete error")

	tests := []struct {
		name            string
		params          temporalcloudcli.DeleteUserParams
		setup           func(cloud *cloudmock.MockCloudServiceClient, handler *cmdmock.MockAsyncOperationHandler)
		wantErr         error
		wantErrContains string
	}{
		{
			name:   "by_id_success",
			params: temporalcloudcli.DeleteUserParams{UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"}},
			setup: func(cloud *cloudmock.MockCloudServiceClient, handler *cmdmock.MockAsyncOperationHandler) {
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1"},
					}, nil)
				cloud.EXPECT().
					DeleteUser(context.Background(), &cloudservice.DeleteUserRequest{
						UserId:          "user-1",
						ResourceVersion: "rv-1",
					}).
					Return(&cloudservice.DeleteUserResponse{AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
		},
		{
			name:   "by_email_success",
			params: temporalcloudcli.DeleteUserParams{UserIdentification: temporalcloudcli.UserIdentificationOptions{UserEmail: "alice@example.com"}},
			setup: func(cloud *cloudmock.MockCloudServiceClient, handler *cmdmock.MockAsyncOperationHandler) {
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1"}},
					}, nil)
				cloud.EXPECT().
					DeleteUser(context.Background(), &cloudservice.DeleteUserRequest{
						UserId:          "user-1",
						ResourceVersion: "rv-1",
					}).
					Return(&cloudservice.DeleteUserResponse{AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
		},
		{
			name: "resource_version_override",
			params: temporalcloudcli.DeleteUserParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				ResourceVersion:    "rv-override",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, handler *cmdmock.MockAsyncOperationHandler) {
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1"},
					}, nil)
				cloud.EXPECT().
					DeleteUser(context.Background(), &cloudservice.DeleteUserRequest{
						UserId:          "user-1",
						ResourceVersion: "rv-override",
					}).
					Return(&cloudservice.DeleteUserResponse{AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
		},
		{
			name:   "by_email_not_found",
			params: temporalcloudcli.DeleteUserParams{UserIdentification: temporalcloudcli.UserIdentificationOptions{UserEmail: "nobody@example.com"}},
			setup: func(cloud *cloudmock.MockCloudServiceClient, handler *cmdmock.MockAsyncOperationHandler) {
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "nobody@example.com"}).
					Return(&cloudservice.GetUsersResponse{Users: nil}, nil)
			},
			wantErrContains: "no user found with email",
		},
		{
			name:            "no_identifier",
			params:          temporalcloudcli.DeleteUserParams{},
			wantErrContains: "must provide either --user-id or --user-email",
		},
		{
			name: "both_identifiers",
			params: temporalcloudcli.DeleteUserParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1", UserEmail: "alice@example.com"},
			},
			wantErrContains: "cannot provide both --user-id and --user-email",
		},
		{
			name:   "api_error",
			params: temporalcloudcli.DeleteUserParams{UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"}},
			setup: func(cloud *cloudmock.MockCloudServiceClient, handler *cmdmock.MockAsyncOperationHandler) {
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1"},
					}, nil)
				cloud.EXPECT().
					DeleteUser(context.Background(), &cloudservice.DeleteUserRequest{
						UserId:          "user-1",
						ResourceVersion: "rv-1",
					}).
					Return(nil, apiErr)
				handler.EXPECT().HandleErr(apiErr).Return(apiErr)
			},
			wantErr: apiErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCloud := cloudmock.NewMockCloudServiceClient(t)
			mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
			if tt.setup != nil {
				tt.setup(mockCloud, mockHandler)
			}
			params := tt.params
			params.Cloud = mockCloud
			params.OperationHandler = mockHandler

			err := temporalcloudcli.DeleteUser(context.Background(), params)
			switch {
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			case tt.wantErrContains != "":
				require.ErrorContains(t, err, tt.wantErrContains)
			default:
				require.NoError(t, err)
			}
		})
	}
}

func TestEditUser(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-edit"}
	promptErr := errors.New("Aborting apply.")
	editorErr := errors.New("editor failed")
	getUserErr := errors.New("not found")

	tests := []struct {
		name            string
		params          temporalcloudcli.EditUserParams
		makeRunEditor   func(t *testing.T) func(proto.Message, proto.Message) error
		setup           func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler)
		wantErr         error
		wantErrContains string
	}{
		{
			name:   "success",
			params: temporalcloudcli.EditUserParams{UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"}},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				existingSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
				}
				editedSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
				}
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec},
					}, nil)
				prompter.EXPECT().PromptApply(existingSpec, editedSpec, false).Return(nil)
				cloud.EXPECT().
					UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            editedSpec,
						ResourceVersion: "rv-1",
					}).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
			makeRunEditor: func(t *testing.T) func(proto.Message, proto.Message) error {
				editedSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
				}
				return func(existing, target proto.Message) error {
					proto.Merge(target, editedSpec)
					return nil
				}
			},
		},
		{
			name: "resource_version_override",
			params: temporalcloudcli.EditUserParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
				ResourceVersion:    "rv-override",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				existingSpec := &identityv1.UserSpec{Email: "alice@example.com"}
				editedSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_READ}},
				}
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec},
					}, nil)
				prompter.EXPECT().PromptApply(existingSpec, editedSpec, false).Return(nil)
				cloud.EXPECT().
					UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            editedSpec,
						ResourceVersion: "rv-override",
					}).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
			makeRunEditor: func(t *testing.T) func(proto.Message, proto.Message) error {
				editedSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_READ}},
				}
				return func(existing, target proto.Message) error {
					proto.Merge(target, editedSpec)
					return nil
				}
			},
		},
		{
			name:   "prompt_declined",
			params: temporalcloudcli.EditUserParams{UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"}},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				existingSpec := &identityv1.UserSpec{Email: "alice@example.com"}
				editedSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
				}
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec},
					}, nil)
				prompter.EXPECT().PromptApply(existingSpec, editedSpec, false).Return(promptErr)
			},
			makeRunEditor: func(t *testing.T) func(proto.Message, proto.Message) error {
				editedSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
				}
				return func(existing, target proto.Message) error {
					proto.Merge(target, editedSpec)
					return nil
				}
			},
			wantErr: promptErr,
		},
		{
			name:   "get_user_error",
			params: temporalcloudcli.EditUserParams{UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"}},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(nil, getUserErr)
			},
			makeRunEditor: func(t *testing.T) func(proto.Message, proto.Message) error {
				called := false
				t.Cleanup(func() { assert.False(t, called, "RunEditor should not be called") })
				return func(existing, target proto.Message) error {
					called = true
					return nil
				}
			},
			wantErr: getUserErr,
		},
		{
			name:   "editor_error",
			params: temporalcloudcli.EditUserParams{UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"}},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: &identityv1.UserSpec{Email: "alice@example.com"}},
					}, nil)
			},
			makeRunEditor: func(t *testing.T) func(proto.Message, proto.Message) error {
				return func(existing, target proto.Message) error { return editorErr }
			},
			wantErr: editorErr,
		},
		{
			name:   "by_email_success",
			params: temporalcloudcli.EditUserParams{UserIdentification: temporalcloudcli.UserIdentificationOptions{UserEmail: "alice@example.com"}},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				existingSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
				}
				editedSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
				}
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec}},
					}, nil)
				prompter.EXPECT().PromptApply(existingSpec, editedSpec, false).Return(nil)
				cloud.EXPECT().
					UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            editedSpec,
						ResourceVersion: "rv-1",
					}).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
			makeRunEditor: func(t *testing.T) func(proto.Message, proto.Message) error {
				editedSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
				}
				return func(existing, target proto.Message) error {
					proto.Merge(target, editedSpec)
					return nil
				}
			},
		},
		{
			name:   "by_email_not_found",
			params: temporalcloudcli.EditUserParams{UserIdentification: temporalcloudcli.UserIdentificationOptions{UserEmail: "nobody@example.com"}},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "nobody@example.com"}).
					Return(&cloudservice.GetUsersResponse{Users: nil}, nil)
			},
			makeRunEditor: func(t *testing.T) func(proto.Message, proto.Message) error {
				return func(existing, target proto.Message) error { return nil }
			},
			wantErrContains: "no user found with email",
		},
		{
			name:   "no_identifier",
			params: temporalcloudcli.EditUserParams{UserIdentification: temporalcloudcli.UserIdentificationOptions{}},
			makeRunEditor: func(t *testing.T) func(proto.Message, proto.Message) error {
				return func(existing, target proto.Message) error { return nil }
			},
			wantErrContains: "must provide either --user-id or --user-email",
		},
		{
			name: "both_identifiers",
			params: temporalcloudcli.EditUserParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1", UserEmail: "alice@example.com"},
			},
			makeRunEditor: func(t *testing.T) func(proto.Message, proto.Message) error {
				return func(existing, target proto.Message) error { return nil }
			},
			wantErrContains: "cannot provide both --user-id and --user-email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCloud := cloudmock.NewMockCloudServiceClient(t)
			mockPrompter := cmdmock.NewMockPrompter(t)
			mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
			if tt.setup != nil {
				tt.setup(mockCloud, mockPrompter, mockHandler)
			}
			params := tt.params
			params.Cloud = mockCloud
			params.Prompter = mockPrompter
			params.OperationHandler = mockHandler
			params.RunEditor = tt.makeRunEditor(t)

			err := temporalcloudcli.EditUser(context.Background(), params)
			switch {
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			case tt.wantErrContains != "":
				require.ErrorContains(t, err, tt.wantErrContains)
			default:
				require.NoError(t, err)
			}
		})
	}
}

func TestApplyUser(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-apply"}
	promptErr := errors.New("Aborting apply.")
	getUsersErr := errors.New("lookup error")
	createErr := errors.New("create error")

	tests := []struct {
		name            string
		params          temporalcloudcli.ApplyUserParams
		setup           func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler)
		wantErr         error
		wantErrContains string
	}{
		{
			name: "create_success",
			params: temporalcloudcli.ApplyUserParams{
				Spec: &identityv1.UserSpec{Email: "alice@example.com"},
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				spec := &identityv1.UserSpec{Email: "alice@example.com"}
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
					Return(&cloudservice.GetUsersResponse{Users: nil}, nil)
				prompter.EXPECT().PromptApply((*identityv1.UserSpec)(nil), spec, false).Return(nil)
				cloud.EXPECT().
					CreateUser(context.Background(), &cloudservice.CreateUserRequest{Spec: spec}).
					Return(&cloudservice.CreateUserResponse{UserId: "user-new", AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
		},
		{
			name: "update_success",
			params: temporalcloudcli.ApplyUserParams{
				Spec: &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
				},
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				newSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
				}
				existingSpec := &identityv1.UserSpec{
					Email:  "alice@example.com",
					Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
				}
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec}},
					}, nil)
				prompter.EXPECT().PromptApply(existingSpec, newSpec, false).Return(nil)
				cloud.EXPECT().
					UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            newSpec,
						ResourceVersion: "rv-1",
					}).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
		},
		{
			name: "update_resource_version_override",
			params: temporalcloudcli.ApplyUserParams{
				Spec:            &identityv1.UserSpec{Email: "alice@example.com"},
				ResourceVersion: "rv-override",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				spec := &identityv1.UserSpec{Email: "alice@example.com"}
				existingSpec := &identityv1.UserSpec{Email: "alice@example.com"}
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec}},
					}, nil)
				prompter.EXPECT().PromptApply(existingSpec, spec, false).Return(nil)
				cloud.EXPECT().
					UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
						UserId:          "user-1",
						Spec:            spec,
						ResourceVersion: "rv-override",
					}).
					Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
		},
		{
			name: "prompt_declined",
			params: temporalcloudcli.ApplyUserParams{
				Spec: &identityv1.UserSpec{Email: "alice@example.com"},
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				spec := &identityv1.UserSpec{Email: "alice@example.com"}
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
					Return(&cloudservice.GetUsersResponse{Users: nil}, nil)
				prompter.EXPECT().PromptApply((*identityv1.UserSpec)(nil), spec, false).Return(promptErr)
			},
			wantErr: promptErr,
		},
		{
			name: "get_users_error",
			params: temporalcloudcli.ApplyUserParams{
				Spec: &identityv1.UserSpec{Email: "alice@example.com"},
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
					Return(nil, getUsersErr)
			},
			wantErr: getUsersErr,
		},
		{
			name: "create_api_error",
			params: temporalcloudcli.ApplyUserParams{
				Spec: &identityv1.UserSpec{Email: "alice@example.com"},
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, prompter *cmdmock.MockPrompter, handler *cmdmock.MockAsyncOperationHandler) {
				spec := &identityv1.UserSpec{Email: "alice@example.com"}
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
					Return(&cloudservice.GetUsersResponse{Users: nil}, nil)
				prompter.EXPECT().PromptApply((*identityv1.UserSpec)(nil), spec, false).Return(nil)
				cloud.EXPECT().
					CreateUser(context.Background(), &cloudservice.CreateUserRequest{Spec: spec}).
					Return(nil, createErr)
				handler.EXPECT().HandleErr(createErr).Return(createErr)
			},
			wantErr: createErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCloud := cloudmock.NewMockCloudServiceClient(t)
			mockPrompter := cmdmock.NewMockPrompter(t)
			mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
			if tt.setup != nil {
				tt.setup(mockCloud, mockPrompter, mockHandler)
			}
			params := tt.params
			params.Cloud = mockCloud
			params.Prompter = mockPrompter
			params.OperationHandler = mockHandler

			err := temporalcloudcli.ApplyUser(context.Background(), params)
			switch {
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			case tt.wantErrContains != "":
				require.ErrorContains(t, err, tt.wantErrContains)
			default:
				require.NoError(t, err)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	apiErr := errors.New("api error")

	tests := []struct {
		name        string
		userID      string
		setup       func(cloud *cloudmock.MockCloudServiceClient)
		wantErr     error
		checkOutput func(t *testing.T, output string)
	}{
		{
			name:   "success",
			userID: "user-1",
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(&cloudservice.GetUserResponse{
						User: &identityv1.User{
							Id:   "user-1",
							Spec: &identityv1.UserSpec{Email: "alice@example.com"},
						},
					}, nil)
			},
			checkOutput: func(t *testing.T, output string) {
				type userOutput struct {
					Id   string `json:"id"`
					Spec struct {
						Email string `json:"email"`
					} `json:"spec"`
				}
				var out userOutput
				require.NoError(t, json.Unmarshal([]byte(output), &out))
				assert.Equal(t, "user-1", out.Id)
				assert.Equal(t, "alice@example.com", out.Spec.Email)
			},
		},
		{
			name:   "api_error",
			userID: "user-1",
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
					Return(nil, apiErr)
			},
			wantErr: apiErr,
			checkOutput: func(t *testing.T, output string) {
				assert.Empty(t, output)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCloud := cloudmock.NewMockCloudServiceClient(t)
			tt.setup(mockCloud)

			var buf bytes.Buffer
			err := temporalcloudcli.GetUser(context.Background(), temporalcloudcli.GetUserParams{
				UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: tt.userID},
				Cloud:   mockCloud,
				Printer: &printer.Printer{Output: &buf, JSON: true},
			})
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			if tt.checkOutput != nil {
				tt.checkOutput(t, buf.String())
			}
		})
	}
}

func TestInviteUser(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-invite"}
	apiErr := errors.New("api error")

	tests := []struct {
		name    string
		params  temporalcloudcli.InviteUserParams
		setup   func(cloud *cloudmock.MockCloudServiceClient, handler *cmdmock.MockAsyncOperationHandler)
		wantErr error
	}{
		{
			name: "success",
			params: temporalcloudcli.InviteUserParams{
				Email:         "alice@example.com",
				AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
				NmspaceAccesses: map[string]*identityv1.NamespaceAccess{
					"my-ns.my-account": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
				},
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient, handler *cmdmock.MockAsyncOperationHandler) {
				expectedSpec := &identityv1.UserSpec{
					Email: "alice@example.com",
					Access: &identityv1.Access{
						AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
						NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
							"my-ns.my-account": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
						},
					},
				}
				cloud.EXPECT().
					CreateUser(context.Background(), &cloudservice.CreateUserRequest{Spec: expectedSpec}).
					Return(&cloudservice.CreateUserResponse{UserId: "user-1", AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
		},
		{
			name:   "no_access",
			params: temporalcloudcli.InviteUserParams{Email: "bob@example.com"},
			setup: func(cloud *cloudmock.MockCloudServiceClient, handler *cmdmock.MockAsyncOperationHandler) {
				cloud.EXPECT().
					CreateUser(context.Background(), &cloudservice.CreateUserRequest{
						Spec: &identityv1.UserSpec{
							Email:  "bob@example.com",
							Access: &identityv1.Access{},
						},
					}).
					Return(&cloudservice.CreateUserResponse{UserId: "user-2", AsyncOperation: op}, nil)
				handler.EXPECT().Handle(op).Return(nil)
			},
		},
		{
			name:   "api_error",
			params: temporalcloudcli.InviteUserParams{Email: "alice@example.com"},
			setup: func(cloud *cloudmock.MockCloudServiceClient, handler *cmdmock.MockAsyncOperationHandler) {
				cloud.EXPECT().
					CreateUser(context.Background(), &cloudservice.CreateUserRequest{
						Spec: &identityv1.UserSpec{
							Email:  "alice@example.com",
							Access: &identityv1.Access{},
						},
					}).
					Return(nil, apiErr)
				handler.EXPECT().HandleErr(apiErr).Return(apiErr)
			},
			wantErr: apiErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCloud := cloudmock.NewMockCloudServiceClient(t)
			mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
			tt.setup(mockCloud, mockHandler)

			params := tt.params
			params.Cloud = mockCloud
			params.OperationHandler = mockHandler

			err := temporalcloudcli.InviteUser(context.Background(), params)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestListUsers(t *testing.T) {
	apiErr := errors.New("api error")

	tests := []struct {
		name        string
		params      temporalcloudcli.ListUsersParams
		setup       func(cloud *cloudmock.MockCloudServiceClient)
		wantErr     error
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "success",
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{}).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{
							{Id: "user-1", Spec: &identityv1.UserSpec{Email: "alice@example.com"}},
							{Id: "user-2", Spec: &identityv1.UserSpec{Email: "bob@example.com"}},
						},
					}, nil)
			},
			checkOutput: func(t *testing.T, output string) {
				type listOutput struct {
					Users []struct {
						Id string `json:"id"`
					} `json:"users"`
				}
				var out listOutput
				require.NoError(t, json.Unmarshal([]byte(output), &out))
				require.Len(t, out.Users, 2)
				assert.Equal(t, "user-1", out.Users[0].Id)
				assert.Equal(t, "user-2", out.Users[1].Id)
			},
		},
		{
			name: "with_filters",
			params: temporalcloudcli.ListUsersParams{
				Email:     "alice@example.com",
				Namespace: "my-namespace.my-account",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{
						Email:     "alice@example.com",
						Namespace: "my-namespace.my-account",
					}).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{
							{Id: "user-1", Spec: &identityv1.UserSpec{Email: "alice@example.com"}},
						},
					}, nil)
			},
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "user-1")
			},
		},
		{
			name: "with_pagination",
			params: temporalcloudcli.ListUsersParams{
				PageSize:  10,
				PageToken: "tok-abc",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{
						PageSize:  10,
						PageToken: "tok-abc",
					}).
					Return(&cloudservice.GetUsersResponse{
						Users:         []*identityv1.User{{Id: "user-3", Spec: &identityv1.UserSpec{Email: "carol@example.com"}}},
						NextPageToken: "tok-def",
					}, nil)
			},
			checkOutput: func(t *testing.T, output string) {
				type listOutput struct {
					NextPageToken string `json:"NextPageToken"`
				}
				var out listOutput
				require.NoError(t, json.Unmarshal([]byte(output), &out))
				assert.Equal(t, "tok-def", out.NextPageToken)
			},
		},
		{
			name: "api_error",
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUsers(context.Background(), &cloudservice.GetUsersRequest{}).
					Return(nil, apiErr)
			},
			wantErr: apiErr,
			checkOutput: func(t *testing.T, output string) {
				assert.Empty(t, output)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCloud := cloudmock.NewMockCloudServiceClient(t)
			tt.setup(mockCloud)

			var buf bytes.Buffer
			params := tt.params
			params.Cloud = mockCloud
			params.Printer = &printer.Printer{Output: &buf, JSON: true}

			err := temporalcloudcli.ListUsers(context.Background(), params)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			if tt.checkOutput != nil {
				tt.checkOutput(t, buf.String())
			}
		})
	}
}
