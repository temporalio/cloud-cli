package temporalcloudcli

import (
	"errors"
	"fmt"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudServiceAccountDeleteCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetServiceAccount(cctx, &cloudservice.GetServiceAccountRequest{ServiceAccountId: c.ServiceAccountId})
	if err != nil {
		return err
	}
	yes, err := cctx.GetPrompter().PromptYes("Delete")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting delete.")
	}
	rv := res.ServiceAccount.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.DeleteServiceAccount(cctx, &cloudservice.DeleteServiceAccountRequest{
		ServiceAccountId: c.ServiceAccountId,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleDeleteOperation(cctx, resp, err)
}

func (c *CloudServiceAccountCreateCommand) run(cctx *CommandContext, _ []string) error {
	accountAccess, err := parseAccountRole(c.AccountRole)
	if err != nil {
		return err
	}
	namespaceAccesses, err := parseNamespaceAccesses(c.NamespaceAccess)
	if err != nil {
		return err
	}
	yes, err := cctx.GetPrompter().PromptYes("Create")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
	}
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	resp, err := client.CreateServiceAccount(cctx, &cloudservice.CreateServiceAccountRequest{
		Spec: &identityv1.ServiceAccountSpec{
			Name:        c.Name,
			Description: c.Description,
			Access: &identityv1.Access{
				AccountAccess:     accountAccess,
				NamespaceAccesses: namespaceAccesses,
			},
		},
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudServiceAccountCreateNamespaceScopedCommand) run(cctx *CommandContext, _ []string) error {
	perm, ok := namespacePermissionNames[c.NamespacePermission]
	if !ok {
		return fmt.Errorf("invalid namespace permission %q: must be one of admin, write, read", c.NamespacePermission)
	}
	yes, err := cctx.GetPrompter().PromptYes("Create")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
	}
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	resp, err := client.CreateServiceAccount(cctx, &cloudservice.CreateServiceAccountRequest{
		Spec: &identityv1.ServiceAccountSpec{
			Name:        c.Name,
			Description: c.Description,
			NamespaceScopedAccess: &identityv1.NamespaceScopedAccess{
				Namespace: c.Namespace,
				Access:    &identityv1.NamespaceAccess{Permission: perm},
			},
		},
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudServiceAccountUpdateCommand) run(cctx *CommandContext, _ []string) error {
	// Validate input formats before any API call.
	accountRoleChanged := c.Command.Flags().Changed("account-role")
	namespaceAccessChanged := c.Command.Flags().Changed("namespace-access")
	namespacePermissionChanged := c.Command.Flags().Changed("namespace-permission")

	if accountRoleChanged {
		if _, ok := accountRoleNames[c.AccountRole]; !ok {
			return fmt.Errorf("invalid account role %q: must be one of owner, admin, developer, finance-admin, read, metrics-read", c.AccountRole)
		}
	}
	if namespaceAccessChanged {
		if _, err := applyNamespaceAccessChanges(nil, c.NamespaceAccess); err != nil {
			return err
		}
	}
	if namespacePermissionChanged {
		if _, ok := namespacePermissionNames[c.NamespacePermission]; !ok {
			return fmt.Errorf("invalid namespace permission %q: must be one of admin, write, read", c.NamespacePermission)
		}
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetServiceAccount(cctx, &cloudservice.GetServiceAccountRequest{ServiceAccountId: c.ServiceAccountId})
	if err != nil {
		return err
	}
	sa := res.ServiceAccount
	newSpec := proto.Clone(sa.Spec).(*identityv1.ServiceAccountSpec)

	isNamespaceScoped := newSpec.NamespaceScopedAccess != nil

	if isNamespaceScoped && (accountRoleChanged || namespaceAccessChanged) {
		return errors.New("--account-role and --namespace-access are not valid for namespace-scoped service accounts")
	}
	if !isNamespaceScoped && namespacePermissionChanged {
		return errors.New("--namespace-permission is not valid for account-scoped service accounts")
	}

	if c.Command.Flags().Changed("name") {
		newSpec.Name = c.Name
	}
	if c.Command.Flags().Changed("description") {
		newSpec.Description = c.Description
	}
	if accountRoleChanged {
		if newSpec.Access == nil {
			newSpec.Access = &identityv1.Access{}
		}
		newSpec.Access.AccountAccess = &identityv1.AccountAccess{Role: accountRoleNames[c.AccountRole]}
	}
	if namespaceAccessChanged {
		if newSpec.Access == nil {
			newSpec.Access = &identityv1.Access{}
		}
		newSpec.Access.NamespaceAccesses, err = applyNamespaceAccessChanges(newSpec.Access.NamespaceAccesses, c.NamespaceAccess)
		if err != nil {
			return err
		}
	}
	if namespacePermissionChanged {
		if newSpec.NamespaceScopedAccess.Access == nil {
			newSpec.NamespaceScopedAccess.Access = &identityv1.NamespaceAccess{}
		}
		newSpec.NamespaceScopedAccess.Access.Permission = namespacePermissionNames[c.NamespacePermission]
	}

	yes, err := cctx.GetPrompter().PromptApply(sa.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting update.")
	}

	rv := sa.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateServiceAccount(cctx, &cloudservice.UpdateServiceAccountRequest{
		ServiceAccountId: c.ServiceAccountId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudServiceAccountEditCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetServiceAccount(cctx, &cloudservice.GetServiceAccountRequest{ServiceAccountId: c.ServiceAccountId})
	if err != nil {
		return err
	}
	sa := res.ServiceAccount

	edited, err := cctx.GetEditor().EditProto(sa.Spec)
	if err != nil {
		return err
	}
	newSpec := edited.(*identityv1.ServiceAccountSpec)

	yes, err := cctx.GetPrompter().PromptApply(sa.Spec, newSpec, c.VerboseDiff)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting edit.")
	}

	rv := sa.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateServiceAccount(cctx, &cloudservice.UpdateServiceAccountRequest{
		ServiceAccountId: c.ServiceAccountId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudServiceAccountGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetServiceAccount(cctx, &cloudservice.GetServiceAccountRequest{ServiceAccountId: c.ServiceAccountId})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResource(res.ServiceAccount, printer.PrintResourceOptions{})
}

func (c *CloudServiceAccountListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetServiceAccounts(cctx, &cloudservice.GetServiceAccountsRequest{
		PageSize:  int32(c.PageSize),
		PageToken: c.PageToken,
	})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResourceList(
		struct {
			ServiceAccounts []*identityv1.ServiceAccount
			NextPageToken   string
		}{
			ServiceAccounts: res.ServiceAccount,
			NextPageToken:   res.NextPageToken,
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Id", "State", "CreatedTime"},
			SpecFields: []string{"Name"},
		},
		printer.TableOptions{},
	)
}
