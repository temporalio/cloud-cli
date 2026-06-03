package temporalcloudcli

import (
	"errors"
	"fmt"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudUserGroupGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetUserGroup(cctx, &cloudservice.GetUserGroupRequest{GroupId: c.GroupId})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResource(res.Group, printer.PrintResourceOptions{})
}

func (c *CloudUserGroupListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	req := &cloudservice.GetUserGroupsRequest{
		PageSize:    int32(c.PageSize),
		PageToken:   c.PageToken,
		Namespace:   c.Namespace,
		DisplayName: c.DisplayName,
	}
	if c.GoogleGroupEmailAddress != "" {
		req.GoogleGroup = &cloudservice.GetUserGroupsRequest_GoogleGroupFilter{EmailAddress: c.GoogleGroupEmailAddress}
	}
	if c.ScimGroupIdpId != "" {
		req.ScimGroup = &cloudservice.GetUserGroupsRequest_SCIMGroupFilter{IdpId: c.ScimGroupIdpId}
	}
	res, err := client.GetUserGroups(cctx, req)
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResourceList(
		struct {
			Groups        []*identityv1.UserGroup
			NextPageToken string
		}{Groups: res.Groups, NextPageToken: res.NextPageToken},
		printer.PrintResourceOptions{
			Fields:     []string{"Id", "State", "CreatedTime"},
			SpecFields: []string{"DisplayName"},
		},
		printer.TableOptions{},
	)
}

func (c *CloudUserGroupDeleteCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetUserGroup(cctx, &cloudservice.GetUserGroupRequest{GroupId: c.GroupId})
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
	rv := res.Group.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.DeleteUserGroup(cctx, &cloudservice.DeleteUserGroupRequest{
		GroupId:          c.GroupId,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleDeleteOperation(cctx, resp, err)
}

func (c *CloudUserGroupCreateCloudGroupCommand) run(cctx *CommandContext, _ []string) error {
	accountAccess, err := parseAccountRole(c.AccountRole)
	if err != nil {
		return err
	}
	namespaceAccesses, err := parseNamespaceAccesses(c.NamespaceAccess)
	if err != nil {
		return err
	}
	if c.Command.Flags().Changed("custom-role") {
		if accountAccess == nil {
			return errors.New("--custom-role requires --account-role; a principal must have a account role")
		}
		accountAccess.CustomRoles = dedupeStrings(c.CustomRole)
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
	spec := &identityv1.UserGroupSpec{
		DisplayName: c.DisplayName,
		GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
	}
	if accountAccess != nil || len(namespaceAccesses) > 0 {
		spec.Access = &identityv1.Access{
			AccountAccess:     accountAccess,
			NamespaceAccesses: namespaceAccesses,
		}
	}
	resp, err := client.CreateUserGroup(cctx, &cloudservice.CreateUserGroupRequest{
		Spec:             spec,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudUserGroupCreateGoogleGroupCommand) run(cctx *CommandContext, _ []string) error {
	accountAccess, err := parseAccountRole(c.AccountRole)
	if err != nil {
		return err
	}
	namespaceAccesses, err := parseNamespaceAccesses(c.NamespaceAccess)
	if err != nil {
		return err
	}
	if c.Command.Flags().Changed("custom-role") {
		if accountAccess == nil {
			return errors.New("--custom-role requires --account-role; a principal must have a account role")
		}
		accountAccess.CustomRoles = dedupeStrings(c.CustomRole)
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
	spec := &identityv1.UserGroupSpec{
		DisplayName: c.DisplayName,
		GroupType: &identityv1.UserGroupSpec_GoogleGroup{
			GoogleGroup: &identityv1.GoogleGroupSpec{EmailAddress: c.GoogleGroupEmail},
		},
	}
	if accountAccess != nil || len(namespaceAccesses) > 0 {
		spec.Access = &identityv1.Access{
			AccountAccess:     accountAccess,
			NamespaceAccesses: namespaceAccesses,
		}
	}
	resp, err := client.CreateUserGroup(cctx, &cloudservice.CreateUserGroupRequest{
		Spec:             spec,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudUserGroupCreateScimGroupCommand) run(cctx *CommandContext, _ []string) error {
	accountAccess, err := parseAccountRole(c.AccountRole)
	if err != nil {
		return err
	}
	namespaceAccesses, err := parseNamespaceAccesses(c.NamespaceAccess)
	if err != nil {
		return err
	}
	if c.Command.Flags().Changed("custom-role") {
		if accountAccess == nil {
			return errors.New("--custom-role requires --account-role; a principal must have a account role")
		}
		accountAccess.CustomRoles = dedupeStrings(c.CustomRole)
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
	spec := &identityv1.UserGroupSpec{
		DisplayName: c.DisplayName,
		GroupType: &identityv1.UserGroupSpec_ScimGroup{
			ScimGroup: &identityv1.SCIMGroupSpec{IdpId: c.ScimIdpId},
		},
	}
	if accountAccess != nil || len(namespaceAccesses) > 0 {
		spec.Access = &identityv1.Access{
			AccountAccess:     accountAccess,
			NamespaceAccesses: namespaceAccesses,
		}
	}
	resp, err := client.CreateUserGroup(cctx, &cloudservice.CreateUserGroupRequest{
		Spec:             spec,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudUserGroupApplyCommand) run(cctx *CommandContext, _ []string) error {
	specData, err := loadJSONSpec(c.Spec)
	if err != nil {
		return err
	}
	spec := &identityv1.UserGroupSpec{}
	if err := cctx.UnmarshalProtoJSON(specData, spec); err != nil {
		return fmt.Errorf("failed to parse JSON spec: %w", err)
	}
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	// Look up existing group by display name.
	res, err := client.GetUserGroups(cctx, &cloudservice.GetUserGroupsRequest{DisplayName: spec.DisplayName})
	if err != nil {
		return err
	}
	var existingSpec *identityv1.UserGroupSpec
	var existingGroupID string
	rv := c.ResourceVersion
	if len(res.Groups) > 0 {
		existing := res.Groups[0]
		existingSpec = existing.Spec
		existingGroupID = existing.Id
		if rv == "" {
			rv = existing.ResourceVersion
		}
	}
	yes, err := cctx.GetPrompter().PromptApply(existingSpec, spec, c.VerboseDiff)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting apply.")
	}
	if existingGroupID != "" {
		resp, err := client.UpdateUserGroup(cctx, &cloudservice.UpdateUserGroupRequest{
			GroupId:          existingGroupID,
			Spec:             spec,
			ResourceVersion:  rv,
			AsyncOperationId: c.AsyncOperationId,
		})
		return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
	}
	resp, err := client.CreateUserGroup(cctx, &cloudservice.CreateUserGroupRequest{
		Spec:             spec,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudUserGroupEditCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetUserGroup(cctx, &cloudservice.GetUserGroupRequest{GroupId: c.GroupId})
	if err != nil {
		return err
	}
	group := res.Group
	edited, err := cctx.GetEditor().EditProto(group.Spec)
	if err != nil {
		return err
	}
	newSpec := edited.(*identityv1.UserGroupSpec)
	yes, err := cctx.GetPrompter().PromptApply(group.Spec, newSpec, c.VerboseDiff)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting edit.")
	}
	rv := group.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateUserGroup(cctx, &cloudservice.UpdateUserGroupRequest{
		GroupId:          c.GroupId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudUserGroupUpdateCommand) run(cctx *CommandContext, _ []string) error {
	customRoleProvided := c.Command.Flags().Changed("custom-role")
	if c.AccountRole == "" && len(c.NamespaceAccess) == 0 && !customRoleProvided {
		return errors.New("must provide at least one of --account-role, --namespace-access, or --custom-role")
	}
	// Validate inputs before any API call.
	if _, err := applyNamespaceAccessChanges(nil, c.NamespaceAccess); err != nil {
		return err
	}
	var accountAccess *identityv1.AccountAccess
	if c.AccountRole != "" {
		var err error
		accountAccess, err = parseAccountRole(c.AccountRole)
		if err != nil {
			return err
		}
	}
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetUserGroup(cctx, &cloudservice.GetUserGroupRequest{GroupId: c.GroupId})
	if err != nil {
		return err
	}
	newSpec := proto.Clone(res.Group.Spec).(*identityv1.UserGroupSpec)
	if newSpec.Access == nil {
		newSpec.Access = &identityv1.Access{}
	}
	if accountAccess != nil {
		// Preserve existing CustomRoles so changing the account role
		// doesn't silently wipe them.
		if newSpec.Access.AccountAccess != nil {
			accountAccess.CustomRoles = newSpec.Access.AccountAccess.CustomRoles
		}
		newSpec.Access.AccountAccess = accountAccess
	}
	if len(c.NamespaceAccess) > 0 {
		namespaceAccesses, err := applyNamespaceAccessChanges(newSpec.Access.NamespaceAccesses, c.NamespaceAccess)
		if err != nil {
			return err
		}
		newSpec.Access.NamespaceAccesses = namespaceAccesses
	}
	if customRoleProvided {
		if newSpec.Access.AccountAccess == nil {
			return errors.New("group has no account access; assign an account role with --account-role first")
		}
		newSpec.Access.AccountAccess.CustomRoles = applyCustomRoleChanges(
			newSpec.Access.AccountAccess.CustomRoles,
			c.CustomRole, customRoleProvided,
		)
	}
	yes, err := cctx.GetPrompter().PromptApply(res.Group.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting update.")
	}
	rv := res.Group.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateUserGroup(cctx, &cloudservice.UpdateUserGroupRequest{
		GroupId:          c.GroupId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudUserGroupSetAccountRoleCommand) run(cctx *CommandContext, _ []string) error {
	accountAccess, err := parseAccountRole(c.AccountRole)
	if err != nil {
		return err
	}
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetUserGroup(cctx, &cloudservice.GetUserGroupRequest{GroupId: c.GroupId})
	if err != nil {
		return err
	}
	newSpec := proto.Clone(res.Group.Spec).(*identityv1.UserGroupSpec)
	if newSpec.Access == nil {
		newSpec.Access = &identityv1.Access{}
	}
	if newSpec.Access.AccountAccess != nil {
		accountAccess.CustomRoles = newSpec.Access.AccountAccess.CustomRoles
	}
	newSpec.Access.AccountAccess = accountAccess
	yes, err := cctx.GetPrompter().PromptApply(res.Group.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting set-account-role.")
	}
	rv := res.Group.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateUserGroup(cctx, &cloudservice.UpdateUserGroupRequest{
		GroupId:          c.GroupId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudUserGroupSetNamespacePermissionsCommand) run(cctx *CommandContext, _ []string) error {
	// Validate inputs before any API call.
	if _, err := applyNamespaceAccessChanges(nil, c.NamespaceAccess); err != nil {
		return err
	}
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetUserGroup(cctx, &cloudservice.GetUserGroupRequest{GroupId: c.GroupId})
	if err != nil {
		return err
	}
	newSpec := proto.Clone(res.Group.Spec).(*identityv1.UserGroupSpec)
	if newSpec.Access == nil {
		newSpec.Access = &identityv1.Access{}
	}
	namespaceAccesses, err := applyNamespaceAccessChanges(newSpec.Access.NamespaceAccesses, c.NamespaceAccess)
	if err != nil {
		return err
	}
	newSpec.Access.NamespaceAccesses = namespaceAccesses
	yes, err := cctx.GetPrompter().PromptApply(res.Group.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting set-namespace-permissions.")
	}
	rv := res.Group.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateUserGroup(cctx, &cloudservice.UpdateUserGroupRequest{
		GroupId:          c.GroupId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudUserGroupSetCustomRolesCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetUserGroup(cctx, &cloudservice.GetUserGroupRequest{GroupId: c.GroupId})
	if err != nil {
		return err
	}
	newSpec := proto.Clone(res.Group.Spec).(*identityv1.UserGroupSpec)
	if newSpec.Access == nil || newSpec.Access.AccountAccess == nil {
		return errors.New("group has no account access; assign an account role with `temporal cloud user-group set-account-role` first")
	}
	newSpec.Access.AccountAccess.CustomRoles = dedupeStrings(c.CustomRole)

	yes, err := cctx.GetPrompter().PromptApply(res.Group.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting set.")
	}
	rv := res.Group.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateUserGroup(cctx, &cloudservice.UpdateUserGroupRequest{
		GroupId:          c.GroupId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}
