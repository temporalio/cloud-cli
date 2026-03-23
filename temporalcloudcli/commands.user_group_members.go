package temporalcloudcli

import (
	"context"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type (
	ListUserGroupMembersParams struct {
		GroupId   string
		PageSize  int32
		PageToken string

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
	}

	AddUserGroupMemberParams struct {
		GroupId            string
		UserIdentification UserIdentificationOptions
		AsyncOperationID   string

		Cloud            cloudservice.CloudServiceClient
		OperationHandler AsyncOperationHandler
	}

	RemoveUserGroupMemberParams struct {
		GroupId            string
		UserIdentification UserIdentificationOptions
		AsyncOperationID   string

		Cloud            cloudservice.CloudServiceClient
		OperationHandler AsyncOperationHandler
	}
)

func ListUserGroupMembers(ctx context.Context, params ListUserGroupMembersParams) error {
	res, err := params.Cloud.GetUserGroupMembers(ctx, &cloudservice.GetUserGroupMembersRequest{
		GroupId:   params.GroupId,
		PageSize:  params.PageSize,
		PageToken: params.PageToken,
	})
	if err != nil {
		return err
	}
	return params.Printer.PrintResourceList(
		struct {
			Members       []*identityv1.UserGroupMember
			NextPageToken string
		}{
			Members:       res.Members,
			NextPageToken: res.NextPageToken,
		},
		printer.PrintResourceOptions{},
		printer.TableOptions{},
	)
}

func AddUserGroupMember(ctx context.Context, params AddUserGroupMemberParams) error {
	if err := validateUserIdentification(params.UserIdentification); err != nil {
		return err
	}
	user, err := resolveUser(ctx, params.Cloud, params.UserIdentification)
	if err != nil {
		return err
	}
	add := wrapCreateOperation(params.Cloud.AddUserGroupMember, params.OperationHandler, func(_ *cloudservice.AddUserGroupMemberResponse) string { return params.GroupId })
	return add(ctx, &cloudservice.AddUserGroupMemberRequest{
		GroupId:          params.GroupId,
		MemberId:         &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: user.Id}},
		AsyncOperationId: params.AsyncOperationID,
	})
}

func RemoveUserGroupMember(ctx context.Context, params RemoveUserGroupMemberParams) error {
	if err := validateUserIdentification(params.UserIdentification); err != nil {
		return err
	}
	user, err := resolveUser(ctx, params.Cloud, params.UserIdentification)
	if err != nil {
		return err
	}
	remove := wrapDeleteOperation(params.Cloud.RemoveUserGroupMember, params.OperationHandler, params.GroupId)
	return remove(ctx, &cloudservice.RemoveUserGroupMemberRequest{
		GroupId:          params.GroupId,
		MemberId:         &identityv1.UserGroupMemberId{MemberType: &identityv1.UserGroupMemberId_UserId{UserId: user.Id}},
		AsyncOperationId: params.AsyncOperationID,
	})
}

func (c *CloudUserGroupMembersListCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return ListUserGroupMembers(cctx.Context, ListUserGroupMembersParams{
		GroupId:   c.GroupId,
		PageSize:  int32(c.PageSize),
		PageToken: c.PageToken,
		Cloud:     cloudClient.CloudService(),
		Printer:   cctx.Printer,
	})
}

func (c *CloudUserGroupMembersAddCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return AddUserGroupMember(cctx.Context, AddUserGroupMemberParams{
		GroupId:            c.GroupId,
		UserIdentification: c.UserIdentificationOptions,
		AsyncOperationID:   c.AsyncOperationId,
		Cloud:              cloudClient.CloudService(),
		OperationHandler:   NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudUserGroupMembersRemoveCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return RemoveUserGroupMember(cctx.Context, RemoveUserGroupMemberParams{
		GroupId:            c.GroupId,
		UserIdentification: c.UserIdentificationOptions,
		AsyncOperationID:   c.AsyncOperationId,
		Cloud:              cloudClient.CloudService(),
		OperationHandler:   NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}
