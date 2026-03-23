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

// ---- DeleteUserGroup ----

func TestDeleteUserGroup_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-del"}

	mockCloud.EXPECT().
		GetUserGroup(context.Background(), &cloudservice.GetUserGroupRequest{GroupId: "group-1"}).
		Return(&cloudservice.GetUserGroupResponse{
			Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1"},
		}, nil)
	mockCloud.EXPECT().
		DeleteUserGroup(context.Background(), &cloudservice.DeleteUserGroupRequest{
			GroupId:         "group-1",
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.DeleteUserGroupResponse{AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.DeleteUserGroup(context.Background(), temporalcloudcli.DeleteUserGroupParams{
		GroupId:          "group-1",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestDeleteUserGroup_ResourceVersionOverride(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-del"}

	mockCloud.EXPECT().
		GetUserGroup(context.Background(), &cloudservice.GetUserGroupRequest{GroupId: "group-1"}).
		Return(&cloudservice.GetUserGroupResponse{
			Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-server"},
		}, nil)
	mockCloud.EXPECT().
		DeleteUserGroup(context.Background(), &cloudservice.DeleteUserGroupRequest{
			GroupId:         "group-1",
			ResourceVersion: "rv-override",
		}).
		Return(&cloudservice.DeleteUserGroupResponse{AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.DeleteUserGroup(context.Background(), temporalcloudcli.DeleteUserGroupParams{
		GroupId:          "group-1",
		ResourceVersion:  "rv-override",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestDeleteUserGroup_GetError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("not found")

	mockCloud.EXPECT().
		GetUserGroup(context.Background(), &cloudservice.GetUserGroupRequest{GroupId: "group-1"}).
		Return(nil, apiErr)

	err := temporalcloudcli.DeleteUserGroup(context.Background(), temporalcloudcli.DeleteUserGroupParams{
		GroupId:          "group-1",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

// ---- UpdateUserGroup ----

func TestUpdateUserGroup_AccountRole(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-upd"}

	existingSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
	}
	newSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
		Access: &identityv1.Access{
			AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
		},
	}

	mockCloud.EXPECT().
		GetUserGroup(context.Background(), &cloudservice.GetUserGroupRequest{GroupId: "group-1"}).
		Return(&cloudservice.GetUserGroupResponse{
			Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1", Spec: existingSpec},
		}, nil)
	mockPrompter.EXPECT().PromptApply(existingSpec, newSpec, false).Return(nil)
	mockCloud.EXPECT().
		UpdateUserGroup(context.Background(), &cloudservice.UpdateUserGroupRequest{
			GroupId:         "group-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateUserGroupResponse{AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.UpdateUserGroup(context.Background(), temporalcloudcli.UpdateUserGroupParams{
		GroupId:          "group-1",
		AccountRole:      "developer",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestUpdateUserGroup_NamespaceAccess(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-upd"}

	existingSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		Access: &identityv1.Access{
			NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
				"old-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
			},
		},
	}
	newSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		Access: &identityv1.Access{
			NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
				"old-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
				"new-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
			},
		},
	}

	mockCloud.EXPECT().
		GetUserGroup(context.Background(), &cloudservice.GetUserGroupRequest{GroupId: "group-1"}).
		Return(&cloudservice.GetUserGroupResponse{
			Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1", Spec: existingSpec},
		}, nil)
	mockPrompter.EXPECT().PromptApply(existingSpec, newSpec, false).Return(nil)
	mockCloud.EXPECT().
		UpdateUserGroup(context.Background(), &cloudservice.UpdateUserGroupRequest{
			GroupId:         "group-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateUserGroupResponse{AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.UpdateUserGroup(context.Background(), temporalcloudcli.UpdateUserGroupParams{
		GroupId:           "group-1",
		NamespaceAccesses: []string{"new-ns.acct=write"},
		Cloud:             mockCloud,
		Prompter:          mockPrompter,
		OperationHandler:  mockHandler,
	})
	require.NoError(t, err)
}

func TestUpdateUserGroup_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	existingSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		Access:      &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN}},
	}
	newSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		Access:      &identityv1.Access{AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER}},
	}

	mockCloud.EXPECT().
		GetUserGroup(context.Background(), &cloudservice.GetUserGroupRequest{GroupId: "group-1"}).
		Return(&cloudservice.GetUserGroupResponse{
			Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1", Spec: existingSpec},
		}, nil)
	mockPrompter.EXPECT().PromptApply(existingSpec, newSpec, false).Return(promptErr)

	err := temporalcloudcli.UpdateUserGroup(context.Background(), temporalcloudcli.UpdateUserGroupParams{
		GroupId:          "group-1",
		AccountRole:      "developer",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, promptErr)
}

func TestUpdateUserGroup_NoFlags(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	err := temporalcloudcli.UpdateUserGroup(context.Background(), temporalcloudcli.UpdateUserGroupParams{
		GroupId:          "group-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorContains(t, err, "must provide at least one of --account-role or --namespace-access")
}

func TestUpdateUserGroup_InvalidRole(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	err := temporalcloudcli.UpdateUserGroup(context.Background(), temporalcloudcli.UpdateUserGroupParams{
		GroupId:          "group-1",
		AccountRole:      "superadmin",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorContains(t, err, "invalid account role")
}

// ---- EditUserGroup ----

func TestEditUserGroup_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-edit"}

	existingSpec := &identityv1.UserGroupSpec{DisplayName: "Engineering"}
	editedSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
	}

	mockCloud.EXPECT().
		GetUserGroup(context.Background(), &cloudservice.GetUserGroupRequest{GroupId: "group-1"}).
		Return(&cloudservice.GetUserGroupResponse{
			Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1", Spec: existingSpec},
		}, nil)
	mockPrompter.EXPECT().PromptApply(existingSpec, editedSpec, false).Return(nil)
	mockCloud.EXPECT().
		UpdateUserGroup(context.Background(), &cloudservice.UpdateUserGroupRequest{
			GroupId:         "group-1",
			Spec:            editedSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateUserGroupResponse{AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.EditUserGroup(context.Background(), temporalcloudcli.EditUserGroupParams{
		GroupId:          "group-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
		RunEditor: func(_, target proto.Message) error {
			proto.Merge(target, editedSpec)
			return nil
		},
	})
	require.NoError(t, err)
}

// ---- SetUserGroupAccountRole ----

func TestSetUserGroupAccountRole_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-role"}

	oldSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		Access:      &identityv1.Access{},
	}
	newSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		Access: &identityv1.Access{
			AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
		},
	}

	mockCloud.EXPECT().
		GetUserGroup(context.Background(), &cloudservice.GetUserGroupRequest{GroupId: "group-1"}).
		Return(&cloudservice.GetUserGroupResponse{
			Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1", Spec: oldSpec},
		}, nil)
	mockPrompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)
	mockCloud.EXPECT().
		UpdateUserGroup(context.Background(), &cloudservice.UpdateUserGroupRequest{
			GroupId:         "group-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateUserGroupResponse{AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.SetUserGroupAccountRole(context.Background(), temporalcloudcli.SetUserGroupAccountRoleParams{
		GroupId:          "group-1",
		AccountRole:      "developer",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestSetUserGroupAccountRole_InvalidRole(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	err := temporalcloudcli.SetUserGroupAccountRole(context.Background(), temporalcloudcli.SetUserGroupAccountRoleParams{
		GroupId:          "group-1",
		AccountRole:      "invalid-role",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorContains(t, err, "invalid account role")
}

// ---- SetUserGroupNamespacePermissions ----

func TestSetUserGroupNamespacePermissions_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-ns"}

	oldSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		Access: &identityv1.Access{
			NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
				"old-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
			},
		},
	}
	newSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		Access: &identityv1.Access{
			NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
				"old-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
				"my-ns.acct":  {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
			},
		},
	}

	mockCloud.EXPECT().
		GetUserGroup(context.Background(), &cloudservice.GetUserGroupRequest{GroupId: "group-1"}).
		Return(&cloudservice.GetUserGroupResponse{
			Group: &identityv1.UserGroup{Id: "group-1", ResourceVersion: "rv-1", Spec: oldSpec},
		}, nil)
	mockPrompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)
	mockCloud.EXPECT().
		UpdateUserGroup(context.Background(), &cloudservice.UpdateUserGroupRequest{
			GroupId:         "group-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateUserGroupResponse{AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.SetUserGroupNamespacePermissions(context.Background(), temporalcloudcli.SetUserGroupNamespacePermissionsParams{
		GroupId:           "group-1",
		NamespaceAccesses: []string{"my-ns.acct=write"},
		Cloud:             mockCloud,
		Prompter:          mockPrompter,
		OperationHandler:  mockHandler,
	})
	require.NoError(t, err)
}

func TestSetUserGroupNamespacePermissions_InvalidFormat(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	err := temporalcloudcli.SetUserGroupNamespacePermissions(context.Background(), temporalcloudcli.SetUserGroupNamespacePermissionsParams{
		GroupId:           "group-1",
		NamespaceAccesses: []string{"no-equals-sign"},
		Cloud:             mockCloud,
		Prompter:          mockPrompter,
		OperationHandler:  mockHandler,
	})
	require.ErrorContains(t, err, "invalid namespace-access")
}

func TestApplyUserGroup_Create(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-1"}

	spec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
	}

	// No existing group found.
	mockCloud.EXPECT().
		GetUserGroups(context.Background(), &cloudservice.GetUserGroupsRequest{DisplayName: "Engineering"}).
		Return(&cloudservice.GetUserGroupsResponse{}, nil)
	mockPrompter.EXPECT().PromptApply((*identityv1.UserGroupSpec)(nil), spec, false).Return(nil)
	mockCloud.EXPECT().
		CreateUserGroup(context.Background(), &cloudservice.CreateUserGroupRequest{Spec: spec}).
		Return(&cloudservice.CreateUserGroupResponse{GroupId: "group-1", AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.ApplyUserGroup(context.Background(), temporalcloudcli.ApplyUserGroupParams{
		Spec:             spec,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestApplyUserGroup_Update(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-2"}

	existingSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
	}
	newSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
		Access: &identityv1.Access{
			AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
		},
	}

	mockCloud.EXPECT().
		GetUserGroups(context.Background(), &cloudservice.GetUserGroupsRequest{DisplayName: "Engineering"}).
		Return(&cloudservice.GetUserGroupsResponse{
			Groups: []*identityv1.UserGroup{
				{Id: "group-1", ResourceVersion: "rv-1", Spec: existingSpec},
			},
		}, nil)
	mockPrompter.EXPECT().PromptApply(existingSpec, newSpec, false).Return(nil)
	mockCloud.EXPECT().
		UpdateUserGroup(context.Background(), &cloudservice.UpdateUserGroupRequest{
			GroupId:         "group-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateUserGroupResponse{AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.ApplyUserGroup(context.Background(), temporalcloudcli.ApplyUserGroupParams{
		Spec:             newSpec,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestApplyUserGroup_UpdateWithResourceVersion(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-3"}

	existingSpec := &identityv1.UserGroupSpec{DisplayName: "Engineering"}
	newSpec := &identityv1.UserGroupSpec{DisplayName: "Engineering"}

	mockCloud.EXPECT().
		GetUserGroups(context.Background(), &cloudservice.GetUserGroupsRequest{DisplayName: "Engineering"}).
		Return(&cloudservice.GetUserGroupsResponse{
			Groups: []*identityv1.UserGroup{
				{Id: "group-1", ResourceVersion: "rv-server", Spec: existingSpec},
			},
		}, nil)
	mockPrompter.EXPECT().PromptApply(existingSpec, newSpec, false).Return(nil)
	mockCloud.EXPECT().
		UpdateUserGroup(context.Background(), &cloudservice.UpdateUserGroupRequest{
			GroupId:         "group-1",
			Spec:            newSpec,
			ResourceVersion: "rv-override",
		}).
		Return(&cloudservice.UpdateUserGroupResponse{AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.ApplyUserGroup(context.Background(), temporalcloudcli.ApplyUserGroupParams{
		Spec:            newSpec,
		ResourceVersion: "rv-override",
		Cloud:           mockCloud,
		Prompter:        mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestApplyUserGroup_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	spec := &identityv1.UserGroupSpec{DisplayName: "Engineering"}

	mockCloud.EXPECT().
		GetUserGroups(context.Background(), &cloudservice.GetUserGroupsRequest{DisplayName: "Engineering"}).
		Return(&cloudservice.GetUserGroupsResponse{}, nil)
	mockPrompter.EXPECT().PromptApply((*identityv1.UserGroupSpec)(nil), spec, false).Return(promptErr)

	err := temporalcloudcli.ApplyUserGroup(context.Background(), temporalcloudcli.ApplyUserGroupParams{
		Spec:             spec,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, promptErr)
}

func TestApplyUserGroup_GetUserGroupsError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("api error")

	spec := &identityv1.UserGroupSpec{DisplayName: "Engineering"}

	mockCloud.EXPECT().
		GetUserGroups(context.Background(), &cloudservice.GetUserGroupsRequest{DisplayName: "Engineering"}).
		Return(nil, apiErr)

	err := temporalcloudcli.ApplyUserGroup(context.Background(), temporalcloudcli.ApplyUserGroupParams{
		Spec:             spec,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

func TestCreateCloudGroup_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-1"}

	expectedSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
		Access: &identityv1.Access{
			AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
			NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
				"my-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
			},
		},
	}

	mockCloud.EXPECT().
		CreateUserGroup(context.Background(), &cloudservice.CreateUserGroupRequest{Spec: expectedSpec}).
		Return(&cloudservice.CreateUserGroupResponse{GroupId: "group-1", AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.CreateCloudGroup(context.Background(), temporalcloudcli.CreateCloudGroupParams{
		DisplayName:   "Engineering",
		AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
		NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
			"my-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
		},
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestCreateCloudGroup_AsyncOperationID(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "my-op-id"}

	expectedSpec := &identityv1.UserGroupSpec{
		DisplayName: "Platform",
		GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
	}

	mockCloud.EXPECT().
		CreateUserGroup(context.Background(), &cloudservice.CreateUserGroupRequest{
			Spec:             expectedSpec,
			AsyncOperationId: "my-op-id",
		}).
		Return(&cloudservice.CreateUserGroupResponse{GroupId: "group-2", AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.CreateCloudGroup(context.Background(), temporalcloudcli.CreateCloudGroupParams{
		DisplayName:      "Platform",
		AsyncOperationID: "my-op-id",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestCreateCloudGroup_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("api error")

	expectedSpec := &identityv1.UserGroupSpec{
		DisplayName: "Engineering",
		GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
	}

	mockCloud.EXPECT().
		CreateUserGroup(context.Background(), &cloudservice.CreateUserGroupRequest{Spec: expectedSpec}).
		Return(nil, apiErr)
	mockHandler.EXPECT().HandleErr(apiErr).Return(apiErr)

	err := temporalcloudcli.CreateCloudGroup(context.Background(), temporalcloudcli.CreateCloudGroupParams{
		DisplayName:      "Engineering",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

func TestCreateGoogleGroup_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-3"}

	expectedSpec := &identityv1.UserGroupSpec{
		DisplayName: "Platform",
		GroupType: &identityv1.UserGroupSpec_GoogleGroup{
			GoogleGroup: &identityv1.GoogleGroupSpec{EmailAddress: "platform@example.com"},
		},
		Access: &identityv1.Access{
			AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_READ},
		},
	}

	mockCloud.EXPECT().
		CreateUserGroup(context.Background(), &cloudservice.CreateUserGroupRequest{Spec: expectedSpec}).
		Return(&cloudservice.CreateUserGroupResponse{GroupId: "group-3", AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.CreateGoogleGroup(context.Background(), temporalcloudcli.CreateGoogleGroupParams{
		DisplayName:      "Platform",
		GoogleGroupEmail: "platform@example.com",
		AccountAccess:    &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_READ},
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestCreateGoogleGroup_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("api error")

	expectedSpec := &identityv1.UserGroupSpec{
		DisplayName: "Platform",
		GroupType: &identityv1.UserGroupSpec_GoogleGroup{
			GoogleGroup: &identityv1.GoogleGroupSpec{EmailAddress: "platform@example.com"},
		},
	}

	mockCloud.EXPECT().
		CreateUserGroup(context.Background(), &cloudservice.CreateUserGroupRequest{Spec: expectedSpec}).
		Return(nil, apiErr)
	mockHandler.EXPECT().HandleErr(apiErr).Return(apiErr)

	err := temporalcloudcli.CreateGoogleGroup(context.Background(), temporalcloudcli.CreateGoogleGroupParams{
		DisplayName:      "Platform",
		GoogleGroupEmail: "platform@example.com",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

func TestCreateSCIMGroup_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-4"}

	expectedSpec := &identityv1.UserGroupSpec{
		DisplayName: "Security",
		GroupType: &identityv1.UserGroupSpec_ScimGroup{
			ScimGroup: &identityv1.SCIMGroupSpec{IdpId: "idp-abc"},
		},
	}

	mockCloud.EXPECT().
		CreateUserGroup(context.Background(), &cloudservice.CreateUserGroupRequest{Spec: expectedSpec}).
		Return(&cloudservice.CreateUserGroupResponse{GroupId: "group-4", AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.CreateSCIMGroup(context.Background(), temporalcloudcli.CreateSCIMGroupParams{
		DisplayName:      "Security",
		ScimIdpId:        "idp-abc",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestCreateSCIMGroup_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("api error")

	expectedSpec := &identityv1.UserGroupSpec{
		DisplayName: "Security",
		GroupType: &identityv1.UserGroupSpec_ScimGroup{
			ScimGroup: &identityv1.SCIMGroupSpec{IdpId: "idp-abc"},
		},
	}

	mockCloud.EXPECT().
		CreateUserGroup(context.Background(), &cloudservice.CreateUserGroupRequest{Spec: expectedSpec}).
		Return(nil, apiErr)
	mockHandler.EXPECT().HandleErr(apiErr).Return(apiErr)

	err := temporalcloudcli.CreateSCIMGroup(context.Background(), temporalcloudcli.CreateSCIMGroupParams{
		DisplayName:      "Security",
		ScimIdpId:        "idp-abc",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

func TestGetUserGroup_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	mockCloud.EXPECT().
		GetUserGroup(context.Background(), &cloudservice.GetUserGroupRequest{GroupId: "group-1"}).
		Return(&cloudservice.GetUserGroupResponse{
			Group: &identityv1.UserGroup{
				Id:   "group-1",
				Spec: &identityv1.UserGroupSpec{DisplayName: "Engineering"},
			},
		}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.GetUserGroup(context.Background(), temporalcloudcli.GetUserGroupParams{
		GroupId: "group-1",
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)

	type groupOutput struct {
		Id string `json:"id"`
	}
	var out groupOutput
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "group-1", out.Id)
}

func TestGetUserGroup_Error(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetUserGroup(context.Background(), &cloudservice.GetUserGroupRequest{GroupId: "group-1"}).
		Return(nil, apiErr)

	var buf bytes.Buffer
	err := temporalcloudcli.GetUserGroup(context.Background(), temporalcloudcli.GetUserGroupParams{
		GroupId: "group-1",
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.ErrorIs(t, err, apiErr)
	assert.Empty(t, buf.String())
}

func TestListUserGroups(t *testing.T) {
	apiErr := errors.New("api error")

	tests := []struct {
		name        string
		params      temporalcloudcli.ListUserGroupsParams
		setup       func(cloud *cloudmock.MockCloudServiceClient)
		wantErr     error
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "success",
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUserGroups(context.Background(), &cloudservice.GetUserGroupsRequest{}).
					Return(&cloudservice.GetUserGroupsResponse{
						Groups: []*identityv1.UserGroup{
							{Id: "group-1", Spec: &identityv1.UserGroupSpec{DisplayName: "Engineering"}},
							{Id: "group-2", Spec: &identityv1.UserGroupSpec{DisplayName: "Platform"}},
						},
					}, nil)
			},
			checkOutput: func(t *testing.T, output string) {
				type listOutput struct {
					Groups []struct {
						Id string `json:"id"`
					} `json:"groups"`
				}
				var out listOutput
				require.NoError(t, json.Unmarshal([]byte(output), &out))
				require.Len(t, out.Groups, 2)
				assert.Equal(t, "group-1", out.Groups[0].Id)
				assert.Equal(t, "group-2", out.Groups[1].Id)
			},
		},
		{
			name: "with_namespace_filter",
			params: temporalcloudcli.ListUserGroupsParams{
				Namespace: "my-namespace.my-account",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUserGroups(context.Background(), &cloudservice.GetUserGroupsRequest{
						Namespace: "my-namespace.my-account",
					}).
					Return(&cloudservice.GetUserGroupsResponse{
						Groups: []*identityv1.UserGroup{
							{Id: "group-1", Spec: &identityv1.UserGroupSpec{DisplayName: "Engineering"}},
						},
					}, nil)
			},
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "group-1")
			},
		},
		{
			name: "with_display_name_filter",
			params: temporalcloudcli.ListUserGroupsParams{
				DisplayName: "Engineering",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUserGroups(context.Background(), &cloudservice.GetUserGroupsRequest{
						DisplayName: "Engineering",
					}).
					Return(&cloudservice.GetUserGroupsResponse{
						Groups: []*identityv1.UserGroup{
							{Id: "group-1", Spec: &identityv1.UserGroupSpec{DisplayName: "Engineering"}},
						},
					}, nil)
			},
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "group-1")
			},
		},
		{
			name: "with_pagination",
			params: temporalcloudcli.ListUserGroupsParams{
				PageSize:  10,
				PageToken: "tok-abc",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUserGroups(context.Background(), &cloudservice.GetUserGroupsRequest{
						PageSize:  10,
						PageToken: "tok-abc",
					}).
					Return(&cloudservice.GetUserGroupsResponse{
						Groups:        []*identityv1.UserGroup{{Id: "group-3", Spec: &identityv1.UserGroupSpec{DisplayName: "Security"}}},
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
			name: "with_google_group_email_address_filter",
			params: temporalcloudcli.ListUserGroupsParams{
				GoogleGroupEmailAddress: "eng@example.com",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUserGroups(context.Background(), &cloudservice.GetUserGroupsRequest{
						GoogleGroup: &cloudservice.GetUserGroupsRequest_GoogleGroupFilter{EmailAddress: "eng@example.com"},
					}).
					Return(&cloudservice.GetUserGroupsResponse{
						Groups: []*identityv1.UserGroup{
							{Id: "group-1", Spec: &identityv1.UserGroupSpec{DisplayName: "Engineering"}},
						},
					}, nil)
			},
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "group-1")
			},
		},
		{
			name: "with_scim_group_idp_id_filter",
			params: temporalcloudcli.ListUserGroupsParams{
				ScimGroupIdpId: "idp-abc123",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUserGroups(context.Background(), &cloudservice.GetUserGroupsRequest{
						ScimGroup: &cloudservice.GetUserGroupsRequest_SCIMGroupFilter{IdpId: "idp-abc123"},
					}).
					Return(&cloudservice.GetUserGroupsResponse{
						Groups: []*identityv1.UserGroup{
							{Id: "group-2", Spec: &identityv1.UserGroupSpec{DisplayName: "SCIM Team"}},
						},
					}, nil)
			},
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "group-2")
			},
		},
		{
			name: "api_error",
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUserGroups(context.Background(), &cloudservice.GetUserGroupsRequest{}).
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

			err := temporalcloudcli.ListUserGroups(context.Background(), params)
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

// ---- ListUserGroupMembers ----

func TestListUserGroupMembers(t *testing.T) {
	apiErr := errors.New("api error")

	tests := []struct {
		name        string
		params      temporalcloudcli.ListUserGroupMembersParams
		setup       func(cloud *cloudmock.MockCloudServiceClient)
		wantErr     error
		checkOutput func(t *testing.T, output string)
	}{
		{
			name: "success",
			params: temporalcloudcli.ListUserGroupMembersParams{
				GroupId: "group-1",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUserGroupMembers(context.Background(), &cloudservice.GetUserGroupMembersRequest{GroupId: "group-1"}).
					Return(&cloudservice.GetUserGroupMembersResponse{
						Members: []*identityv1.UserGroupMember{
							{MemberId: &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: "user-1"}}},
							{MemberId: &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: "user-2"}}},
						},
					}, nil)
			},
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "user-1")
				assert.Contains(t, output, "user-2")
			},
		},
		{
			name: "with_pagination",
			params: temporalcloudcli.ListUserGroupMembersParams{
				GroupId:   "group-1",
				PageSize:  10,
				PageToken: "tok-abc",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUserGroupMembers(context.Background(), &cloudservice.GetUserGroupMembersRequest{
						GroupId:   "group-1",
						PageSize:  10,
						PageToken: "tok-abc",
					}).
					Return(&cloudservice.GetUserGroupMembersResponse{
						Members:       []*identityv1.UserGroupMember{{MemberId: &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: "user-3"}}}},
						NextPageToken: "tok-def",
					}, nil)
			},
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "tok-def")
			},
		},
		{
			name: "api_error",
			params: temporalcloudcli.ListUserGroupMembersParams{
				GroupId: "group-1",
			},
			setup: func(cloud *cloudmock.MockCloudServiceClient) {
				cloud.EXPECT().
					GetUserGroupMembers(context.Background(), &cloudservice.GetUserGroupMembersRequest{GroupId: "group-1"}).
					Return(nil, apiErr)
			},
			wantErr:     apiErr,
			checkOutput: func(t *testing.T, output string) { assert.Empty(t, output) },
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

			err := temporalcloudcli.ListUserGroupMembers(context.Background(), params)
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

// ---- AddUserGroupMember ----

func TestAddUserGroupMember_ByUserId(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-1"}

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{User: &identityv1.User{Id: "user-1"}}, nil)
	mockCloud.EXPECT().
		AddUserGroupMember(context.Background(), &cloudservice.AddUserGroupMemberRequest{
			GroupId:  "group-1",
			MemberId: &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: "user-1"}},
		}).
		Return(&cloudservice.AddUserGroupMemberResponse{AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.AddUserGroupMember(context.Background(), temporalcloudcli.AddUserGroupMemberParams{
		GroupId:            "group-1",
		UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
		Cloud:              mockCloud,
		OperationHandler:   mockHandler,
	})
	require.NoError(t, err)
}

func TestAddUserGroupMember_ByUserEmail(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-1"}

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
		Return(&cloudservice.GetUsersResponse{Users: []*identityv1.User{{Id: "user-1"}}}, nil)
	mockCloud.EXPECT().
		AddUserGroupMember(context.Background(), &cloudservice.AddUserGroupMemberRequest{
			GroupId:  "group-1",
			MemberId: &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: "user-1"}},
		}).
		Return(&cloudservice.AddUserGroupMemberResponse{AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.AddUserGroupMember(context.Background(), temporalcloudcli.AddUserGroupMemberParams{
		GroupId:            "group-1",
		UserIdentification: temporalcloudcli.UserIdentificationOptions{UserEmail: "alice@example.com"},
		Cloud:              mockCloud,
		OperationHandler:   mockHandler,
	})
	require.NoError(t, err)
}

func TestAddUserGroupMember_MissingIdentification(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	err := temporalcloudcli.AddUserGroupMember(context.Background(), temporalcloudcli.AddUserGroupMemberParams{
		GroupId:          "group-1",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.ErrorContains(t, err, "must provide either --user-id or --user-email")
}

func TestAddUserGroupMember_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{User: &identityv1.User{Id: "user-1"}}, nil)
	mockCloud.EXPECT().
		AddUserGroupMember(context.Background(), &cloudservice.AddUserGroupMemberRequest{
			GroupId:  "group-1",
			MemberId: &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: "user-1"}},
		}).
		Return(nil, apiErr)
	mockHandler.EXPECT().HandleErr(apiErr).Return(apiErr)

	err := temporalcloudcli.AddUserGroupMember(context.Background(), temporalcloudcli.AddUserGroupMemberParams{
		GroupId:            "group-1",
		UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
		Cloud:              mockCloud,
		OperationHandler:   mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

// ---- RemoveUserGroupMember ----

func TestRemoveUserGroupMember_ByUserId(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-1"}

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{User: &identityv1.User{Id: "user-1"}}, nil)
	mockCloud.EXPECT().
		RemoveUserGroupMember(context.Background(), &cloudservice.RemoveUserGroupMemberRequest{
			GroupId:  "group-1",
			MemberId: &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: "user-1"}},
		}).
		Return(&cloudservice.RemoveUserGroupMemberResponse{AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.RemoveUserGroupMember(context.Background(), temporalcloudcli.RemoveUserGroupMemberParams{
		GroupId:            "group-1",
		UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
		Cloud:              mockCloud,
		OperationHandler:   mockHandler,
	})
	require.NoError(t, err)
}

func TestRemoveUserGroupMember_ByUserEmail(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	op := &operation.AsyncOperation{Id: "op-1"}

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
		Return(&cloudservice.GetUsersResponse{Users: []*identityv1.User{{Id: "user-1"}}}, nil)
	mockCloud.EXPECT().
		RemoveUserGroupMember(context.Background(), &cloudservice.RemoveUserGroupMemberRequest{
			GroupId:  "group-1",
			MemberId: &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: "user-1"}},
		}).
		Return(&cloudservice.RemoveUserGroupMemberResponse{AsyncOperation: op}, nil)
	mockHandler.EXPECT().Handle(op).Return(nil)

	err := temporalcloudcli.RemoveUserGroupMember(context.Background(), temporalcloudcli.RemoveUserGroupMemberParams{
		GroupId:            "group-1",
		UserIdentification: temporalcloudcli.UserIdentificationOptions{UserEmail: "alice@example.com"},
		Cloud:              mockCloud,
		OperationHandler:   mockHandler,
	})
	require.NoError(t, err)
}

func TestRemoveUserGroupMember_MissingIdentification(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	err := temporalcloudcli.RemoveUserGroupMember(context.Background(), temporalcloudcli.RemoveUserGroupMemberParams{
		GroupId:          "group-1",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.ErrorContains(t, err, "must provide either --user-id or --user-email")
}

func TestRemoveUserGroupMember_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetUser(context.Background(), &cloudservice.GetUserRequest{UserId: "user-1"}).
		Return(&cloudservice.GetUserResponse{User: &identityv1.User{Id: "user-1"}}, nil)
	mockCloud.EXPECT().
		RemoveUserGroupMember(context.Background(), &cloudservice.RemoveUserGroupMemberRequest{
			GroupId:  "group-1",
			MemberId: &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: "user-1"}},
		}).
		Return(nil, apiErr)
	mockHandler.EXPECT().HandleErr(apiErr).Return(apiErr)

	err := temporalcloudcli.RemoveUserGroupMember(context.Background(), temporalcloudcli.RemoveUserGroupMemberParams{
		GroupId:            "group-1",
		UserIdentification: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
		Cloud:              mockCloud,
		OperationHandler:   mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}
