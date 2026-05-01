package temporalcloudcli

import (
	"errors"
	"fmt"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudCustomRoleGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetCustomRole(cctx, &cloudservice.GetCustomRoleRequest{RoleId: c.RoleId})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResource(res.CustomRole, printer.PrintResourceOptions{})
}

func (c *CloudCustomRoleListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetCustomRoles(cctx, &cloudservice.GetCustomRolesRequest{
		PageSize:  int32(c.PageSize),
		PageToken: c.PageToken,
	})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResourceList(
		struct {
			CustomRoles   []*identityv1.CustomRole
			NextPageToken string
		}{CustomRoles: res.CustomRoles, NextPageToken: res.NextPageToken},
		printer.PrintResourceOptions{
			Fields:     []string{"Id", "State", "CreatedTime"},
			SpecFields: []string{"Name", "Description"},
		},
		printer.TableOptions{},
	)
}

func (c *CloudCustomRoleCreateCommand) run(cctx *CommandContext, _ []string) error {
	spec, err := parseCustomRoleSpec(cctx, c.Spec)
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
	resp, err := client.CreateCustomRole(cctx, &cloudservice.CreateCustomRoleRequest{
		Spec:             spec,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).
		HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudCustomRoleUpdateCommand) run(cctx *CommandContext, _ []string) error {
	spec, err := parseCustomRoleSpec(cctx, c.Spec)
	if err != nil {
		return err
	}
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetCustomRole(cctx, &cloudservice.GetCustomRoleRequest{RoleId: c.RoleId})
	if err != nil {
		return err
	}
	rv := res.CustomRole.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateCustomRole(cctx, &cloudservice.UpdateCustomRoleRequest{
		RoleId:           c.RoleId,
		Spec:             spec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudCustomRoleDeleteCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetCustomRole(cctx, &cloudservice.GetCustomRoleRequest{RoleId: c.RoleId})
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
	rv := res.CustomRole.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.DeleteCustomRole(cctx, &cloudservice.DeleteCustomRoleRequest{
		RoleId:           c.RoleId,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleDeleteOperation(cctx, resp, err)
}

func (c *CloudCustomRoleApplyCommand) run(cctx *CommandContext, _ []string) error {
	spec, err := parseCustomRoleSpec(cctx, c.Spec)
	if err != nil {
		return err
	}
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	existing, err := findCustomRoleByName(cctx, client, spec.Name)
	if err != nil {
		return err
	}
	var existingSpec *identityv1.CustomRoleSpec
	var existingRoleID string
	rv := c.ResourceVersion
	if existing != nil {
		existingSpec = existing.Spec
		existingRoleID = existing.Id
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
	if existingRoleID != "" {
		resp, err := client.UpdateCustomRole(cctx, &cloudservice.UpdateCustomRoleRequest{
			RoleId:           existingRoleID,
			Spec:             spec,
			ResourceVersion:  rv,
			AsyncOperationId: c.AsyncOperationId,
		})
		return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
	}
	resp, err := client.CreateCustomRole(cctx, &cloudservice.CreateCustomRoleRequest{
		Spec:             spec,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).
		HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudCustomRoleEditCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetCustomRole(cctx, &cloudservice.GetCustomRoleRequest{RoleId: c.RoleId})
	if err != nil {
		return err
	}
	role := res.CustomRole
	edited, err := cctx.GetEditor().EditProto(role.Spec)
	if err != nil {
		return err
	}
	newSpec := edited.(*identityv1.CustomRoleSpec)
	yes, err := cctx.GetPrompter().PromptApply(role.Spec, newSpec, c.VerboseDiff)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting edit.")
	}
	rv := role.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateCustomRole(cctx, &cloudservice.UpdateCustomRoleRequest{
		RoleId:           c.RoleId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func parseCustomRoleSpec(cctx *CommandContext, raw string) (*identityv1.CustomRoleSpec, error) {
	data, err := loadJSONSpec(raw)
	if err != nil {
		return nil, err
	}
	spec := &identityv1.CustomRoleSpec{}
	if err := cctx.UnmarshalProtoJSON(data, spec); err != nil {
		return nil, fmt.Errorf("failed to parse JSON spec: %w", err)
	}
	return spec, nil
}

// AIDEV-NOTE: GetCustomRoles has no name filter; we page client-side.
// Roles per account are typically small (tens), so paging is fine.
// Names are not unique server-side, so we scan every page and error on
// duplicates rather than silently picking one.
func findCustomRoleByName(
	cctx *CommandContext,
	client cloudservice.CloudServiceClient,
	name string,
) (*identityv1.CustomRole, error) {
	var match *identityv1.CustomRole
	var pageToken string
	for {
		res, err := client.GetCustomRoles(cctx, &cloudservice.GetCustomRolesRequest{PageToken: pageToken})
		if err != nil {
			return nil, err
		}
		for _, r := range res.CustomRoles {
			if r.Spec == nil || r.Spec.Name != name {
				continue
			}
			if match != nil {
				return nil, fmt.Errorf("multiple custom roles found with name %q (ids: %s, %s); use update with --role-id to disambiguate", name, match.Id, r.Id)
			}
			match = r
		}
		if res.NextPageToken == "" {
			return match, nil
		}
		pageToken = res.NextPageToken
	}
}
