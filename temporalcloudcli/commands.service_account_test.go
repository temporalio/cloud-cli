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
			expectedErr: "--account-role, --namespace-access, and --custom-role are not valid for namespace-scoped service accounts",
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
			expectedErr: "--namespace-permission is not valid for account-scoped service accounts",
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
