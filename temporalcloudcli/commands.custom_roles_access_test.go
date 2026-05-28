package temporalcloudcli_test

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

// AIDEV-NOTE: tests in this file rely on Flags().Changed("custom-role") inside
// run() — that requires the flag to be both registered AND Set() on a real
// pflag.FlagSet. The bindCustomRoleFlags helpers register against the cmd's
// own FlagSet and bind to the embedded CustomRoleOptions struct so the run
// method reads back what the test set.

func bindCustomRoleFlags(t *testing.T, fs *pflag.FlagSet, opts *temporalcloudcli.CustomRoleOptions) {
	t.Helper()
	// Production CustomRoleOptions.BuildFlags uses StringArrayVar (no comma
	// splitting). Mirror that here so tests exercise the real parsing
	// behavior — StringSliceVar would split "a,b" into ["a","b"] and mask
	// any bug involving role IDs that contain commas.
	fs.StringArrayVar(&opts.CustomRole, "custom-role", nil, "")
}

// bindCustomRoleFlag registers --custom-role bound to a plain []string field
// (used by commands that don't embed CustomRoleOptions, e.g. create/invite).
func bindCustomRoleFlag(t *testing.T, fs *pflag.FlagSet, dst *[]string) {
	t.Helper()
	fs.StringArrayVar(dst, "custom-role", nil, "")
}

func bindAccountRoleFlag(t *testing.T, fs *pflag.FlagSet, dst *string) {
	t.Helper()
	fs.StringVar(dst, "account-role", "", "")
}

func setCustomRoles(t *testing.T, fs *pflag.FlagSet, roles ...string) {
	t.Helper()
	for _, r := range roles {
		require.NoError(t, fs.Set("custom-role", r))
	}
}

func userWithRoles(role identityv1.AccountAccess_Role, customRoles ...string) *identityv1.User {
	return &identityv1.User{
		Id:              "user-1",
		ResourceVersion: "rv-1",
		Spec: &identityv1.UserSpec{
			Email: "alice@example.com",
			Access: &identityv1.Access{
				AccountAccess: &identityv1.AccountAccess{
					Role:        role,
					CustomRoles: customRoles,
				},
			},
		},
	}
}

func groupWithRoles(role identityv1.AccountAccess_Role, customRoles ...string) *identityv1.UserGroup {
	return &identityv1.UserGroup{
		Id:              "group-1",
		ResourceVersion: "rv-1",
		Spec: &identityv1.UserGroupSpec{
			DisplayName: "Engineering",
			GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
			Access: &identityv1.Access{
				AccountAccess: &identityv1.AccountAccess{
					Role:        role,
					CustomRoles: customRoles,
				},
			},
		},
	}
}

func saWithRoles(role identityv1.AccountAccess_Role, customRoles ...string) *identityv1.ServiceAccount {
	return &identityv1.ServiceAccount{
		Id:              "sa-1",
		ResourceVersion: "rv-1",
		Spec: &identityv1.ServiceAccountSpec{
			Name: "my-sa",
			Access: &identityv1.Access{
				AccountAccess: &identityv1.AccountAccess{
					Role:        role,
					CustomRoles: customRoles,
				},
			},
		},
	}
}

func saNamespaceScoped() *identityv1.ServiceAccount {
	return &identityv1.ServiceAccount{
		Id:              "sa-ns-1",
		ResourceVersion: "rv-ns-1",
		Spec: &identityv1.ServiceAccountSpec{
			Name: "my-ns-sa",
			NamespaceScopedAccess: &identityv1.NamespaceScopedAccess{
				Namespace: "my-ns.acct",
				Access:    &identityv1.NamespaceAccess{Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
			},
		},
	}
}

func matchUpdateUserCR(role identityv1.AccountAccess_Role, customRoles []string) any {
	return mock.MatchedBy(func(r *cloudservice.UpdateUserRequest) bool {
		acc := r.Spec.GetAccess().GetAccountAccess()
		return r.UserId == "user-1" && acc.GetRole() == role && stringSlicesEqual(acc.CustomRoles, customRoles)
	})
}

func matchUpdateGroupCR(role identityv1.AccountAccess_Role, customRoles []string) any {
	return mock.MatchedBy(func(r *cloudservice.UpdateUserGroupRequest) bool {
		acc := r.Spec.GetAccess().GetAccountAccess()
		return r.GroupId == "group-1" && acc.GetRole() == role && stringSlicesEqual(acc.CustomRoles, customRoles)
	})
}

func matchUpdateSACR(role identityv1.AccountAccess_Role, customRoles []string) any {
	return mock.MatchedBy(func(r *cloudservice.UpdateServiceAccountRequest) bool {
		acc := r.Spec.GetAccess().GetAccountAccess()
		return r.ServiceAccountId == "sa-1" && acc.GetRole() == role && stringSlicesEqual(acc.CustomRoles, customRoles)
	})
}

func matchCreateUserCR(role identityv1.AccountAccess_Role, customRoles []string) any {
	return mock.MatchedBy(func(r *cloudservice.CreateUserRequest) bool {
		acc := r.Spec.GetAccess().GetAccountAccess()
		return acc.GetRole() == role && stringSlicesEqual(acc.CustomRoles, customRoles)
	})
}

func matchCreateSACR(role identityv1.AccountAccess_Role, customRoles []string) any {
	return mock.MatchedBy(func(r *cloudservice.CreateServiceAccountRequest) bool {
		acc := r.Spec.GetAccess().GetAccountAccess()
		return acc.GetRole() == role && stringSlicesEqual(acc.CustomRoles, customRoles)
	})
}

func matchCreateUserGroupCR(role identityv1.AccountAccess_Role, customRoles []string) any {
	return mock.MatchedBy(func(r *cloudservice.CreateUserGroupRequest) bool {
		acc := r.Spec.GetAccess().GetAccountAccess()
		return acc.GetRole() == role && stringSlicesEqual(acc.CustomRoles, customRoles)
	})
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ---- User: set-custom-roles ----

func TestUserSetCustomRoles(t *testing.T) {
	role := identityv1.AccountAccess_ROLE_DEVELOPER
	type tc struct {
		name            string
		setup           func(*temporalcloudcli.CloudUserSetCustomRolesCommand)
		existing        *identityv1.User
		expectGetUser   bool
		expectUpdate    any
		promptOpts      temporalcloudcli.TestPromptOptions
		asyncPollerOpts temporalcloudcli.TestAsyncPollerOptions
		expectedErr     string
	}
	tests := []tc{
		{
			name: "Replace",
			setup: func(cmd *temporalcloudcli.CloudUserSetCustomRolesCommand) {
				bindCustomRoleFlags(t, cmd.Command.Flags(), &cmd.CustomRoleOptions)
				setCustomRoles(t, cmd.Command.Flags(), "role-a", "role-b")
			},
			existing:        userWithRoles(role, "old"),
			expectGetUser:   true,
			expectUpdate:    matchUpdateUserCR(role, []string{"role-a", "role-b"}),
			promptOpts:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOpts: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op"},
		},
		{
			name: "EmptyClears",
			setup: func(cmd *temporalcloudcli.CloudUserSetCustomRolesCommand) {
				bindCustomRoleFlags(t, cmd.Command.Flags(), &cmd.CustomRoleOptions)
			},
			existing:        userWithRoles(role, "old-1", "old-2"),
			expectGetUser:   true,
			expectUpdate:    matchUpdateUserCR(role, nil),
			promptOpts:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOpts: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op"},
		},
		{
			name: "Dedupes",
			setup: func(cmd *temporalcloudcli.CloudUserSetCustomRolesCommand) {
				bindCustomRoleFlags(t, cmd.Command.Flags(), &cmd.CustomRoleOptions)
				setCustomRoles(t, cmd.Command.Flags(), "a", "b", "a")
			},
			existing:        userWithRoles(role),
			expectGetUser:   true,
			expectUpdate:    matchUpdateUserCR(role, []string{"a", "b"}),
			promptOpts:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOpts: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op"},
		},
		{
			name: "NoAccountAccess",
			setup: func(cmd *temporalcloudcli.CloudUserSetCustomRolesCommand) {
				bindCustomRoleFlags(t, cmd.Command.Flags(), &cmd.CustomRoleOptions)
				setCustomRoles(t, cmd.Command.Flags(), "a")
			},
			existing:      &identityv1.User{Id: "user-1", Spec: &identityv1.UserSpec{Email: "a@b.com"}},
			expectGetUser: true,
			expectedErr:   "user has no account access",
		},
		{
			name: "PromptDeclined",
			setup: func(cmd *temporalcloudcli.CloudUserSetCustomRolesCommand) {
				bindCustomRoleFlags(t, cmd.Command.Flags(), &cmd.CustomRoleOptions)
				setCustomRoles(t, cmd.Command.Flags(), "a")
			},
			existing:      userWithRoles(role),
			expectGetUser: true,
			promptOpts:    temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting set.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &temporalcloudcli.CloudUserSetCustomRolesCommand{
				UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
			}
			tt.setup(cmd)
			temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					if tt.expectGetUser {
						c.EXPECT().
							GetUser(mock.Anything, &cloudservice.GetUserRequest{UserId: "user-1"}, mock.Anything).
							Return(&cloudservice.GetUserResponse{User: tt.existing}, nil)
					}
					if tt.expectUpdate != nil {
						c.EXPECT().
							UpdateUser(mock.Anything, tt.expectUpdate, mock.Anything).
							Return(&cloudservice.UpdateUserResponse{
								AsyncOperation: &operation.AsyncOperation{Id: "op"},
							}, nil)
					}
				},
				PromptOptions:      tt.promptOpts,
				AsyncPollerOptions: tt.asyncPollerOpts,
				JSONOutput:         true,
				ExpectedError:      tt.expectedErr,
			})
		})
	}
}

// ---- User-group: set-custom-roles ----

func TestUserGroupSetCustomRoles(t *testing.T) {
	role := identityv1.AccountAccess_ROLE_DEVELOPER
	type tc struct {
		name            string
		setup           func(*temporalcloudcli.CloudUserGroupSetCustomRolesCommand)
		existing        *identityv1.UserGroup
		expectGetGroup  bool
		expectUpdate    any
		promptOpts      temporalcloudcli.TestPromptOptions
		asyncPollerOpts temporalcloudcli.TestAsyncPollerOptions
		expectedErr     string
	}
	tests := []tc{
		{
			name: "Replace",
			setup: func(cmd *temporalcloudcli.CloudUserGroupSetCustomRolesCommand) {
				bindCustomRoleFlags(t, cmd.Command.Flags(), &cmd.CustomRoleOptions)
				setCustomRoles(t, cmd.Command.Flags(), "role-a")
			},
			existing:        groupWithRoles(role, "old"),
			expectGetGroup:  true,
			expectUpdate:    matchUpdateGroupCR(role, []string{"role-a"}),
			promptOpts:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOpts: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op"},
		},
		{
			name: "EmptyClears",
			setup: func(cmd *temporalcloudcli.CloudUserGroupSetCustomRolesCommand) {
				bindCustomRoleFlags(t, cmd.Command.Flags(), &cmd.CustomRoleOptions)
			},
			existing:        groupWithRoles(role, "old"),
			expectGetGroup:  true,
			expectUpdate:    matchUpdateGroupCR(role, nil),
			promptOpts:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOpts: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op"},
		},
		{
			name: "NoAccountAccess",
			setup: func(cmd *temporalcloudcli.CloudUserGroupSetCustomRolesCommand) {
				bindCustomRoleFlags(t, cmd.Command.Flags(), &cmd.CustomRoleOptions)
				setCustomRoles(t, cmd.Command.Flags(), "a")
			},
			existing:       &identityv1.UserGroup{Id: "group-1", Spec: &identityv1.UserGroupSpec{DisplayName: "X"}},
			expectGetGroup: true,
			expectedErr:    "group has no account access",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &temporalcloudcli.CloudUserGroupSetCustomRolesCommand{
				GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"},
			}
			tt.setup(cmd)
			temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					if tt.expectGetGroup {
						c.EXPECT().
							GetUserGroup(mock.Anything, &cloudservice.GetUserGroupRequest{GroupId: "group-1"}, mock.Anything).
							Return(&cloudservice.GetUserGroupResponse{Group: tt.existing}, nil)
					}
					if tt.expectUpdate != nil {
						c.EXPECT().
							UpdateUserGroup(mock.Anything, tt.expectUpdate, mock.Anything).
							Return(&cloudservice.UpdateUserGroupResponse{
								AsyncOperation: &operation.AsyncOperation{Id: "op"},
							}, nil)
					}
				},
				PromptOptions:      tt.promptOpts,
				AsyncPollerOptions: tt.asyncPollerOpts,
				JSONOutput:         true,
				ExpectedError:      tt.expectedErr,
			})
		})
	}
}

// ---- Service account: set-custom-roles ----

func TestServiceAccountSetCustomRoles(t *testing.T) {
	role := identityv1.AccountAccess_ROLE_DEVELOPER
	type tc struct {
		name            string
		setup           func(*temporalcloudcli.CloudServiceAccountSetCustomRolesCommand)
		serviceAccount  *identityv1.ServiceAccount
		expectGetSA     bool
		expectUpdate    any
		promptOpts      temporalcloudcli.TestPromptOptions
		asyncPollerOpts temporalcloudcli.TestAsyncPollerOptions
		expectedErr     string
	}
	tests := []tc{
		{
			name: "Replace",
			setup: func(cmd *temporalcloudcli.CloudServiceAccountSetCustomRolesCommand) {
				bindCustomRoleFlags(t, cmd.Command.Flags(), &cmd.CustomRoleOptions)
				setCustomRoles(t, cmd.Command.Flags(), "role-a")
			},
			serviceAccount:  saWithRoles(role, "old"),
			expectGetSA:     true,
			expectUpdate:    matchUpdateSACR(role, []string{"role-a"}),
			promptOpts:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOpts: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op"},
		},
		{
			name: "NamespaceScopedSAError",
			setup: func(cmd *temporalcloudcli.CloudServiceAccountSetCustomRolesCommand) {
				bindCustomRoleFlags(t, cmd.Command.Flags(), &cmd.CustomRoleOptions)
				setCustomRoles(t, cmd.Command.Flags(), "role-a")
			},
			serviceAccount: saNamespaceScoped(),
			expectGetSA:    true,
			expectedErr:    "not valid for namespace-scoped service accounts",
		},
		{
			name: "EmptyClears",
			setup: func(cmd *temporalcloudcli.CloudServiceAccountSetCustomRolesCommand) {
				bindCustomRoleFlags(t, cmd.Command.Flags(), &cmd.CustomRoleOptions)
			},
			serviceAccount:  saWithRoles(role, "old"),
			expectGetSA:     true,
			expectUpdate:    matchUpdateSACR(role, nil),
			promptOpts:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOpts: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &temporalcloudcli.CloudServiceAccountSetCustomRolesCommand{
				ServiceAccountId: "sa-1",
			}
			tt.setup(cmd)
			temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					if tt.expectGetSA {
						c.EXPECT().
							GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
							Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: tt.serviceAccount}, nil)
					}
					if tt.expectUpdate != nil {
						c.EXPECT().
							UpdateServiceAccount(mock.Anything, tt.expectUpdate, mock.Anything).
							Return(&cloudservice.UpdateServiceAccountResponse{
								AsyncOperation: &operation.AsyncOperation{Id: "op"},
							}, nil)
					}
				},
				PromptOptions:      tt.promptOpts,
				AsyncPollerOptions: tt.asyncPollerOpts,
				JSONOutput:         true,
				ExpectedError:      tt.expectedErr,
			})
		})
	}
}

// ---- Preserve CustomRoles when changing the built-in role ----

func TestSetAccountRolePreservesCustomRoles_User(t *testing.T) {
	existing := userWithRoles(identityv1.AccountAccess_ROLE_DEVELOPER, "preserved")
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudUserSetAccountRoleCommand{
		UserIdentificationOptions: temporalcloudcli.UserIdentificationOptions{UserId: "user-1"},
		AccountRole:               "admin",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetUser(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetUserResponse{User: existing}, nil)
			c.EXPECT().
				UpdateUser(mock.Anything,
					matchUpdateUserCR(identityv1.AccountAccess_ROLE_ADMIN, []string{"preserved"}),
					mock.Anything).
				Return(&cloudservice.UpdateUserResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op"},
		JSONOutput:         true,
	})
}

func TestSetAccountRolePreservesCustomRoles_UserGroup(t *testing.T) {
	existing := groupWithRoles(identityv1.AccountAccess_ROLE_DEVELOPER, "preserved")
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudUserGroupSetAccountRoleCommand{
		GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"},
		AccountRole:    "admin",
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetUserGroup(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetUserGroupResponse{Group: existing}, nil)
			c.EXPECT().
				UpdateUserGroup(mock.Anything,
					matchUpdateGroupCR(identityv1.AccountAccess_ROLE_ADMIN, []string{"preserved"}),
					mock.Anything).
				Return(&cloudservice.UpdateUserGroupResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op"},
		JSONOutput:         true,
	})
}

func TestServiceAccountUpdateAccountRolePreservesCustomRoles(t *testing.T) {
	existing := saWithRoles(identityv1.AccountAccess_ROLE_DEVELOPER, "preserved")
	cmd := &temporalcloudcli.CloudServiceAccountUpdateCommand{
		ServiceAccountId: "sa-1",
	}
	bindAccountRoleFlag(t, cmd.Command.Flags(), &cmd.AccountRole)
	require.NoError(t, cmd.Command.Flags().Set("account-role", "admin"))
	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: existing}, nil)
			c.EXPECT().
				UpdateServiceAccount(mock.Anything,
					matchUpdateSACR(identityv1.AccountAccess_ROLE_ADMIN, []string{"preserved"}),
					mock.Anything).
				Return(&cloudservice.UpdateServiceAccountResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op"},
		JSONOutput:         true,
	})
}

// ---- Create / invite with --custom-role ----

func TestUserInviteWithCustomRoles_Success(t *testing.T) {
	cmd := &temporalcloudcli.CloudUserInviteCommand{
		Email:       "alice@example.com",
		AccountRole: "developer",
	}
	bindCustomRoleFlag(t, cmd.Command.Flags(), &cmd.CustomRole)
	setCustomRoles(t, cmd.Command.Flags(), "r1")
	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				CreateUser(mock.Anything,
					matchCreateUserCR(identityv1.AccountAccess_ROLE_DEVELOPER, []string{"r1"}),
					mock.Anything).
				Return(&cloudservice.CreateUserResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op"},
		JSONOutput:         true,
	})
}

func TestUserInviteWithCustomRoles_NoAccountRoleError(t *testing.T) {
	cmd := &temporalcloudcli.CloudUserInviteCommand{Email: "alice@example.com"}
	bindCustomRoleFlag(t, cmd.Command.Flags(), &cmd.CustomRole)
	setCustomRoles(t, cmd.Command.Flags(), "r1")
	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		JSONOutput:    true,
		ExpectedError: "--custom-role requires --account-role",
	})
}

func TestServiceAccountCreateWithCustomRoles_Success(t *testing.T) {
	cmd := &temporalcloudcli.CloudServiceAccountCreateCommand{
		Name:        "my-sa",
		AccountRole: "developer",
	}
	bindCustomRoleFlag(t, cmd.Command.Flags(), &cmd.CustomRole)
	setCustomRoles(t, cmd.Command.Flags(), "r1", "r2")
	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				CreateServiceAccount(mock.Anything,
					matchCreateSACR(identityv1.AccountAccess_ROLE_DEVELOPER, []string{"r1", "r2"}),
					mock.Anything).
				Return(&cloudservice.CreateServiceAccountResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op"},
		JSONOutput:         true,
	})
}

func TestServiceAccountCreateWithCustomRoles_NoAccountRoleError(t *testing.T) {
	cmd := &temporalcloudcli.CloudServiceAccountCreateCommand{Name: "my-sa"}
	bindCustomRoleFlag(t, cmd.Command.Flags(), &cmd.CustomRole)
	setCustomRoles(t, cmd.Command.Flags(), "r1")
	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		JSONOutput:    true,
		ExpectedError: "--custom-role requires --account-role",
	})
}

func TestUserGroupCreateCloudGroupWithCustomRoles_Success(t *testing.T) {
	cmd := &temporalcloudcli.CloudUserGroupCreateCloudGroupCommand{
		DisplayName: "Engineering",
		AccountRole: "developer",
	}
	bindCustomRoleFlag(t, cmd.Command.Flags(), &cmd.CustomRole)
	setCustomRoles(t, cmd.Command.Flags(), "r1")
	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				CreateUserGroup(mock.Anything,
					matchCreateUserGroupCR(identityv1.AccountAccess_ROLE_DEVELOPER, []string{"r1"}),
					mock.Anything).
				Return(&cloudservice.CreateUserGroupResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op"},
		JSONOutput:         true,
	})
}

func TestUserGroupCreateCloudGroupWithCustomRoles_NoAccountRoleError(t *testing.T) {
	cmd := &temporalcloudcli.CloudUserGroupCreateCloudGroupCommand{DisplayName: "Engineering"}
	bindCustomRoleFlag(t, cmd.Command.Flags(), &cmd.CustomRole)
	setCustomRoles(t, cmd.Command.Flags(), "r1")
	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		JSONOutput:    true,
		ExpectedError: "--custom-role requires --account-role",
	})
}

func TestUserGroupUpdateWithCustomRoles_Replace(t *testing.T) {
	existing := groupWithRoles(identityv1.AccountAccess_ROLE_DEVELOPER, "old")
	cmd := &temporalcloudcli.CloudUserGroupUpdateCommand{
		GroupIdOptions: temporalcloudcli.GroupIdOptions{GroupId: "group-1"},
	}
	bindCustomRoleFlags(t, cmd.Command.Flags(), &cmd.CustomRoleOptions)
	setCustomRoles(t, cmd.Command.Flags(), "new")
	temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetUserGroup(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetUserGroupResponse{Group: existing}, nil)
			c.EXPECT().
				UpdateUserGroup(mock.Anything,
					matchUpdateGroupCR(identityv1.AccountAccess_ROLE_DEVELOPER, []string{"new"}),
					mock.Anything).
				Return(&cloudservice.UpdateUserGroupResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op"},
				}, nil)
		},
		PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op"},
		JSONOutput:         true,
	})
}

