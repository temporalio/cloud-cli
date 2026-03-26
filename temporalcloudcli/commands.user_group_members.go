package temporalcloudcli

import (
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudUserGroupMembersListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetUserGroupMembers(cctx, &cloudservice.GetUserGroupMembersRequest{
		GroupId:   c.GroupId,
		PageSize:  int32(c.PageSize),
		PageToken: c.PageToken,
	})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResourceList(
		struct {
			Members       []*identityv1.UserGroupMember
			NextPageToken string
		}{Members: res.Members, NextPageToken: res.NextPageToken},
		printer.PrintResourceOptions{},
		printer.TableOptions{},
	)
}

func (c *CloudUserGroupMembersAddCommand) run(cctx *CommandContext, _ []string) error {
	if err := validateUserIdentification(c.UserIdentificationOptions); err != nil {
		return err
	}
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	user, err := resolveUser(cctx, client, c.UserIdentificationOptions)
	if err != nil {
		return err
	}
	resp, err := client.AddUserGroupMember(cctx, &cloudservice.AddUserGroupMemberRequest{
		GroupId:          c.GroupId,
		MemberId:         &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: user.Id}},
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudUserGroupMembersRemoveCommand) run(cctx *CommandContext, _ []string) error {
	if err := validateUserIdentification(c.UserIdentificationOptions); err != nil {
		return err
	}
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	user, err := resolveUser(cctx, client, c.UserIdentificationOptions)
	if err != nil {
		return err
	}
	resp, err := client.RemoveUserGroupMember(cctx, &cloudservice.RemoveUserGroupMemberRequest{
		GroupId:          c.GroupId,
		MemberId:         &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: user.Id}},
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleDeleteOperation(cctx, resp, err)
}
