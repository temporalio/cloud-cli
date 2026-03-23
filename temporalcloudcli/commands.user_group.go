package temporalcloudcli

import (
	"context"
	"errors"
	"fmt"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type (
	CreateCloudGroupParams struct {
		DisplayName       string
		AccountAccess     *identityv1.AccountAccess
		NamespaceAccesses map[string]*identityv1.NamespaceAccess
		AsyncOperationID  string

		Cloud            cloudservice.CloudServiceClient
		OperationHandler AsyncOperationHandler
	}

	CreateGoogleGroupParams struct {
		DisplayName       string
		GoogleGroupEmail  string
		AccountAccess     *identityv1.AccountAccess
		NamespaceAccesses map[string]*identityv1.NamespaceAccess
		AsyncOperationID  string

		Cloud            cloudservice.CloudServiceClient
		OperationHandler AsyncOperationHandler
	}

	CreateSCIMGroupParams struct {
		DisplayName       string
		ScimIdpId         string
		AccountAccess     *identityv1.AccountAccess
		NamespaceAccesses map[string]*identityv1.NamespaceAccess
		AsyncOperationID  string

		Cloud            cloudservice.CloudServiceClient
		OperationHandler AsyncOperationHandler
	}

	ApplyUserGroupParams struct {
		Spec             *identityv1.UserGroupSpec
		ResourceVersion  string
		AsyncOperationID string
		VerboseDiff      bool

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	DeleteUserGroupParams struct {
		GroupId          string
		ResourceVersion  string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		OperationHandler AsyncOperationHandler
	}

	UpdateUserGroupParams struct {
		GroupId          string
		AccountRole      string
		// NamespaceAccesses lists changes in "namespace=permission" format.
		// An empty permission (e.g. "testns=") removes that namespace.
		NamespaceAccesses []string
		ResourceVersion   string
		AsyncOperationID  string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	EditUserGroupParams struct {
		GroupId          string
		ResourceVersion  string
		AsyncOperationID string
		VerboseDiff      bool

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
		// RunEditor opens the existing spec in an editor and writes the result into target.
		// Injected so the function is unit-testable without a real editor process.
		RunEditor func(existing, target proto.Message) error
	}

	SetUserGroupAccountRoleParams struct {
		GroupId          string
		AccountRole      string
		ResourceVersion  string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	SetUserGroupNamespacePermissionsParams struct {
		GroupId string
		// NamespaceAccesses lists changes in "namespace=permission" format.
		// An empty permission (e.g. "testns=") removes that namespace.
		NamespaceAccesses []string
		ResourceVersion   string
		AsyncOperationID  string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	GetUserGroupParams struct {
		GroupId string

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
	}

	ListUserGroupsParams struct {
		PageSize                int32
		PageToken               string
		Namespace               string
		DisplayName             string
		GoogleGroupEmailAddress string
		ScimGroupIdpId          string

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
	}
)

func ApplyUserGroup(ctx context.Context, params ApplyUserGroupParams) error {
	// Look up existing group by display name.
	res, err := params.Cloud.GetUserGroups(ctx, &cloudservice.GetUserGroupsRequest{DisplayName: params.Spec.DisplayName})
	if err != nil {
		return err
	}

	var existingSpec *identityv1.UserGroupSpec
	var existingGroupID string
	rv := params.ResourceVersion

	if len(res.Groups) > 0 {
		existing := res.Groups[0]
		existingSpec = existing.Spec
		existingGroupID = existing.Id
		if rv == "" {
			rv = existing.ResourceVersion
		}
	}

	if err := params.Prompter.PromptApply(existingSpec, params.Spec, params.VerboseDiff); err != nil {
		return err
	}

	if existingGroupID != "" {
		update := runAsyncOperation(params.Cloud.UpdateUserGroup, params.OperationHandler)
		return update(ctx, &cloudservice.UpdateUserGroupRequest{
			GroupId:          existingGroupID,
			Spec:             params.Spec,
			ResourceVersion:  rv,
			AsyncOperationId: params.AsyncOperationID,
		})
	}

	create := runAsyncOperation(params.Cloud.CreateUserGroup, params.OperationHandler)
	return create(ctx, &cloudservice.CreateUserGroupRequest{
		Spec:             params.Spec,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func CreateCloudGroup(ctx context.Context, params CreateCloudGroupParams) error {
	spec := &identityv1.UserGroupSpec{
		DisplayName: params.DisplayName,
		GroupType:   &identityv1.UserGroupSpec_CloudGroup{CloudGroup: &identityv1.CloudGroupSpec{}},
	}
	if params.AccountAccess != nil || len(params.NamespaceAccesses) > 0 {
		spec.Access = &identityv1.Access{
			AccountAccess:     params.AccountAccess,
			NamespaceAccesses: params.NamespaceAccesses,
		}
	}
	create := runAsyncOperation(params.Cloud.CreateUserGroup, params.OperationHandler)
	return create(ctx, &cloudservice.CreateUserGroupRequest{
		Spec:             spec,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func CreateGoogleGroup(ctx context.Context, params CreateGoogleGroupParams) error {
	spec := &identityv1.UserGroupSpec{
		DisplayName: params.DisplayName,
		GroupType: &identityv1.UserGroupSpec_GoogleGroup{
			GoogleGroup: &identityv1.GoogleGroupSpec{EmailAddress: params.GoogleGroupEmail},
		},
	}
	if params.AccountAccess != nil || len(params.NamespaceAccesses) > 0 {
		spec.Access = &identityv1.Access{
			AccountAccess:     params.AccountAccess,
			NamespaceAccesses: params.NamespaceAccesses,
		}
	}
	create := runAsyncOperation(params.Cloud.CreateUserGroup, params.OperationHandler)
	return create(ctx, &cloudservice.CreateUserGroupRequest{
		Spec:             spec,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func CreateSCIMGroup(ctx context.Context, params CreateSCIMGroupParams) error {
	spec := &identityv1.UserGroupSpec{
		DisplayName: params.DisplayName,
		GroupType: &identityv1.UserGroupSpec_ScimGroup{
			ScimGroup: &identityv1.SCIMGroupSpec{IdpId: params.ScimIdpId},
		},
	}
	if params.AccountAccess != nil || len(params.NamespaceAccesses) > 0 {
		spec.Access = &identityv1.Access{
			AccountAccess:     params.AccountAccess,
			NamespaceAccesses: params.NamespaceAccesses,
		}
	}
	create := runAsyncOperation(params.Cloud.CreateUserGroup, params.OperationHandler)
	return create(ctx, &cloudservice.CreateUserGroupRequest{
		Spec:             spec,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func GetUserGroup(ctx context.Context, params GetUserGroupParams) error {
	res, err := params.Cloud.GetUserGroup(ctx, &cloudservice.GetUserGroupRequest{GroupId: params.GroupId})
	if err != nil {
		return err
	}
	return params.Printer.PrintResource(res.Group, printer.PrintResourceOptions{})
}

func ListUserGroups(ctx context.Context, params ListUserGroupsParams) error {
	req := &cloudservice.GetUserGroupsRequest{
		PageSize:    params.PageSize,
		PageToken:   params.PageToken,
		Namespace:   params.Namespace,
		DisplayName: params.DisplayName,
	}
	if params.GoogleGroupEmailAddress != "" {
		req.GoogleGroup = &cloudservice.GetUserGroupsRequest_GoogleGroupFilter{EmailAddress: params.GoogleGroupEmailAddress}
	}
	if params.ScimGroupIdpId != "" {
		req.ScimGroup = &cloudservice.GetUserGroupsRequest_SCIMGroupFilter{IdpId: params.ScimGroupIdpId}
	}
	res, err := params.Cloud.GetUserGroups(ctx, req)
	if err != nil {
		return err
	}

	return params.Printer.PrintResourceList(
		struct {
			Groups        []*identityv1.UserGroup
			NextPageToken string
		}{
			Groups:        res.Groups,
			NextPageToken: res.NextPageToken,
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Id", "State", "CreatedTime"},
			SpecFields: []string{"DisplayName"},
		},
		printer.TableOptions{},
	)
}

func DeleteUserGroup(ctx context.Context, params DeleteUserGroupParams) error {
	group, err := params.Cloud.GetUserGroup(ctx, &cloudservice.GetUserGroupRequest{GroupId: params.GroupId})
	if err != nil {
		return err
	}
	rv := group.Group.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}
	del := runAsyncOperation(params.Cloud.DeleteUserGroup, params.OperationHandler)
	return del(ctx, &cloudservice.DeleteUserGroupRequest{
		GroupId:          params.GroupId,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func UpdateUserGroup(ctx context.Context, params UpdateUserGroupParams) error {
	if params.AccountRole == "" && len(params.NamespaceAccesses) == 0 {
		return errors.New("must provide at least one of --account-role or --namespace-access")
	}
	// Validate namespace access changes before any API call.
	if _, err := applyNamespaceAccessChanges(nil, params.NamespaceAccesses); err != nil {
		return err
	}
	var accountAccess *identityv1.AccountAccess
	if params.AccountRole != "" {
		var err error
		accountAccess, err = parseAccountRole(params.AccountRole)
		if err != nil {
			return err
		}
	}
	res, err := params.Cloud.GetUserGroup(ctx, &cloudservice.GetUserGroupRequest{GroupId: params.GroupId})
	if err != nil {
		return err
	}
	newSpec := proto.Clone(res.Group.Spec).(*identityv1.UserGroupSpec)
	if newSpec.Access == nil {
		newSpec.Access = &identityv1.Access{}
	}
	if accountAccess != nil {
		newSpec.Access.AccountAccess = accountAccess
	}
	if len(params.NamespaceAccesses) > 0 {
		namespaceAccesses, err := applyNamespaceAccessChanges(newSpec.Access.NamespaceAccesses, params.NamespaceAccesses)
		if err != nil {
			return err
		}
		newSpec.Access.NamespaceAccesses = namespaceAccesses
	}
	if err := params.Prompter.PromptApply(res.Group.Spec, newSpec, false); err != nil {
		return err
	}
	rv := res.Group.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}
	update := runAsyncOperation(params.Cloud.UpdateUserGroup, params.OperationHandler)
	return update(ctx, &cloudservice.UpdateUserGroupRequest{
		GroupId:          params.GroupId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func EditUserGroup(ctx context.Context, params EditUserGroupParams) error {
	res, err := params.Cloud.GetUserGroup(ctx, &cloudservice.GetUserGroupRequest{GroupId: params.GroupId})
	if err != nil {
		return err
	}
	newSpec := &identityv1.UserGroupSpec{}
	if err := params.RunEditor(res.Group.Spec, newSpec); err != nil {
		return err
	}
	if err := params.Prompter.PromptApply(res.Group.Spec, newSpec, params.VerboseDiff); err != nil {
		return err
	}
	rv := res.Group.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}
	update := runAsyncOperation(params.Cloud.UpdateUserGroup, params.OperationHandler)
	return update(ctx, &cloudservice.UpdateUserGroupRequest{
		GroupId:          params.GroupId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func SetUserGroupAccountRole(ctx context.Context, params SetUserGroupAccountRoleParams) error {
	accountAccess, err := parseAccountRole(params.AccountRole)
	if err != nil {
		return err
	}
	res, err := params.Cloud.GetUserGroup(ctx, &cloudservice.GetUserGroupRequest{GroupId: params.GroupId})
	if err != nil {
		return err
	}
	newSpec := proto.Clone(res.Group.Spec).(*identityv1.UserGroupSpec)
	if newSpec.Access == nil {
		newSpec.Access = &identityv1.Access{}
	}
	newSpec.Access.AccountAccess = accountAccess
	if err := params.Prompter.PromptApply(res.Group.Spec, newSpec, false); err != nil {
		return err
	}
	rv := res.Group.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}
	update := runAsyncOperation(params.Cloud.UpdateUserGroup, params.OperationHandler)
	return update(ctx, &cloudservice.UpdateUserGroupRequest{
		GroupId:          params.GroupId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func SetUserGroupNamespacePermissions(ctx context.Context, params SetUserGroupNamespacePermissionsParams) error {
	// Validate inputs before any API call.
	if _, err := applyNamespaceAccessChanges(nil, params.NamespaceAccesses); err != nil {
		return err
	}
	res, err := params.Cloud.GetUserGroup(ctx, &cloudservice.GetUserGroupRequest{GroupId: params.GroupId})
	if err != nil {
		return err
	}
	newSpec := proto.Clone(res.Group.Spec).(*identityv1.UserGroupSpec)
	if newSpec.Access == nil {
		newSpec.Access = &identityv1.Access{}
	}
	namespaceAccesses, err := applyNamespaceAccessChanges(newSpec.Access.NamespaceAccesses, params.NamespaceAccesses)
	if err != nil {
		return err
	}
	newSpec.Access.NamespaceAccesses = namespaceAccesses
	if err := params.Prompter.PromptApply(res.Group.Spec, newSpec, false); err != nil {
		return err
	}
	rv := res.Group.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}
	update := runAsyncOperation(params.Cloud.UpdateUserGroup, params.OperationHandler)
	return update(ctx, &cloudservice.UpdateUserGroupRequest{
		GroupId:          params.GroupId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
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

	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return ApplyUserGroup(cctx.Context, ApplyUserGroupParams{
		Spec:             spec,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		VerboseDiff:      c.VerboseDiff,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, spec.DisplayName, c.ClientOptions),
	})
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
	yes, err := cctx.promptYes("Create (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
	}
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return CreateCloudGroup(cctx.Context, CreateCloudGroupParams{
		DisplayName:       c.DisplayName,
		AccountAccess:     accountAccess,
		NamespaceAccesses: namespaceAccesses,
		AsyncOperationID:  c.AsyncOperationId,
		Cloud:             cloudClient.CloudService(),
		OperationHandler:  NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.DisplayName, c.ClientOptions),
	})
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
	yes, err := cctx.promptYes("Create (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
	}
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return CreateGoogleGroup(cctx.Context, CreateGoogleGroupParams{
		DisplayName:       c.DisplayName,
		GoogleGroupEmail:  c.GoogleGroupEmail,
		AccountAccess:     accountAccess,
		NamespaceAccesses: namespaceAccesses,
		AsyncOperationID:  c.AsyncOperationId,
		Cloud:             cloudClient.CloudService(),
		OperationHandler:  NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.DisplayName, c.ClientOptions),
	})
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
	yes, err := cctx.promptYes("Create (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
	}
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return CreateSCIMGroup(cctx.Context, CreateSCIMGroupParams{
		DisplayName:       c.DisplayName,
		ScimIdpId:         c.ScimIdpId,
		AccountAccess:     accountAccess,
		NamespaceAccesses: namespaceAccesses,
		AsyncOperationID:  c.AsyncOperationId,
		Cloud:             cloudClient.CloudService(),
		OperationHandler:  NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.DisplayName, c.ClientOptions),
	})
}

func (c *CloudUserGroupDeleteCommand) run(cctx *CommandContext, _ []string) error {
	yes, err := cctx.promptYes("Delete (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting delete.")
	}
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return DeleteUserGroup(cctx.Context, DeleteUserGroupParams{
		GroupId:          c.GroupId,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		OperationHandler: NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.GroupId, c.ClientOptions),
	})
}

func (c *CloudUserGroupUpdateCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return UpdateUserGroup(cctx.Context, UpdateUserGroupParams{
		GroupId:           c.GroupId,
		AccountRole:       c.AccountRole,
		NamespaceAccesses: c.NamespaceAccess,
		ResourceVersion:   c.ResourceVersion,
		AsyncOperationID:  c.AsyncOperationId,
		Cloud:             cloudClient.CloudService(),
		Prompter:          newPrompter(cctx),
		OperationHandler:  NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.GroupId, c.ClientOptions),
	})
}

func (c *CloudUserGroupEditCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return EditUserGroup(cctx.Context, EditUserGroupParams{
		GroupId:          c.GroupId,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		VerboseDiff:      c.VerboseDiff,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.GroupId, c.ClientOptions),
		RunEditor:        runEditorForJSONEditForProtos,
	})
}

func (c *CloudUserGroupSetAccountRoleCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return SetUserGroupAccountRole(cctx.Context, SetUserGroupAccountRoleParams{
		GroupId:          c.GroupId,
		AccountRole:      c.AccountRole,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.GroupId, c.ClientOptions),
	})
}

func (c *CloudUserGroupSetNamespacePermissionsCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return SetUserGroupNamespacePermissions(cctx.Context, SetUserGroupNamespacePermissionsParams{
		GroupId:           c.GroupId,
		NamespaceAccesses: c.NamespaceAccess,
		ResourceVersion:   c.ResourceVersion,
		AsyncOperationID:  c.AsyncOperationId,
		Cloud:             cloudClient.CloudService(),
		Prompter:          newPrompter(cctx),
		OperationHandler:  NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.GroupId, c.ClientOptions),
	})
}

func (c *CloudUserGroupGetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return GetUserGroup(cctx.Context, GetUserGroupParams{
		GroupId: c.GroupId,
		Cloud:   cloudClient.CloudService(),
		Printer: cctx.Printer,
	})
}

func (c *CloudUserGroupListCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return ListUserGroups(cctx.Context, ListUserGroupsParams{
		PageSize:                int32(c.PageSize),
		PageToken:               c.PageToken,
		Namespace:               c.Namespace,
		DisplayName:             c.DisplayName,
		GoogleGroupEmailAddress: c.GoogleGroupEmailAddress,
		ScimGroupIdpId:          c.ScimGroupIdpId,
		Cloud:                   cloudClient.CloudService(),
		Printer:                 cctx.Printer,
	})
}
