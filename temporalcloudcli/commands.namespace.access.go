package temporalcloudcli

import (
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

// AIDEV-NOTE: These commands wrap the read-only Get*NamespaceAssignments RPCs
// added in cloud-sdk-go v0.14.x. They list the identities (users, service
// accounts, user groups) that have access to a namespace, including access
// inherited through account/project roles (InheritedAccess).

func (c *CloudNamespaceUserListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetUserNamespaceAssignments(cctx, &cloudservice.GetUserNamespaceAssignmentsRequest{
		Namespace: c.Namespace,
		PageSize:  int32(c.PageSize),
		PageToken: c.PageToken,
	})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResourceList(
		struct {
			Users         []*identityv1.UserNamespaceAssignment
			NextPageToken string
		}{
			Users:         res.Users,
			NextPageToken: res.NextPageToken,
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Id", "InheritedAccess"},
			SpecFields: []string{"Email", "NamespaceAccess"},
		},
		printer.TableOptions{},
	)
}

func (c *CloudNamespaceServiceAccountListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetServiceAccountNamespaceAssignments(cctx, &cloudservice.GetServiceAccountNamespaceAssignmentsRequest{
		Namespace: c.Namespace,
		PageSize:  int32(c.PageSize),
		PageToken: c.PageToken,
	})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResourceList(
		struct {
			ServiceAccounts []*identityv1.ServiceAccountNamespaceAssignment
			NextPageToken   string
		}{
			ServiceAccounts: res.ServiceAccounts,
			NextPageToken:   res.NextPageToken,
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Id", "InheritedAccess"},
			SpecFields: []string{"Name", "NamespaceAccess"},
		},
		printer.TableOptions{},
	)
}

func (c *CloudNamespaceUserGroupListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetUserGroupNamespaceAssignments(cctx, &cloudservice.GetUserGroupNamespaceAssignmentsRequest{
		Namespace: c.Namespace,
		PageSize:  int32(c.PageSize),
		PageToken: c.PageToken,
	})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResourceList(
		struct {
			Groups        []*identityv1.UserGroupNamespaceAssignment
			NextPageToken string
		}{
			Groups:        res.Groups,
			NextPageToken: res.NextPageToken,
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Id", "InheritedAccess"},
			SpecFields: []string{"DisplayName", "NamespaceAccess"},
		},
		printer.TableOptions{},
	)
}
