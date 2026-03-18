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

// TestSetNamespacePermissions_ByID_Success verifies that new permissions are merged with existing ones.
func TestSetNamespacePermissions_ByID_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	oldSpec := &identityv1.UserSpec{
		Email: "alice@example.com",
		Access: &identityv1.Access{
			NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
				"old-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
			},
		},
	}
	// old-ns.acct is preserved; my-ns.acct and other-ns.acct are added.
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

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{
			User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
		}, nil)

	mockPrompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-ns"}
	mockCloud.EXPECT().
		UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
			UserId:          "user-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.SetNamespacePermissions(context.Background(), temporalcloudcli.SetNamespacePermissionsParams{
		UserID:            "user-1",
		NamespaceAccesses: []string{"my-ns.acct=write", "other-ns.acct=read"},
		Cloud:             mockCloud,
		Prompter:          mockPrompter,
		OperationHandler:  mockHandler,
	})
	require.NoError(t, err)
}

// TestSetNamespacePermissions_ByEmail_Success verifies permissions are set when resolved by email.
func TestSetNamespacePermissions_ByEmail_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	oldSpec := &identityv1.UserSpec{Email: "alice@example.com"}
	newSpec := &identityv1.UserSpec{
		Email: "alice@example.com",
		Access: &identityv1.Access{
			NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
				"my-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_ADMIN},
			},
		},
	}

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
		Return(&cloudservice.GetUsersResponse{
			Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec}},
		}, nil)

	mockPrompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-ns"}
	mockCloud.EXPECT().
		UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
			UserId:          "user-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.SetNamespacePermissions(context.Background(), temporalcloudcli.SetNamespacePermissionsParams{
		UserEmail:         "alice@example.com",
		NamespaceAccesses: []string{"my-ns.acct=admin"},
		Cloud:             mockCloud,
		Prompter:          mockPrompter,
		OperationHandler:  mockHandler,
	})
	require.NoError(t, err)
}

// TestSetNamespacePermissions_InvalidPermission verifies that a bad permission string returns an error before any API call.
func TestSetNamespacePermissions_InvalidPermission(t *testing.T) {
	err := temporalcloudcli.SetNamespacePermissions(context.Background(), temporalcloudcli.SetNamespacePermissionsParams{
		UserID:            "user-1",
		NamespaceAccesses: []string{"my-ns.acct=superwrite"},
		Cloud:             cloudmock.NewMockCloudServiceClient(t),
		Prompter:          cmdmock.NewMockPrompter(t),
		OperationHandler:  cmdmock.NewMockAsyncOperationHandler(t),
	})
	require.ErrorContains(t, err, "invalid permission")
}

// TestSetNamespacePermissions_PromptDeclined verifies UpdateUser is not called when the prompt is declined.
func TestSetNamespacePermissions_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	oldSpec := &identityv1.UserSpec{Email: "alice@example.com"}
	newSpec := &identityv1.UserSpec{
		Email: "alice@example.com",
		Access: &identityv1.Access{
			NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
				"my-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
			},
		},
	}

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{
			User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
		}, nil)

	mockPrompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(promptErr)

	err := temporalcloudcli.SetNamespacePermissions(context.Background(), temporalcloudcli.SetNamespacePermissionsParams{
		UserID:            "user-1",
		NamespaceAccesses: []string{"my-ns.acct=write"},
		Cloud:             mockCloud,
		Prompter:          mockPrompter,
		OperationHandler:  mockHandler,
	})
	require.ErrorIs(t, err, promptErr)
}

// TestSetNamespacePermissions_NoIdentifier verifies an error is returned when neither identifier is set.
func TestSetNamespacePermissions_NoIdentifier(t *testing.T) {
	err := temporalcloudcli.SetNamespacePermissions(context.Background(), temporalcloudcli.SetNamespacePermissionsParams{
		NamespaceAccesses: []string{"my-ns.acct=write"},
		Cloud:             cloudmock.NewMockCloudServiceClient(t),
		Prompter:          cmdmock.NewMockPrompter(t),
		OperationHandler:  cmdmock.NewMockAsyncOperationHandler(t),
	})
	require.ErrorContains(t, err, "must provide either --user-id or --user-email")
}

// TestSetNamespacePermissions_BothIdentifiers verifies an error is returned when both identifiers are set.
func TestSetNamespacePermissions_BothIdentifiers(t *testing.T) {
	err := temporalcloudcli.SetNamespacePermissions(context.Background(), temporalcloudcli.SetNamespacePermissionsParams{
		UserID:            "user-1",
		UserEmail:         "alice@example.com",
		NamespaceAccesses: []string{"my-ns.acct=write"},
		Cloud:             cloudmock.NewMockCloudServiceClient(t),
		Prompter:          cmdmock.NewMockPrompter(t),
		OperationHandler:  cmdmock.NewMockAsyncOperationHandler(t),
	})
	require.ErrorContains(t, err, "cannot provide both --user-id and --user-email")
}

// TestSetNamespacePermissions_RemoveAccess verifies that an empty permission (e.g. "testns=") removes that namespace.
func TestSetNamespacePermissions_RemoveAccess(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	oldSpec := &identityv1.UserSpec{
		Email: "alice@example.com",
		Access: &identityv1.Access{
			NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
				"ns1.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
				"ns2.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
			},
		},
	}
	// ns1.acct is removed; ns2.acct is preserved.
	newSpec := &identityv1.UserSpec{
		Email: "alice@example.com",
		Access: &identityv1.Access{
			NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
				"ns2.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
			},
		},
	}

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{
			User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
		}, nil)

	mockPrompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-ns"}
	mockCloud.EXPECT().
		UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
			UserId:          "user-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.SetNamespacePermissions(context.Background(), temporalcloudcli.SetNamespacePermissionsParams{
		UserID:            "user-1",
		NamespaceAccesses: []string{"ns1.acct="},
		Cloud:             mockCloud,
		Prompter:          mockPrompter,
		OperationHandler:  mockHandler,
	})
	require.NoError(t, err)
}

// TestSetAccountRole_ByID_Success verifies that SetAccountRole updates the account role when resolved by ID.
func TestSetAccountRole_ByID_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	oldSpec := &identityv1.UserSpec{
		Email:  "alice@example.com",
		Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
	}
	newSpec := &identityv1.UserSpec{
		Email:  "alice@example.com",
		Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
	}

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{
			User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
		}, nil)

	mockPrompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-role"}
	mockCloud.EXPECT().
		UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
			UserId:          "user-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.SetAccountRole(context.Background(), temporalcloudcli.SetAccountRoleParams{
		UserID:           "user-1",
		AccountRole:      "admin",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestSetAccountRole_ByEmail_Success verifies SetAccountRole resolves the user via GetUsers when --user-email is provided.
func TestSetAccountRole_ByEmail_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	oldSpec := &identityv1.UserSpec{
		Email:  "alice@example.com",
		Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
	}
	newSpec := &identityv1.UserSpec{
		Email:  "alice@example.com",
		Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_READ}},
	}

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
		Return(&cloudservice.GetUsersResponse{
			Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec}},
		}, nil)

	mockPrompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-role"}
	mockCloud.EXPECT().
		UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
			UserId:          "user-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.SetAccountRole(context.Background(), temporalcloudcli.SetAccountRoleParams{
		UserEmail:        "alice@example.com",
		AccountRole:      "read",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestSetAccountRole_NilAccess verifies SetAccountRole correctly initialises Access when it is nil on the existing spec.
func TestSetAccountRole_NilAccess(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	oldSpec := &identityv1.UserSpec{Email: "alice@example.com"}
	newSpec := &identityv1.UserSpec{
		Email:  "alice@example.com",
		Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
	}

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{
			User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
		}, nil)

	mockPrompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-role"}
	mockCloud.EXPECT().
		UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
			UserId:          "user-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.SetAccountRole(context.Background(), temporalcloudcli.SetAccountRoleParams{
		UserID:           "user-1",
		AccountRole:      "developer",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestSetAccountRole_InvalidRole verifies that an unrecognised role returns an error before any API call.
func TestSetAccountRole_InvalidRole(t *testing.T) {
	err := temporalcloudcli.SetAccountRole(context.Background(), temporalcloudcli.SetAccountRoleParams{
		UserID:           "user-1",
		AccountRole:      "superuser",
		Cloud:            cloudmock.NewMockCloudServiceClient(t),
		Prompter:         cmdmock.NewMockPrompter(t),
		OperationHandler: cmdmock.NewMockAsyncOperationHandler(t),
	})
	require.ErrorContains(t, err, "invalid account role")
}

// TestSetAccountRole_PromptDeclined verifies UpdateUser is not called when the prompt is declined.
func TestSetAccountRole_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	oldSpec := &identityv1.UserSpec{
		Email:  "alice@example.com",
		Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
	}
	newSpec := &identityv1.UserSpec{
		Email:  "alice@example.com",
		Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
	}

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{
			User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: oldSpec},
		}, nil)

	mockPrompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(promptErr)

	err := temporalcloudcli.SetAccountRole(context.Background(), temporalcloudcli.SetAccountRoleParams{
		UserID:           "user-1",
		AccountRole:      "admin",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, promptErr)
}

// TestSetAccountRole_NoIdentifier verifies an error is returned when neither identifier is set.
func TestSetAccountRole_NoIdentifier(t *testing.T) {
	err := temporalcloudcli.SetAccountRole(context.Background(), temporalcloudcli.SetAccountRoleParams{
		AccountRole:      "admin",
		Cloud:            cloudmock.NewMockCloudServiceClient(t),
		Prompter:         cmdmock.NewMockPrompter(t),
		OperationHandler: cmdmock.NewMockAsyncOperationHandler(t),
	})
	require.ErrorContains(t, err, "must provide either --user-id or --user-email")
}

// TestSetAccountRole_BothIdentifiers verifies an error is returned when both identifiers are set.
func TestSetAccountRole_BothIdentifiers(t *testing.T) {
	err := temporalcloudcli.SetAccountRole(context.Background(), temporalcloudcli.SetAccountRoleParams{
		UserID:           "user-1",
		UserEmail:        "alice@example.com",
		AccountRole:      "admin",
		Cloud:            cloudmock.NewMockCloudServiceClient(t),
		Prompter:         cmdmock.NewMockPrompter(t),
		OperationHandler: cmdmock.NewMockAsyncOperationHandler(t),
	})
	require.ErrorContains(t, err, "cannot provide both --user-id and --user-email")
}

// TestDeleteUser_ByID_Success verifies DeleteUser calls DeleteUser SDK with the resolved resource version.
func TestDeleteUser_ByID_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{
			User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1"},
		}, nil)

	op := &operation.AsyncOperation{Id: "op-delete"}
	mockCloud.EXPECT().
		DeleteUser(context.Background(), &cloudservice.DeleteUserRequest{
			UserId:          "user-1",
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.DeleteUserResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.DeleteUser(context.Background(), temporalcloudcli.DeleteUserParams{
		UserID:           "user-1",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestDeleteUser_ByEmail_Success verifies DeleteUser resolves the user via GetUsers when --user-email is provided.
func TestDeleteUser_ByEmail_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
		Return(&cloudservice.GetUsersResponse{
			Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1"}},
		}, nil)

	op := &operation.AsyncOperation{Id: "op-delete"}
	mockCloud.EXPECT().
		DeleteUser(context.Background(), &cloudservice.DeleteUserRequest{
			UserId:          "user-1",
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.DeleteUserResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.DeleteUser(context.Background(), temporalcloudcli.DeleteUserParams{
		UserEmail:        "alice@example.com",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestDeleteUser_ResourceVersionOverride verifies --resource-version overrides the fetched version.
func TestDeleteUser_ResourceVersionOverride(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{
			User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1"},
		}, nil)

	op := &operation.AsyncOperation{Id: "op-delete"}
	mockCloud.EXPECT().
		DeleteUser(context.Background(), &cloudservice.DeleteUserRequest{
			UserId:          "user-1",
			ResourceVersion: "rv-override",
		}).
		Return(&cloudservice.DeleteUserResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.DeleteUser(context.Background(), temporalcloudcli.DeleteUserParams{
		UserID:           "user-1",
		ResourceVersion:  "rv-override",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestDeleteUser_ByEmail_NotFound verifies an error is returned when the email matches no user.
func TestDeleteUser_ByEmail_NotFound(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "nobody@example.com"}).
		Return(&cloudservice.GetUsersResponse{Users: nil}, nil)

	err := temporalcloudcli.DeleteUser(context.Background(), temporalcloudcli.DeleteUserParams{
		UserEmail:        "nobody@example.com",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.ErrorContains(t, err, "no user found with email")
}

// TestDeleteUser_NoIdentifier verifies an error is returned when neither --user-id nor --user-email is set.
func TestDeleteUser_NoIdentifier(t *testing.T) {
	err := temporalcloudcli.DeleteUser(context.Background(), temporalcloudcli.DeleteUserParams{
		Cloud:            cloudmock.NewMockCloudServiceClient(t),
		OperationHandler: cmdmock.NewMockAsyncOperationHandler(t),
	})
	require.ErrorContains(t, err, "must provide either --user-id or --user-email")
}

// TestDeleteUser_BothIdentifiers verifies an error is returned when both identifiers are set.
func TestDeleteUser_BothIdentifiers(t *testing.T) {
	err := temporalcloudcli.DeleteUser(context.Background(), temporalcloudcli.DeleteUserParams{
		UserID:           "user-1",
		UserEmail:        "alice@example.com",
		Cloud:            cloudmock.NewMockCloudServiceClient(t),
		OperationHandler: cmdmock.NewMockAsyncOperationHandler(t),
	})
	require.ErrorContains(t, err, "cannot provide both --user-id and --user-email")
}

// TestDeleteUser_APIError verifies a DeleteUser failure is forwarded through HandleErr.
func TestDeleteUser_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("delete error")

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{
			User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1"},
		}, nil)

	mockCloud.EXPECT().
		DeleteUser(context.Background(), &cloudservice.DeleteUserRequest{
			UserId:          "user-1",
			ResourceVersion: "rv-1",
		}).
		Return(nil, apiErr)

	mockHandler.EXPECT().HandleErr(apiErr).Return(apiErr)

	err := temporalcloudcli.DeleteUser(context.Background(), temporalcloudcli.DeleteUserParams{
		UserID:           "user-1",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

// TestEditUser_Success verifies that EditUser fetches the user, opens the editor, prompts, and calls UpdateUser.
func TestEditUser_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	existingSpec := &identityv1.UserSpec{
		Email:  "alice@example.com",
		Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
	}
	editedSpec := &identityv1.UserSpec{
		Email:  "alice@example.com",
		Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
	}

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{
			User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec},
		}, nil)

	mockPrompter.EXPECT().PromptApply(existingSpec, editedSpec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-edit"}
	mockCloud.EXPECT().
		UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
			UserId:          "user-1",
			Spec:            editedSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.EditUser(context.Background(), temporalcloudcli.EditUserParams{
		UserID:           "user-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
		RunEditor: func(existing, target proto.Message) error {
			proto.Merge(target, editedSpec)
			return nil
		},
	})
	require.NoError(t, err)
}

// TestEditUser_ResourceVersionOverride verifies that --resource-version overrides the fetched version.
func TestEditUser_ResourceVersionOverride(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	existingSpec := &identityv1.UserSpec{Email: "alice@example.com"}
	editedSpec := &identityv1.UserSpec{
		Email:  "alice@example.com",
		Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_READ}},
	}

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{
			User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec},
		}, nil)

	mockPrompter.EXPECT().PromptApply(existingSpec, editedSpec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-edit"}
	mockCloud.EXPECT().
		UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
			UserId:          "user-1",
			Spec:            editedSpec,
			ResourceVersion: "rv-override",
		}).
		Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.EditUser(context.Background(), temporalcloudcli.EditUserParams{
		UserID:          "user-1",
		ResourceVersion: "rv-override",
		Cloud:           mockCloud,
		Prompter:        mockPrompter,
		OperationHandler: mockHandler,
		RunEditor: func(existing, target proto.Message) error {
			proto.Merge(target, editedSpec)
			return nil
		},
	})
	require.NoError(t, err)
}

// TestEditUser_PromptDeclined verifies that UpdateUser is not called when the prompt is declined.
func TestEditUser_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	existingSpec := &identityv1.UserSpec{Email: "alice@example.com"}
	editedSpec := &identityv1.UserSpec{
		Email:  "alice@example.com",
		Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
	}

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{
			User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec},
		}, nil)

	mockPrompter.EXPECT().PromptApply(existingSpec, editedSpec, false).Return(promptErr)

	err := temporalcloudcli.EditUser(context.Background(), temporalcloudcli.EditUserParams{
		UserID:           "user-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
		RunEditor: func(existing, target proto.Message) error {
			proto.Merge(target, editedSpec)
			return nil
		},
	})
	require.ErrorIs(t, err, promptErr)
}

// TestEditUser_GetUserError verifies that a GetUser error propagates before opening the editor.
func TestEditUser_GetUserError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("not found")

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(nil, apiErr)

	editorCalled := false
	err := temporalcloudcli.EditUser(context.Background(), temporalcloudcli.EditUserParams{
		UserID:           "user-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
		RunEditor: func(existing, target proto.Message) error {
			editorCalled = true
			return nil
		},
	})
	require.ErrorIs(t, err, apiErr)
	assert.False(t, editorCalled)
}

// TestEditUser_EditorError verifies that an editor failure propagates before calling UpdateUser.
func TestEditUser_EditorError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	editorErr := errors.New("editor failed")

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{
			User: &identityv1.User{Id: "user-1", ResourceVersion: "rv-1", Spec: &identityv1.UserSpec{Email: "alice@example.com"}},
		}, nil)

	err := temporalcloudcli.EditUser(context.Background(), temporalcloudcli.EditUserParams{
		UserID:           "user-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
		RunEditor: func(existing, target proto.Message) error {
			return editorErr
		},
	})
	require.ErrorIs(t, err, editorErr)
}

// TestEditUser_ByEmail_Success verifies that EditUser resolves the user via GetUsers when --user-email is provided.
func TestEditUser_ByEmail_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	existingSpec := &identityv1.UserSpec{
		Email:  "alice@example.com",
		Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
	}
	editedSpec := &identityv1.UserSpec{
		Email:  "alice@example.com",
		Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
	}

	// Lookup by email instead of GetUser.
	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
		Return(&cloudservice.GetUsersResponse{
			Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec}},
		}, nil)

	mockPrompter.EXPECT().PromptApply(existingSpec, editedSpec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-edit"}
	mockCloud.EXPECT().
		UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
			UserId:          "user-1",
			Spec:            editedSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.EditUser(context.Background(), temporalcloudcli.EditUserParams{
		UserEmail:        "alice@example.com",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
		RunEditor: func(existing, target proto.Message) error {
			proto.Merge(target, editedSpec)
			return nil
		},
	})
	require.NoError(t, err)
}

// TestEditUser_ByEmail_NotFound verifies that an error is returned when the email matches no user.
func TestEditUser_ByEmail_NotFound(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "nobody@example.com"}).
		Return(&cloudservice.GetUsersResponse{Users: nil}, nil)

	err := temporalcloudcli.EditUser(context.Background(), temporalcloudcli.EditUserParams{
		UserEmail:        "nobody@example.com",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
		RunEditor: func(existing, target proto.Message) error { return nil },
	})
	require.ErrorContains(t, err, "no user found with email")
}

// TestEditUser_NoIdentifier verifies that an error is returned when neither --user-id nor --user-email is set.
func TestEditUser_NoIdentifier(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	err := temporalcloudcli.EditUser(context.Background(), temporalcloudcli.EditUserParams{
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
		RunEditor:        func(existing, target proto.Message) error { return nil },
	})
	require.ErrorContains(t, err, "must provide either --user-id or --user-email")
}

// TestEditUser_BothIdentifiers verifies that an error is returned when both identifiers are set.
func TestEditUser_BothIdentifiers(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	err := temporalcloudcli.EditUser(context.Background(), temporalcloudcli.EditUserParams{
		UserID:           "user-1",
		UserEmail:        "alice@example.com",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
		RunEditor:        func(existing, target proto.Message) error { return nil },
	})
	require.ErrorContains(t, err, "cannot provide both --user-id and --user-email")
}

// TestApplyUser_Create_Success verifies that ApplyUser calls CreateUser when the email is not found.
func TestApplyUser_Create_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	spec := &identityv1.UserSpec{Email: "alice@example.com"}

	// User not found
	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
		Return(&cloudservice.GetUsersResponse{Users: nil}, nil)

	mockPrompter.EXPECT().PromptApply((*identityv1.UserSpec)(nil), spec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-create"}
	mockCloud.EXPECT().
		CreateUser(context.Background(), &cloudservice.CreateUserRequest{Spec: spec}).
		Return(&cloudservice.CreateUserResponse{UserId: "user-new", AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.ApplyUser(context.Background(), temporalcloudcli.ApplyUserParams{
		Spec:             spec,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestApplyUser_Update_Success verifies that ApplyUser calls UpdateUser when the email is found.
func TestApplyUser_Update_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	newSpec := &identityv1.UserSpec{
		Email:  "alice@example.com",
		Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
	}
	existingSpec := &identityv1.UserSpec{
		Email:  "alice@example.com",
		Access: &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
	}

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
		Return(&cloudservice.GetUsersResponse{
			Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec}},
		}, nil)

	mockPrompter.EXPECT().PromptApply(existingSpec, newSpec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-update"}
	mockCloud.EXPECT().
		UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
			UserId:          "user-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.ApplyUser(context.Background(), temporalcloudcli.ApplyUserParams{
		Spec:             newSpec,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestApplyUser_Update_ResourceVersionOverride verifies that --resource-version overrides the fetched version.
func TestApplyUser_Update_ResourceVersionOverride(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	spec := &identityv1.UserSpec{Email: "alice@example.com"}
	existingSpec := &identityv1.UserSpec{Email: "alice@example.com"}

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
		Return(&cloudservice.GetUsersResponse{
			Users: []*identityv1.User{{Id: "user-1", ResourceVersion: "rv-1", Spec: existingSpec}},
		}, nil)

	mockPrompter.EXPECT().PromptApply(existingSpec, spec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-update"}
	mockCloud.EXPECT().
		UpdateUser(context.Background(), &cloudservice.UpdateUserRequest{
			UserId:          "user-1",
			Spec:            spec,
			ResourceVersion: "rv-override",
		}).
		Return(&cloudservice.UpdateUserResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.ApplyUser(context.Background(), temporalcloudcli.ApplyUserParams{
		Spec:            spec,
		ResourceVersion: "rv-override",
		Cloud:           mockCloud,
		Prompter:        mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestApplyUser_PromptDeclined verifies that neither CreateUser nor UpdateUser is called when the prompt is declined.
func TestApplyUser_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	spec := &identityv1.UserSpec{Email: "alice@example.com"}

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
		Return(&cloudservice.GetUsersResponse{Users: nil}, nil)

	mockPrompter.EXPECT().PromptApply((*identityv1.UserSpec)(nil), spec, false).Return(promptErr)

	err := temporalcloudcli.ApplyUser(context.Background(), temporalcloudcli.ApplyUserParams{
		Spec:             spec,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, promptErr)
}

// TestApplyUser_GetUsersError verifies that a GetUsers failure propagates before any mutation.
func TestApplyUser_GetUsersError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("lookup error")

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
		Return(nil, apiErr)

	err := temporalcloudcli.ApplyUser(context.Background(), temporalcloudcli.ApplyUserParams{
		Spec:             &identityv1.UserSpec{Email: "alice@example.com"},
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

// TestApplyUser_CreateAPIError verifies that a CreateUser failure is forwarded through HandleErr.
func TestApplyUser_CreateAPIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("create error")

	spec := &identityv1.UserSpec{Email: "alice@example.com"}

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
		Return(&cloudservice.GetUsersResponse{Users: nil}, nil)

	mockPrompter.EXPECT().PromptApply((*identityv1.UserSpec)(nil), spec, false).Return(nil)

	mockCloud.EXPECT().
		CreateUser(context.Background(), &cloudservice.CreateUserRequest{Spec: spec}).
		Return(nil, apiErr)

	mockHandler.EXPECT().HandleErr(apiErr).Return(apiErr)

	err := temporalcloudcli.ApplyUser(context.Background(), temporalcloudcli.ApplyUserParams{
		Spec:             spec,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

// TestGetUser_Success verifies that GetUser prints the user returned by the API.
func TestGetUser_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{
			User: &identityv1.User{
				Id:   "user-1",
				Spec: &identityv1.UserSpec{Email: "alice@example.com"},
			},
		}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.GetUser(context.Background(), temporalcloudcli.GetUserParams{
		UserID:  "user-1",
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)

	type userOutput struct {
		Id   string `json:"id"`
		Spec struct {
			Email string `json:"email"`
		} `json:"spec"`
	}
	var out userOutput
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "user-1", out.Id)
	assert.Equal(t, "alice@example.com", out.Spec.Email)
}

// TestGetUser_APIError verifies that a GetUser error propagates.
func TestGetUser_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(nil, apiErr)

	var buf bytes.Buffer
	err := temporalcloudcli.GetUser(context.Background(), temporalcloudcli.GetUserParams{
		UserID:  "user-1",
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.ErrorIs(t, err, apiErr)
	assert.Empty(t, buf.String())
}

// TestInviteUser_Success verifies CreateUser is called with the correct spec and the async op is handled.
func TestInviteUser_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	expectedSpec := &identityv1.UserSpec{
		Email: "alice@example.com",
		Access: &identityv1.Access{
			AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
			NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
				"my-ns.my-account": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
			},
		},
	}

	op := &operation.AsyncOperation{Id: "op-123"}
	mockCloud.EXPECT().
		CreateUser(context.Background(), &cloudservice.CreateUserRequest{Spec: expectedSpec}).
		Return(&cloudservice.CreateUserResponse{UserId: "user-1", AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.InviteUser(context.Background(), temporalcloudcli.InviteUserParams{
		Email:         "alice@example.com",
		AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
		NmspaceAccesses: map[string]*identityv1.NamespaceAccess{
			"my-ns.my-account": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
		},
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestInviteUser_NoAccess verifies CreateUser works with no role or namespace access specified.
func TestInviteUser_NoAccess(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	op := &operation.AsyncOperation{Id: "op-456"}
	mockCloud.EXPECT().
		CreateUser(context.Background(), &cloudservice.CreateUserRequest{
			Spec: &identityv1.UserSpec{
				Email:  "bob@example.com",
				Access: &identityv1.Access{},
			},
		}).
		Return(&cloudservice.CreateUserResponse{UserId: "user-2", AsyncOperation: op}, nil)

	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.InviteUser(context.Background(), temporalcloudcli.InviteUserParams{
		Email:            "bob@example.com",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// TestInviteUser_APIError verifies that a CreateUser error is forwarded through HandleErr.
func TestInviteUser_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		CreateUser(context.Background(), &cloudservice.CreateUserRequest{
			Spec: &identityv1.UserSpec{
				Email:  "alice@example.com",
				Access: &identityv1.Access{},
			},
		}).
		Return(nil, apiErr)

	mockHandler.EXPECT().HandleErr(apiErr).Return(apiErr)

	err := temporalcloudcli.InviteUser(context.Background(), temporalcloudcli.InviteUserParams{
		Email:            "alice@example.com",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

// TestListUsers_Success verifies that ListUsers prints users returned by the API.
func TestListUsers_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{}).
		Return(&cloudservice.GetUsersResponse{
			Users: []*identityv1.User{
				{
					Id:   "user-1",
					Spec: &identityv1.UserSpec{Email: "alice@example.com"},
				},
				{
					Id:   "user-2",
					Spec: &identityv1.UserSpec{Email: "bob@example.com"},
				},
			},
			NextPageToken: "",
		}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.ListUsers(context.Background(), temporalcloudcli.ListUsersParams{
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)

	type listOutput struct {
		Users []struct {
			Id string `json:"id"`
		} `json:"users"`
	}
	var out listOutput
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	require.Len(t, out.Users, 2)
	assert.Equal(t, "user-1", out.Users[0].Id)
	assert.Equal(t, "user-2", out.Users[1].Id)
}

// TestListUsers_WithFilters verifies that email and namespace filters are forwarded to the API.
func TestListUsers_WithFilters(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{
			Email:     "alice@example.com",
			Namespace: "my-namespace.my-account",
		}).
		Return(&cloudservice.GetUsersResponse{
			Users: []*identityv1.User{
				{
					Id:   "user-1",
					Spec: &identityv1.UserSpec{Email: "alice@example.com"},
				},
			},
		}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.ListUsers(context.Background(), temporalcloudcli.ListUsersParams{
		Email:     "alice@example.com",
		Namespace: "my-namespace.my-account",
		Cloud:     mockCloud,
		Printer:   &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "user-1")
}

// TestListUsers_WithPagination verifies that page-size and page-token are forwarded and NextPageToken is returned.
func TestListUsers_WithPagination(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{
			PageSize:  10,
			PageToken: "tok-abc",
		}).
		Return(&cloudservice.GetUsersResponse{
			Users:         []*identityv1.User{{Id: "user-3", Spec: &identityv1.UserSpec{Email: "carol@example.com"}}},
			NextPageToken: "tok-def",
		}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.ListUsers(context.Background(), temporalcloudcli.ListUsersParams{
		PageSize:  10,
		PageToken: "tok-abc",
		Cloud:     mockCloud,
		Printer:   &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)

	type listOutput struct {
		NextPageToken string `json:"NextPageToken"`
	}
	var out listOutput
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "tok-def", out.NextPageToken)
}

// TestListUsers_APIError verifies that a GetUsers error propagates.
func TestListUsers_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{}).
		Return(nil, apiErr)

	var buf bytes.Buffer
	err := temporalcloudcli.ListUsers(context.Background(), temporalcloudcli.ListUsersParams{
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.ErrorIs(t, err, apiErr)
	assert.Empty(t, buf.String())
}
