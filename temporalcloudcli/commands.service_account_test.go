package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

// --- DeleteServiceAccount ---

func TestDeleteServiceAccount(t *testing.T) {
	testSA := &identityv1.ServiceAccount{
		Id:              "sa-1",
		ResourceVersion: "rv-1",
		Spec:            &identityv1.ServiceAccountSpec{Name: "my-sa"},
	}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudServiceAccountDeleteCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudServiceAccountDeleteCommand{ServiceAccountId: "sa-1"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: testSA}, nil)
				c.EXPECT().
					DeleteServiceAccount(mock.Anything, &cloudservice.DeleteServiceAccountRequest{
						ServiceAccountId: "sa-1",
						ResourceVersion:  "rv-1",
					}, mock.Anything).
					Return(&cloudservice.DeleteServiceAccountResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudServiceAccountDeleteCommand{
				ServiceAccountId:       "sa-1",
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: testSA}, nil)
				c.EXPECT().
					DeleteServiceAccount(mock.Anything, &cloudservice.DeleteServiceAccountRequest{
						ServiceAccountId: "sa-1",
						ResourceVersion:  "rv-override",
					}, mock.Anything).
					Return(&cloudservice.DeleteServiceAccountResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name: "GetServiceAccountError",
			cmd:  temporalcloudcli.CloudServiceAccountDeleteCommand{ServiceAccountId: "sa-1"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
		},
		{
			name: "PromptDeclined",
			cmd:  temporalcloudcli.CloudServiceAccountDeleteCommand{ServiceAccountId: "sa-1"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: testSA}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting delete.",
		},
		{
			name: "DeleteServiceAccountError",
			cmd:  temporalcloudcli.CloudServiceAccountDeleteCommand{ServiceAccountId: "sa-1"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: testSA}, nil)
				c.EXPECT().
					DeleteServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("delete failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{
				ExpectPromptYes:        true,
				ExpectPromptYesMessage: "Delete",
				PromptResult:           true,
			},
			expectedErr: "delete operation failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- CreateServiceAccount ---

func TestCreateServiceAccount(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-create-sa"}

	t.Run("Success", func(t *testing.T) {
		cmd := temporalcloudcli.CloudServiceAccountCreateCommand{
			Name:            "my-sa",
			Description:     "a test SA",
			AccountRole:     "developer",
			NamespaceAccess: []string{"my-ns.acct=write"},
			ProjectAccess:   []string{"project-1=member"},
		}
		temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
			CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateServiceAccount(mock.Anything, &cloudservice.CreateServiceAccountRequest{
						Spec: &identityv1.ServiceAccountSpec{
							Name:        "my-sa",
							Description: "a test SA",
							Access: &identityv1.Access{
								AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
								NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
									"my-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
								},
								ProjectAccesses: map[string]*identityv1.ProjectAccess{
									"project-1": {Role: identityv1.ProjectAccess_PROJECT_ROLE_MEMBER},
								},
							},
						},
					}, mock.Anything).
					Return(&cloudservice.CreateServiceAccountResponse{
						ServiceAccountId: "sa-new",
						AsyncOperation:   op,
					}, nil)
			},
			AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-create-sa"},
			PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
		})
	})

	t.Run("PromptDeclined", func(t *testing.T) {
		cmd := temporalcloudcli.CloudServiceAccountCreateCommand{Name: "my-sa"}
		temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
			PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			ExpectedError: "Aborting create.",
		})
	})

	t.Run("InvalidAccountRole", func(t *testing.T) {
		cmd := temporalcloudcli.CloudServiceAccountCreateCommand{Name: "my-sa", AccountRole: "superadmin"}
		temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
			ExpectedError: `invalid account role "superadmin"`,
		})
	})

	t.Run("InvalidNamespaceAccess", func(t *testing.T) {
		cmd := temporalcloudcli.CloudServiceAccountCreateCommand{Name: "my-sa", NamespaceAccess: []string{"bad-format"}}
		temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
			ExpectedError: `invalid namespace-access "bad-format"`,
		})
	})

	t.Run("InvalidProjectAccess", func(t *testing.T) {
		cmd := temporalcloudcli.CloudServiceAccountCreateCommand{Name: "my-sa", ProjectAccess: []string{"project-1=developer"}}
		temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
			ExpectedError: `invalid project role "developer": must be one of admin, write, read, list, contribute, member in project-access "project-1=developer"`,
		})
	})

	t.Run("APIError", func(t *testing.T) {
		cmd := temporalcloudcli.CloudServiceAccountCreateCommand{Name: "my-sa"}
		temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
			CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("internal error"))
			},
			PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			ExpectedError: "internal error",
		})
	})
}

// --- CreateNamespaceScopedServiceAccount ---

func TestCreateNamespaceScopedServiceAccount(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-create-ns-sa"}

	t.Run("Success", func(t *testing.T) {
		cmd := temporalcloudcli.CloudServiceAccountCreateNamespaceScopedCommand{
			Name:                "my-ns-sa",
			Namespace:           "my-ns.acct",
			NamespacePermission: "read",
		}
		temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
			CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateServiceAccount(mock.Anything, &cloudservice.CreateServiceAccountRequest{
						Spec: &identityv1.ServiceAccountSpec{
							Name: "my-ns-sa",
							NamespaceScopedAccess: &identityv1.NamespaceScopedAccess{
								Namespace: "my-ns.acct",
								Access:    &identityv1.NamespaceAccess{Permission: identityv1.NamespaceAccess_PERMISSION_READ},
							},
						},
					}, mock.Anything).
					Return(&cloudservice.CreateServiceAccountResponse{
						ServiceAccountId: "sa-ns-new",
						AsyncOperation:   op,
					}, nil)
			},
			AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-create-ns-sa"},
			PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
		})
	})

	t.Run("PromptDeclined", func(t *testing.T) {
		cmd := temporalcloudcli.CloudServiceAccountCreateNamespaceScopedCommand{
			Name:                "my-ns-sa",
			Namespace:           "my-ns.acct",
			NamespacePermission: "write",
		}
		temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
			PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			ExpectedError: "Aborting create.",
		})
	})

	t.Run("InvalidPermission", func(t *testing.T) {
		cmd := temporalcloudcli.CloudServiceAccountCreateNamespaceScopedCommand{
			Name:                "my-ns-sa",
			Namespace:           "my-ns.acct",
			NamespacePermission: "superwrite",
		}
		temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
			ExpectedError: `invalid namespace permission "superwrite"`,
		})
	})

	t.Run("APIError", func(t *testing.T) {
		cmd := temporalcloudcli.CloudServiceAccountCreateNamespaceScopedCommand{
			Name:                "my-ns-sa",
			Namespace:           "my-ns.acct",
			NamespacePermission: "admin",
		}
		temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
			CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("quota exceeded"))
			},
			PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			ExpectedError: "quota exceeded",
		})
	})
}

// --- CreateProjectScopedServiceAccount ---

func TestCreateProjectScopedServiceAccount(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-create-project-sa"}

	t.Run("Success", func(t *testing.T) {
		cmd := temporalcloudcli.CloudServiceAccountCreateProjectScopedCommand{
			Name:            "my-project-sa",
			Description:     "a project SA",
			ProjectId:       "project-1",
			ProjectRole:     "write",
			NamespaceAccess: []string{"my-ns.acct=read"},
		}
		temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
			CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateServiceAccount(mock.Anything, &cloudservice.CreateServiceAccountRequest{
						Spec: &identityv1.ServiceAccountSpec{
							Name:        "my-project-sa",
							Description: "a project SA",
							ProjectScopedAccess: &identityv1.ProjectScopedAccess{
								ProjectId: "project-1",
								Access: &identityv1.ProjectAccess{
									Role: identityv1.ProjectAccess_PROJECT_ROLE_WRITE,
								},
								NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
									"my-ns.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
								},
							},
						},
					}, mock.Anything).
					Return(&cloudservice.CreateServiceAccountResponse{
						ServiceAccountId: "sa-project-new",
						AsyncOperation:   op,
					}, nil)
			},
			AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-create-project-sa"},
			PromptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
		})
	})

	t.Run("InvalidProjectRole", func(t *testing.T) {
		cmd := temporalcloudcli.CloudServiceAccountCreateProjectScopedCommand{
			Name:        "my-project-sa",
			ProjectId:   "project-1",
			ProjectRole: "developer",
		}
		temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
			ExpectedError: `invalid project role "developer"`,
		})
	})

	t.Run("InvalidNamespaceAccess", func(t *testing.T) {
		cmd := temporalcloudcli.CloudServiceAccountCreateProjectScopedCommand{
			Name:            "my-project-sa",
			ProjectId:       "project-1",
			ProjectRole:     "write",
			NamespaceAccess: []string{"my-ns.acct=superread"},
		}
		temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
			ExpectedError: `invalid permission "superread" in namespace-access "my-ns.acct=superread"`,
		})
	})

	t.Run("PromptDeclined", func(t *testing.T) {
		cmd := temporalcloudcli.CloudServiceAccountCreateProjectScopedCommand{
			Name:        "my-project-sa",
			ProjectId:   "project-1",
			ProjectRole: "write",
		}
		temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
			PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			ExpectedError: "Aborting create.",
		})
	})
}

// --- UpdateServiceAccount ---
//
// AIDEV-NOTE: Tests use a setupCmd func to manually register flags on the cobra FlagSet and call
// Flags().Set() so that Flags().Changed() returns true inside run(). The TestCommand harness calls
// run() directly (bypassing cobra flag parsing), so this is the only way to simulate explicit flags.

func TestUpdateServiceAccount(t *testing.T) {
	op := &operation.AsyncOperation{Id: "op-update-sa"}

	accountScopedSA := &identityv1.ServiceAccount{
		Id:              "sa-1",
		ResourceVersion: "rv-1",
		Spec: &identityv1.ServiceAccountSpec{
			Name:        "my-sa",
			Description: "original desc",
			Access: &identityv1.Access{
				AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
				NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
					"ns1.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
				},
				ProjectAccesses: map[string]*identityv1.ProjectAccess{
					"project-1": {Role: identityv1.ProjectAccess_PROJECT_ROLE_READ},
				},
			},
		},
	}

	namespaceScopedSA := &identityv1.ServiceAccount{
		Id:              "sa-2",
		ResourceVersion: "rv-2",
		Spec: &identityv1.ServiceAccountSpec{
			Name: "my-ns-sa",
			NamespaceScopedAccess: &identityv1.NamespaceScopedAccess{
				Namespace: "ns1.acct",
				Access:    &identityv1.NamespaceAccess{Permission: identityv1.NamespaceAccess_PERMISSION_READ},
			},
		},
	}

	projectScopedSA := &identityv1.ServiceAccount{
		Id:              "sa-3",
		ResourceVersion: "rv-3",
		Spec: &identityv1.ServiceAccountSpec{
			Name: "my-project-sa",
			ProjectScopedAccess: &identityv1.ProjectScopedAccess{
				ProjectId: "project-1",
				Access:    &identityv1.ProjectAccess{Role: identityv1.ProjectAccess_PROJECT_ROLE_READ},
				NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
					"ns1.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
				},
			},
		},
	}

	tests := []struct {
		name                    string
		setupCmd                func(*temporalcloudcli.CloudServiceAccountUpdateCommand)
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "UpdateAccountScopedRole",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.Command.Flags().StringVar(&cmd.AccountRole, "account-role", "", "")
				require.NoError(t, cmd.Command.Flags().Set("account-role", "admin"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: accountScopedSA}, nil)
				c.EXPECT().
					UpdateServiceAccount(mock.Anything, &cloudservice.UpdateServiceAccountRequest{
						ServiceAccountId: "sa-1",
						Spec: &identityv1.ServiceAccountSpec{
							Name:        "my-sa",
							Description: "original desc",
							Access: &identityv1.Access{
								AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN},
								NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
									"ns1.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
								},
								ProjectAccesses: map[string]*identityv1.ProjectAccess{
									"project-1": {Role: identityv1.ProjectAccess_PROJECT_ROLE_READ},
								},
							},
						},
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateServiceAccountResponse{AsyncOperation: op}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-update-sa"},
		},
		{
			name: "UpdateAccountScopedNamespaceAccess",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.Command.Flags().StringArrayVar(&cmd.NamespaceAccess, "namespace-access", nil, "")
				require.NoError(t, cmd.Command.Flags().Set("namespace-access", "ns2.acct=write"))
				require.NoError(t, cmd.Command.Flags().Set("namespace-access", "ns1.acct="))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: accountScopedSA}, nil)
				c.EXPECT().
					UpdateServiceAccount(mock.Anything, &cloudservice.UpdateServiceAccountRequest{
						ServiceAccountId: "sa-1",
						Spec: &identityv1.ServiceAccountSpec{
							Name:        "my-sa",
							Description: "original desc",
							Access: &identityv1.Access{
								AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
								NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
									"ns2.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
								},
								ProjectAccesses: map[string]*identityv1.ProjectAccess{
									"project-1": {Role: identityv1.ProjectAccess_PROJECT_ROLE_READ},
								},
							},
						},
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateServiceAccountResponse{AsyncOperation: op}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-update-sa"},
		},
		{
			name: "UpdateAccountScopedProjectAccess",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.Command.Flags().StringArrayVar(&cmd.ProjectAccess, "project-access", nil, "")
				require.NoError(t, cmd.Command.Flags().Set("project-access", "project-2=write"))
				require.NoError(t, cmd.Command.Flags().Set("project-access", "project-1="))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: accountScopedSA}, nil)
				c.EXPECT().
					UpdateServiceAccount(mock.Anything, &cloudservice.UpdateServiceAccountRequest{
						ServiceAccountId: "sa-1",
						Spec: &identityv1.ServiceAccountSpec{
							Name:        "my-sa",
							Description: "original desc",
							Access: &identityv1.Access{
								AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
								NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
									"ns1.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
								},
								ProjectAccesses: map[string]*identityv1.ProjectAccess{
									"project-2": {Role: identityv1.ProjectAccess_PROJECT_ROLE_WRITE},
								},
							},
						},
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateServiceAccountResponse{AsyncOperation: op}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-update-sa"},
		},
		{
			name: "UpdateNamespaceScopedPermission",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.ServiceAccountId = "sa-2"
				cmd.Command.Flags().StringVar(&cmd.NamespacePermission, "namespace-permission", "", "")
				require.NoError(t, cmd.Command.Flags().Set("namespace-permission", "admin"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-2"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: namespaceScopedSA}, nil)
				c.EXPECT().
					UpdateServiceAccount(mock.Anything, &cloudservice.UpdateServiceAccountRequest{
						ServiceAccountId: "sa-2",
						Spec: &identityv1.ServiceAccountSpec{
							Name: "my-ns-sa",
							NamespaceScopedAccess: &identityv1.NamespaceScopedAccess{
								Namespace: "ns1.acct",
								Access:    &identityv1.NamespaceAccess{Permission: identityv1.NamespaceAccess_PERMISSION_ADMIN},
							},
						},
						ResourceVersion: "rv-2",
					}, mock.Anything).
					Return(&cloudservice.UpdateServiceAccountResponse{AsyncOperation: op}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-update-sa"},
		},
		{
			name: "UpdateName",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.Command.Flags().StringVar(&cmd.Name, "name", "", "")
				require.NoError(t, cmd.Command.Flags().Set("name", "new-name"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: accountScopedSA}, nil)
				c.EXPECT().
					UpdateServiceAccount(mock.Anything, &cloudservice.UpdateServiceAccountRequest{
						ServiceAccountId: "sa-1",
						Spec: &identityv1.ServiceAccountSpec{
							Name:        "new-name",
							Description: "original desc",
							Access: &identityv1.Access{
								AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
								NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
									"ns1.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
								},
								ProjectAccesses: map[string]*identityv1.ProjectAccess{
									"project-1": {Role: identityv1.ProjectAccess_PROJECT_ROLE_READ},
								},
							},
						},
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateServiceAccountResponse{AsyncOperation: op}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-update-sa"},
		},
		{
			name: "UpdateProjectScopedRole",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.ServiceAccountId = "sa-3"
				cmd.Command.Flags().StringVar(&cmd.ProjectRole, "project-role", "", "")
				require.NoError(t, cmd.Command.Flags().Set("project-role", "write"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-3"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: projectScopedSA}, nil)
				c.EXPECT().
					UpdateServiceAccount(mock.Anything, &cloudservice.UpdateServiceAccountRequest{
						ServiceAccountId: "sa-3",
						Spec: &identityv1.ServiceAccountSpec{
							Name: "my-project-sa",
							ProjectScopedAccess: &identityv1.ProjectScopedAccess{
								ProjectId: "project-1",
								Access:    &identityv1.ProjectAccess{Role: identityv1.ProjectAccess_PROJECT_ROLE_WRITE},
								NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
									"ns1.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_READ},
								},
							},
						},
						ResourceVersion: "rv-3",
					}, mock.Anything).
					Return(&cloudservice.UpdateServiceAccountResponse{AsyncOperation: op}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-update-sa"},
		},
		{
			name: "UpdateProjectScopedNamespaceAccess",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.ServiceAccountId = "sa-3"
				cmd.Command.Flags().StringArrayVar(&cmd.NamespaceAccess, "namespace-access", nil, "")
				require.NoError(t, cmd.Command.Flags().Set("namespace-access", "ns2.acct=write"))
				require.NoError(t, cmd.Command.Flags().Set("namespace-access", "ns1.acct="))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-3"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: projectScopedSA}, nil)
				c.EXPECT().
					UpdateServiceAccount(mock.Anything, &cloudservice.UpdateServiceAccountRequest{
						ServiceAccountId: "sa-3",
						Spec: &identityv1.ServiceAccountSpec{
							Name: "my-project-sa",
							ProjectScopedAccess: &identityv1.ProjectScopedAccess{
								ProjectId: "project-1",
								Access:    &identityv1.ProjectAccess{Role: identityv1.ProjectAccess_PROJECT_ROLE_READ},
								NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
									"ns2.acct": {Permission: identityv1.NamespaceAccess_PERMISSION_WRITE},
								},
							},
						},
						ResourceVersion: "rv-3",
					}, mock.Anything).
					Return(&cloudservice.UpdateServiceAccountResponse{AsyncOperation: op}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-update-sa"},
		},
		{
			name: "PromptDeclined",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.Command.Flags().StringVar(&cmd.Name, "name", "", "")
				require.NoError(t, cmd.Command.Flags().Set("name", "new-name"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: accountScopedSA}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting update.",
		},
		{
			name: "AccountRoleOnNamespaceScopedSA",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.ServiceAccountId = "sa-2"
				cmd.Command.Flags().StringVar(&cmd.AccountRole, "account-role", "", "")
				require.NoError(t, cmd.Command.Flags().Set("account-role", "admin"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: namespaceScopedSA}, nil)
			},
			expectedErr: "--account-role, --namespace-access, --project-access, --project-role, and --custom-role are not valid for namespace-scoped service accounts",
		},
		{
			name: "AccountRoleOnProjectScopedSA",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.ServiceAccountId = "sa-3"
				cmd.Command.Flags().StringVar(&cmd.AccountRole, "account-role", "", "")
				require.NoError(t, cmd.Command.Flags().Set("account-role", "admin"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: projectScopedSA}, nil)
			},
			expectedErr: "--account-role, --project-access, --namespace-permission, and --custom-role are not valid for project-scoped service accounts",
		},
		{
			name: "NamespacePermissionOnAccountScopedSA",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.Command.Flags().StringVar(&cmd.NamespacePermission, "namespace-permission", "", "")
				require.NoError(t, cmd.Command.Flags().Set("namespace-permission", "write"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: accountScopedSA}, nil)
			},
			expectedErr: "--namespace-permission is only valid for namespace-scoped service accounts",
		},
		{
			name: "ProjectRoleOnAccountScopedSA",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.Command.Flags().StringVar(&cmd.ProjectRole, "project-role", "", "")
				require.NoError(t, cmd.Command.Flags().Set("project-role", "write"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: accountScopedSA}, nil)
			},
			expectedErr: "--project-role is only valid for project-scoped service accounts",
		},
		{
			name: "ProjectAccessOnProjectScopedSA",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.ServiceAccountId = "sa-3"
				cmd.Command.Flags().StringArrayVar(&cmd.ProjectAccess, "project-access", nil, "")
				require.NoError(t, cmd.Command.Flags().Set("project-access", "project-2=write"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: projectScopedSA}, nil)
			},
			expectedErr: "--account-role, --project-access, --namespace-permission, and --custom-role are not valid for project-scoped service accounts",
		},
		{
			name: "InvalidAccountRole",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.Command.Flags().StringVar(&cmd.AccountRole, "account-role", "", "")
				require.NoError(t, cmd.Command.Flags().Set("account-role", "superadmin"))
			},
			expectedErr: `invalid account role "superadmin"`,
		},
		{
			name: "InvalidNamespacePermission",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.Command.Flags().StringVar(&cmd.NamespacePermission, "namespace-permission", "", "")
				require.NoError(t, cmd.Command.Flags().Set("namespace-permission", "superwrite"))
			},
			expectedErr: `invalid namespace permission "superwrite"`,
		},
		{
			name: "GetServiceAccountError",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.Command.Flags().StringVar(&cmd.Name, "name", "", "")
				require.NoError(t, cmd.Command.Flags().Set("name", "new-name"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
		},
		{
			name: "UpdateServiceAccountError",
			setupCmd: func(cmd *temporalcloudcli.CloudServiceAccountUpdateCommand) {
				cmd.Command.Flags().StringVar(&cmd.Name, "name", "", "")
				require.NoError(t, cmd.Command.Flags().Set("name", "new-name"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: accountScopedSA}, nil)
				c.EXPECT().
					UpdateServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("internal error"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			expectedErr:   "internal error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := temporalcloudcli.CloudServiceAccountUpdateCommand{ServiceAccountId: "sa-1"}
			if tt.setupCmd != nil {
				tt.setupCmd(&cmd)
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- EditServiceAccount ---

func TestEditServiceAccount(t *testing.T) {
	oldSpec := &identityv1.ServiceAccountSpec{
		Name: "my-sa",
		Access: &identityv1.Access{
			AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_DEVELOPER},
		},
	}
	editedSpec := &identityv1.ServiceAccountSpec{
		Name: "my-sa-renamed",
		Access: &identityv1.Access{
			AccountAccess: &identityv1.AccountAccess{Role: identityv1.AccountAccess_ROLE_ADMIN},
		},
	}
	op := &operation.AsyncOperation{Id: "op-edit-sa"}

	tests := []struct {
		name                    string
		resourceVersion         string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		editorOptions           temporalcloudcli.TestEditorOptions
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{
						ServiceAccount: &identityv1.ServiceAccount{Id: "sa-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateServiceAccount(mock.Anything, &cloudservice.UpdateServiceAccountRequest{
						ServiceAccountId: "sa-1",
						Spec:             editedSpec,
						ResourceVersion:  "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateServiceAccountResponse{AsyncOperation: op}, nil)
			},
			editorOptions:      temporalcloudcli.TestEditorOptions{Modified: editedSpec},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-edit-sa"},
		},
		{
			name:            "ResourceVersionOverride",
			resourceVersion: "rv-override",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{
						ServiceAccount: &identityv1.ServiceAccount{Id: "sa-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateServiceAccount(mock.Anything, &cloudservice.UpdateServiceAccountRequest{
						ServiceAccountId: "sa-1",
						Spec:             editedSpec,
						ResourceVersion:  "rv-override",
					}, mock.Anything).
					Return(&cloudservice.UpdateServiceAccountResponse{AsyncOperation: op}, nil)
			},
			editorOptions:      temporalcloudcli.TestEditorOptions{Modified: editedSpec},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-edit-sa"},
		},
		{
			name: "GetServiceAccountError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
		},
		{
			name: "EditorError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{
						ServiceAccount: &identityv1.ServiceAccount{Id: "sa-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
			},
			editorOptions: temporalcloudcli.TestEditorOptions{EditorError: errors.New("editor cancelled")},
			expectedErr:   "editor cancelled",
		},
		{
			name: "PromptDeclined",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{
						ServiceAccount: &identityv1.ServiceAccount{Id: "sa-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
			},
			editorOptions: temporalcloudcli.TestEditorOptions{Modified: editedSpec},
			promptOptions: temporalcloudcli.TestPromptOptions{
				ExpectPrompApply: true,
				PromptError:      errors.New("Aborting apply."),
			},
			expectedErr: "Aborting apply.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &temporalcloudcli.CloudServiceAccountEditCommand{ServiceAccountId: "sa-1"}
			cmd.ResourceVersion = tt.resourceVersion
			temporalcloudcli.TestCommand(t, cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				EditorOptions:           tt.editorOptions,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- SetServiceAccountProjectAccess ---

func TestSetServiceAccountProjectAccess(t *testing.T) {
	testSA := &identityv1.ServiceAccount{
		Id:              "sa-1",
		ResourceVersion: "rv-1",
		Spec:            &identityv1.ServiceAccountSpec{Name: "my-sa"},
	}
	op := &operation.AsyncOperation{Id: "op-set-project-access"}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudServiceAccountSetProjectAccessCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudServiceAccountSetProjectAccessCommand{
				ServiceAccountId: "sa-1",
				ProjectId:        "project-1",
				ProjectRole:      "write",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: testSA}, nil)
				c.EXPECT().
					SetServiceAccountProjectAccess(mock.Anything, &cloudservice.SetServiceAccountProjectAccessRequest{
						ProjectId:        "project-1",
						ServiceAccountId: "sa-1",
						Access:           &identityv1.ProjectAccess{Role: identityv1.ProjectAccess_PROJECT_ROLE_WRITE},
						ResourceVersion:  "rv-1",
					}, mock.Anything).
					Return(&cloudservice.SetServiceAccountProjectAccessResponse{AsyncOperation: op}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-set-project-access"},
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudServiceAccountSetProjectAccessCommand{
				ServiceAccountId:       "sa-1",
				ProjectId:              "project-1",
				ProjectRole:            "member",
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: testSA}, nil)
				c.EXPECT().
					SetServiceAccountProjectAccess(mock.Anything, &cloudservice.SetServiceAccountProjectAccessRequest{
						ProjectId:        "project-1",
						ServiceAccountId: "sa-1",
						Access:           &identityv1.ProjectAccess{Role: identityv1.ProjectAccess_PROJECT_ROLE_MEMBER},
						ResourceVersion:  "rv-override",
					}, mock.Anything).
					Return(&cloudservice.SetServiceAccountProjectAccessResponse{AsyncOperation: op}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-set-project-access"},
		},
		{
			name: "InvalidProjectRole",
			cmd: temporalcloudcli.CloudServiceAccountSetProjectAccessCommand{
				ServiceAccountId: "sa-1",
				ProjectId:        "project-1",
				ProjectRole:      "developer",
			},
			expectedErr: `invalid project role "developer"`,
		},
		{
			name: "GetServiceAccountError",
			cmd: temporalcloudcli.CloudServiceAccountSetProjectAccessCommand{
				ServiceAccountId: "sa-1",
				ProjectId:        "project-1",
				ProjectRole:      "write",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudServiceAccountSetProjectAccessCommand{
				ServiceAccountId: "sa-1",
				ProjectId:        "project-1",
				ProjectRole:      "write",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: testSA}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting set.",
		},
		{
			name: "APIError",
			cmd: temporalcloudcli.CloudServiceAccountSetProjectAccessCommand{
				ServiceAccountId: "sa-1",
				ProjectId:        "project-1",
				ProjectRole:      "write",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: testSA}, nil)
				c.EXPECT().
					SetServiceAccountProjectAccess(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("internal error"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedErr:   "internal error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- RemoveServiceAccountProjectAccess ---

func TestRemoveServiceAccountProjectAccess(t *testing.T) {
	testSA := &identityv1.ServiceAccount{
		Id:              "sa-1",
		ResourceVersion: "rv-1",
		Spec:            &identityv1.ServiceAccountSpec{Name: "my-sa"},
	}
	op := &operation.AsyncOperation{Id: "op-remove-project-access"}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudServiceAccountRemoveProjectAccessCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudServiceAccountRemoveProjectAccessCommand{
				ServiceAccountId: "sa-1",
				ProjectId:        "project-1",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: testSA}, nil)
				c.EXPECT().
					SetServiceAccountProjectAccess(mock.Anything, &cloudservice.SetServiceAccountProjectAccessRequest{
						ProjectId:        "project-1",
						ServiceAccountId: "sa-1",
						ResourceVersion:  "rv-1",
					}, mock.Anything).
					Return(&cloudservice.SetServiceAccountProjectAccessResponse{AsyncOperation: op}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-remove-project-access"},
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudServiceAccountRemoveProjectAccessCommand{
				ServiceAccountId:       "sa-1",
				ProjectId:              "project-1",
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: testSA}, nil)
				c.EXPECT().
					SetServiceAccountProjectAccess(mock.Anything, &cloudservice.SetServiceAccountProjectAccessRequest{
						ProjectId:        "project-1",
						ServiceAccountId: "sa-1",
						ResourceVersion:  "rv-override",
					}, mock.Anything).
					Return(&cloudservice.SetServiceAccountProjectAccessResponse{AsyncOperation: op}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-remove-project-access"},
		},
		{
			name: "GetServiceAccountError",
			cmd: temporalcloudcli.CloudServiceAccountRemoveProjectAccessCommand{
				ServiceAccountId: "sa-1",
				ProjectId:        "project-1",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudServiceAccountRemoveProjectAccessCommand{
				ServiceAccountId: "sa-1",
				ProjectId:        "project-1",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: testSA}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: false},
			expectedErr:   "Aborting remove.",
		},
		{
			name: "APIError",
			cmd: temporalcloudcli.CloudServiceAccountRemoveProjectAccessCommand{
				ServiceAccountId: "sa-1",
				ProjectId:        "project-1",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: testSA}, nil)
				c.EXPECT().
					SetServiceAccountProjectAccess(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("internal error"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptResult: true},
			expectedErr:   "internal error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- GetServiceAccount ---

func TestGetServiceAccount(t *testing.T) {
	testServiceAccount := &identityv1.ServiceAccount{
		Id:   "sa-1",
		Spec: &identityv1.ServiceAccountSpec{Name: "my-sa"},
	}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudServiceAccountGetCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudServiceAccountGetCommand{ServiceAccountId: "sa-1"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, &cloudservice.GetServiceAccountRequest{ServiceAccountId: "sa-1"}, mock.Anything).
					Return(&cloudservice.GetServiceAccountResponse{ServiceAccount: testServiceAccount}, nil)
			},
			expectedJsonOutput: testServiceAccount,
		},
		{
			name: "APIError",
			cmd:  temporalcloudcli.CloudServiceAccountGetCommand{ServiceAccountId: "sa-1"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccount(mock.Anything, mock.Anything, mock.Anything).
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

// --- ListServiceAccounts ---

func TestListServiceAccounts(t *testing.T) {
	testServiceAccounts := []*identityv1.ServiceAccount{
		{Id: "sa-1", Spec: &identityv1.ServiceAccountSpec{Name: "my-sa"}},
		{Id: "sa-2", Spec: &identityv1.ServiceAccountSpec{Name: "other-sa"}},
	}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudServiceAccountListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudServiceAccountListCommand{},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccounts(mock.Anything, &cloudservice.GetServiceAccountsRequest{}, mock.Anything).
					Return(&cloudservice.GetServiceAccountsResponse{
						ServiceAccount: testServiceAccounts,
					}, nil)
			},
			expectedJsonOutput: struct {
				ServiceAccounts []*identityv1.ServiceAccount `json:"ServiceAccounts"`
				NextPageToken   string                       `json:"NextPageToken"`
			}{
				ServiceAccounts: testServiceAccounts,
				NextPageToken:   "",
			},
		},
		{
			name: "WithPagination",
			cmd:  temporalcloudcli.CloudServiceAccountListCommand{PageSize: 10, PageToken: "tok-abc"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccounts(mock.Anything, &cloudservice.GetServiceAccountsRequest{
						PageSize:  10,
						PageToken: "tok-abc",
					}, mock.Anything).
					Return(&cloudservice.GetServiceAccountsResponse{
						ServiceAccount: testServiceAccounts,
						NextPageToken:  "tok-next",
					}, nil)
			},
			expectedJsonOutput: struct {
				ServiceAccounts []*identityv1.ServiceAccount `json:"ServiceAccounts"`
				NextPageToken   string                       `json:"NextPageToken"`
			}{
				ServiceAccounts: testServiceAccounts,
				NextPageToken:   "tok-next",
			},
		},
		{
			name: "WithProjectID",
			cmd: temporalcloudcli.CloudServiceAccountListCommand{
				ProjectId: "project-1",
				PageSize:  10,
				PageToken: "tok-abc",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetProjectScopedServiceAccounts(mock.Anything, &cloudservice.GetProjectScopedServiceAccountsRequest{
						ProjectId: "project-1",
						PageSize:  10,
						PageToken: "tok-abc",
					}, mock.Anything).
					Return(&cloudservice.GetProjectScopedServiceAccountsResponse{
						ServiceAccounts: testServiceAccounts,
						NextPageToken:   "tok-next",
					}, nil)
			},
			expectedJsonOutput: struct {
				ServiceAccounts []*identityv1.ServiceAccount `json:"ServiceAccounts"`
				NextPageToken   string                       `json:"NextPageToken"`
			}{
				ServiceAccounts: testServiceAccounts,
				NextPageToken:   "tok-next",
			},
		},
		{
			name: "Empty",
			cmd:  temporalcloudcli.CloudServiceAccountListCommand{},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccounts(mock.Anything, &cloudservice.GetServiceAccountsRequest{}, mock.Anything).
					Return(&cloudservice.GetServiceAccountsResponse{}, nil)
			},
			expectedJsonOutput: struct {
				ServiceAccounts []*identityv1.ServiceAccount `json:"ServiceAccounts"`
				NextPageToken   string                       `json:"NextPageToken"`
			}{
				ServiceAccounts: nil,
				NextPageToken:   "",
			},
		},
		{
			name: "APIError",
			cmd:  temporalcloudcli.CloudServiceAccountListCommand{},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccounts(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("internal error"))
			},
			expectedErr: "internal error",
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

// --- ProjectServiceAccountList ---

func TestProjectServiceAccountList(t *testing.T) {
	assignments := []*identityv1.ServiceAccountProjectAssignment{
		{
			Id:   "sa-1",
			Name: "my-sa",
			ProjectAccess: &identityv1.ProjectAccess{
				Role: identityv1.ProjectAccess_PROJECT_ROLE_WRITE,
			},
		},
	}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudProjectServiceAccountListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudProjectServiceAccountListCommand{
				ProjectId: "project-1",
				PageSize:  10,
				PageToken: "tok-abc",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccountProjectAssignments(mock.Anything, &cloudservice.GetServiceAccountProjectAssignmentsRequest{
						ProjectId: "project-1",
						PageSize:  10,
						PageToken: "tok-abc",
					}, mock.Anything).
					Return(&cloudservice.GetServiceAccountProjectAssignmentsResponse{
						ServiceAccounts: assignments,
						NextPageToken:   "tok-next",
					}, nil)
			},
			expectedJsonOutput: struct {
				ServiceAccounts []*identityv1.ServiceAccountProjectAssignment `json:"ServiceAccounts"`
				NextPageToken   string                                        `json:"NextPageToken"`
			}{
				ServiceAccounts: assignments,
				NextPageToken:   "tok-next",
			},
		},
		{
			name: "APIError",
			cmd:  temporalcloudcli.CloudProjectServiceAccountListCommand{ProjectId: "project-1"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetServiceAccountProjectAssignments(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("internal error"))
			},
			expectedErr: "internal error",
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
