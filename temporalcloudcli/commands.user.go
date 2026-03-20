package temporalcloudcli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

// AIDEV-NOTE: accountRoleNames and namespacePermissionNames are the canonical
// string-to-enum mappings used by invite (and future update) commands.
// The keys are the exact strings accepted by --account-role and --namespace-access flags.
var (
	accountRoleNames = map[string]identityv1.AccountAccess_Role{
		"owner":         identityv1.AccountAccess_ROLE_OWNER,
		"admin":         identityv1.AccountAccess_ROLE_ADMIN,
		"developer":     identityv1.AccountAccess_ROLE_DEVELOPER,
		"finance-admin": identityv1.AccountAccess_ROLE_FINANCE_ADMIN,
		"read":          identityv1.AccountAccess_ROLE_READ,
		"metrics-read":  identityv1.AccountAccess_ROLE_METRICS_READ,
	}

	namespacePermissionNames = map[string]identityv1.NamespaceAccess_Permission{
		"admin": identityv1.NamespaceAccess_PERMISSION_ADMIN,
		"write": identityv1.NamespaceAccess_PERMISSION_WRITE,
		"read":  identityv1.NamespaceAccess_PERMISSION_READ,
	}
)

type (
	SetNamespacePermissionsParams struct {
		UserIdentification UserIdentificationOptions
		// NamespaceAccesses lists changes to apply in "namespace=permission" format.
		// Each entry adds or overwrites a namespace; an empty permission (e.g. "testns=") removes it.
		NamespaceAccesses []string
		ResourceVersion   string
		AsyncOperationID  string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	SetAccountRoleParams struct {
		UserIdentification UserIdentificationOptions
		AccountRole        string
		ResourceVersion    string
		AsyncOperationID   string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	DeleteUserParams struct {
		UserIdentification UserIdentificationOptions
		ResourceVersion    string
		AsyncOperationID   string

		Cloud            cloudservice.CloudServiceClient
		OperationHandler AsyncOperationHandler
	}

	EditUserParams struct {
		UserIdentification UserIdentificationOptions
		ResourceVersion    string
		AsyncOperationID   string
		VerboseDiff        bool

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
		// RunEditor opens the existing spec in an editor and writes the result into target.
		// Injected so the function is unit-testable without a real editor process.
		RunEditor func(existing, target proto.Message) error
	}

	ApplyUserParams struct {
		Spec             *identityv1.UserSpec
		ResourceVersion  string
		AsyncOperationID string
		VerboseDiff      bool

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	GetUserParams struct {
		UserIdentification UserIdentificationOptions

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
	}

	InviteUserParams struct {
		Email             string
		AccountAccess     *identityv1.AccountAccess
		NamespaceAccesses map[string]*identityv1.NamespaceAccess
		AsyncOperationID  string

		Cloud            cloudservice.CloudServiceClient
		OperationHandler AsyncOperationHandler
	}

	ListUsersParams struct {
		PageSize  int32
		PageToken string
		Email     string
		Namespace string

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
	}
)

func SetNamespacePermissions(ctx context.Context, params SetNamespacePermissionsParams) error {
	if err := validateUserIdentification(params.UserIdentification); err != nil {
		return err
	}

	// Validate changes before making any API calls.
	if _, err := applyNamespaceAccessChanges(nil, params.NamespaceAccesses); err != nil {
		return err
	}

	user, err := resolveUser(ctx, params.Cloud, params.UserIdentification)
	if err != nil {
		return err
	}

	newSpec := proto.Clone(user.Spec).(*identityv1.UserSpec)
	if newSpec.Access == nil {
		newSpec.Access = &identityv1.Access{}
	}

	namespaceAccesses, err := applyNamespaceAccessChanges(newSpec.Access.NamespaceAccesses, params.NamespaceAccesses)
	if err != nil {
		return err
	}
	newSpec.Access.NamespaceAccesses = namespaceAccesses

	if err := params.Prompter.PromptApply(user.Spec, newSpec, false); err != nil {
		return err
	}

	rv := user.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}

	updateUser := wrapUpdateOperation(params.Cloud.UpdateUser, params.OperationHandler, user.Id)
	return updateUser(ctx, &cloudservice.UpdateUserRequest{
		UserId:           user.Id,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func SetAccountRole(ctx context.Context, params SetAccountRoleParams) error {
	if err := validateUserIdentification(params.UserIdentification); err != nil {
		return err
	}
	accountAccess, err := parseAccountRole(params.AccountRole)
	if err != nil {
		return err
	}

	user, err := resolveUser(ctx, params.Cloud, params.UserIdentification)
	if err != nil {
		return err
	}

	newSpec := proto.Clone(user.Spec).(*identityv1.UserSpec)
	if newSpec.Access == nil {
		newSpec.Access = &identityv1.Access{}
	}
	newSpec.Access.AccountAccess = accountAccess

	if err := params.Prompter.PromptApply(user.Spec, newSpec, false); err != nil {
		return err
	}

	rv := user.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}

	updateUser := wrapUpdateOperation(params.Cloud.UpdateUser, params.OperationHandler, user.Id)
	return updateUser(ctx, &cloudservice.UpdateUserRequest{
		UserId:           user.Id,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func DeleteUser(ctx context.Context, params DeleteUserParams) error {
	if err := validateUserIdentification(params.UserIdentification); err != nil {
		return err
	}

	user, err := resolveUser(ctx, params.Cloud, params.UserIdentification)
	if err != nil {
		return err
	}

	rv := user.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}

	deleteUser := wrapDeleteOperation(params.Cloud.DeleteUser, params.OperationHandler, user.Id)
	return deleteUser(ctx, &cloudservice.DeleteUserRequest{
		UserId:           user.Id,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func EditUser(ctx context.Context, params EditUserParams) error {
	if err := validateUserIdentification(params.UserIdentification); err != nil {
		return err
	}

	user, err := resolveUser(ctx, params.Cloud, params.UserIdentification)
	if err != nil {
		return err
	}

	newSpec := &identityv1.UserSpec{}
	if err := params.RunEditor(user.Spec, newSpec); err != nil {
		return err
	}

	if err := params.Prompter.PromptApply(user.Spec, newSpec, params.VerboseDiff); err != nil {
		return err
	}

	rv := user.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}

	updateUser := wrapUpdateOperation(params.Cloud.UpdateUser, params.OperationHandler, user.Id)
	return updateUser(ctx, &cloudservice.UpdateUserRequest{
		UserId:           user.Id,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

// resolveUser fetches a user by ID or by email (via GetUsers).
// Exactly one of userID or email should be non-empty.
func resolveUser(ctx context.Context, cloud cloudservice.CloudServiceClient, identification UserIdentificationOptions) (*identityv1.User, error) {
	if identification.UserId != "" {
		res, err := cloud.GetUser(ctx, &cloudservice.GetUserRequest{UserId: identification.UserId})
		if err != nil {
			return nil, err
		}
		return res.User, nil
	}
	res, err := cloud.GetUsers(ctx, &cloudservice.GetUsersRequest{Email: identification.UserEmail})
	if err != nil {
		return nil, err
	}
	if len(res.Users) == 0 {
		return nil, fmt.Errorf("no user found with email %q", identification.UserEmail)
	}
	return res.Users[0], nil
}

func ApplyUser(ctx context.Context, params ApplyUserParams) error {
	// Look up existing user by email.
	res, err := params.Cloud.GetUsers(ctx, &cloudservice.GetUsersRequest{Email: params.Spec.Email})
	if err != nil {
		return err
	}

	var existingSpec *identityv1.UserSpec
	var existingUserID string
	rv := params.ResourceVersion

	if len(res.Users) > 0 {
		existing := res.Users[0]
		existingSpec = existing.Spec
		existingUserID = existing.Id
		if rv == "" {
			rv = existing.ResourceVersion
		}
	}

	if err := params.Prompter.PromptApply(existingSpec, params.Spec, params.VerboseDiff); err != nil {
		return err
	}

	if existingUserID != "" {
		updateUser := wrapUpdateOperation(params.Cloud.UpdateUser, params.OperationHandler, existingUserID)
		return updateUser(ctx, &cloudservice.UpdateUserRequest{
			UserId:           existingUserID,
			Spec:             params.Spec,
			ResourceVersion:  rv,
			AsyncOperationId: params.AsyncOperationID,
		})
	}

	createUser := wrapCreateOperation(params.Cloud.CreateUser, params.OperationHandler, func(res *cloudservice.CreateUserResponse) string { return res.GetUserId() })
	return createUser(ctx, &cloudservice.CreateUserRequest{
		Spec:             params.Spec,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func GetUser(ctx context.Context, params GetUserParams) error {
	if err := validateUserIdentification(params.UserIdentification); err != nil {
		return err
	}
	user, err := resolveUser(ctx, params.Cloud, params.UserIdentification)
	if err != nil {
		return err
	}
	return params.Printer.PrintResource(user, printer.PrintResourceOptions{})
}

func InviteUser(ctx context.Context, params InviteUserParams) error {
	spec := &identityv1.UserSpec{
		Email: params.Email,
		Access: &identityv1.Access{
			AccountAccess:     params.AccountAccess,
			NamespaceAccesses: params.NamespaceAccesses,
		},
	}

	createUser := wrapCreateOperation(params.Cloud.CreateUser, params.OperationHandler, func(res *cloudservice.CreateUserResponse) string { return res.GetUserId() })
	return createUser(ctx, &cloudservice.CreateUserRequest{
		Spec:             spec,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func ListUsers(ctx context.Context, params ListUsersParams) error {
	res, err := params.Cloud.GetUsers(ctx, &cloudservice.GetUsersRequest{
		PageSize:  params.PageSize,
		PageToken: params.PageToken,
		Email:     params.Email,
		Namespace: params.Namespace,
	})
	if err != nil {
		return err
	}

	return params.Printer.PrintResourceList(
		struct {
			Users         []*identityv1.User
			NextPageToken string
		}{
			Users:         res.Users,
			NextPageToken: res.NextPageToken,
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Id", "State", "CreatedTime"},
			SpecFields: []string{"Email"},
		},
		printer.TableOptions{},
	)
}

func (c *CloudUserSetNamespacePermissionsCommand) run(cctx *CommandContext, _ []string) error {
	if err := validateUserIdentification(c.UserIdentificationOptions); err != nil {
		return err
	}

	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	return SetNamespacePermissions(cctx.Context, SetNamespacePermissionsParams{
		UserIdentification: c.UserIdentificationOptions,
		NamespaceAccesses:  c.NamespaceAccess,
		ResourceVersion:    c.ResourceVersion,
		AsyncOperationID:   c.AsyncOperationId,
		Cloud:              cloudClient.CloudService(),
		Prompter:           newPrompter(cctx),
		OperationHandler:   NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudUserSetAccountRoleCommand) run(cctx *CommandContext, _ []string) error {
	if err := validateUserIdentification(c.UserIdentificationOptions); err != nil {
		return err
	}

	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	return SetAccountRole(cctx.Context, SetAccountRoleParams{
		UserIdentification: c.UserIdentificationOptions,
		AccountRole:        c.AccountRole,
		ResourceVersion:    c.ResourceVersion,
		AsyncOperationID:   c.AsyncOperationId,
		Cloud:              cloudClient.CloudService(),
		Prompter:           newPrompter(cctx),
		OperationHandler:   NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudUserDeleteCommand) run(cctx *CommandContext, _ []string) error {
	if err := validateUserIdentification(c.UserIdentificationOptions); err != nil {
		return err
	}

	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	yes, err := cctx.promptYes("Delete (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting delete.")
	}

	return DeleteUser(cctx.Context, DeleteUserParams{
		UserIdentification: c.UserIdentificationOptions,
		ResourceVersion:    c.ResourceVersion,
		AsyncOperationID:   c.AsyncOperationId,
		Cloud:              cloudClient.CloudService(),
		OperationHandler:   NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudUserEditCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return EditUser(cctx.Context, EditUserParams{
		UserIdentification: c.UserIdentificationOptions,
		ResourceVersion:    c.ResourceVersion,
		AsyncOperationID:   c.AsyncOperationId,
		VerboseDiff:        c.VerboseDiff,
		Cloud:              cloudClient.CloudService(),
		Prompter:           newPrompter(cctx),
		OperationHandler:   NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
		RunEditor:          runEditorForJSONEditForProtos,
	})
}

func (c *CloudUserApplyCommand) run(cctx *CommandContext, _ []string) error {
	specData, err := loadJSONSpec(c.Spec)
	if err != nil {
		return err
	}

	spec := &identityv1.UserSpec{}
	if err := cctx.UnmarshalProtoJSON(specData, spec); err != nil {
		return fmt.Errorf("failed to parse JSON spec: %w", err)
	}

	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	return ApplyUser(cctx.Context, ApplyUserParams{
		Spec:             spec,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		VerboseDiff:      c.VerboseDiff,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudUserGetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return GetUser(cctx.Context, GetUserParams{
		UserIdentification: c.UserIdentificationOptions,
		Cloud:              cloudClient.CloudService(),
		Printer:            cctx.Printer,
	})
}

func (c *CloudUserInviteCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	accountAccess, err := parseAccountRole(c.AccountRole)
	if err != nil {
		return err
	}

	namespaceAccesses, err := parseNamespaceAccesses(c.NamespaceAccess)
	if err != nil {
		return err
	}

	yes, err := cctx.promptYes("Invite (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting invite.")
	}

	return InviteUser(cctx.Context, InviteUserParams{
		Email:             c.Email,
		AccountAccess:     accountAccess,
		NamespaceAccesses: namespaceAccesses,
		AsyncOperationID:  c.AsyncOperationId,
		Cloud:             cloudClient.CloudService(),
		OperationHandler:  NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudUserListCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return ListUsers(cctx.Context, ListUsersParams{
		PageSize:  int32(c.PageSize),
		PageToken: c.PageToken,
		Email:     c.Email,
		Namespace: c.Namespace,
		Cloud:     cloudClient.CloudService(),
		Printer:   cctx.Printer,
	})
}

// parseAccountRole converts the --account-role flag string to an AccountAccess proto.
// Returns nil (no account access) when role is empty.
func parseAccountRole(role string) (*identityv1.AccountAccess, error) {
	if role == "" {
		return nil, nil
	}
	r, ok := accountRoleNames[role]
	if !ok {
		return nil, fmt.Errorf("invalid account role %q: must be one of owner, admin, developer, finance-admin, read, metrics-read", role)
	}
	return &identityv1.AccountAccess{Role: r}, nil
}

// parseNamespaceAccesses converts --namespace-access flag values (format: "namespace=permission")
// to a NamespaceAccesses map. Returns nil when no accesses are provided.
func parseNamespaceAccesses(accesses []string) (map[string]*identityv1.NamespaceAccess, error) {
	if len(accesses) == 0 {
		return nil, nil
	}
	result := make(map[string]*identityv1.NamespaceAccess, len(accesses))
	for _, a := range accesses {
		ns, perm, ok := strings.Cut(a, "=")
		if !ok {
			return nil, fmt.Errorf("invalid namespace-access %q: must be in the format 'namespace=permission'", a)
		}
		p, ok := namespacePermissionNames[perm]
		if !ok {
			return nil, fmt.Errorf("invalid permission %q in namespace-access %q: must be one of admin, write, read", perm, a)
		}
		result[ns] = &identityv1.NamespaceAccess{Permission: p}
	}
	return result, nil
}

// applyNamespaceAccessChanges merges a set of namespace access changes into an existing map.
// Each change is in "namespace=permission" format. An empty permission (e.g. "testns=") removes
// that namespace; a non-empty permission adds or overwrites it. Returns nil when the result is empty.
func applyNamespaceAccessChanges(existing map[string]*identityv1.NamespaceAccess, changes []string) (map[string]*identityv1.NamespaceAccess, error) {
	result := make(map[string]*identityv1.NamespaceAccess, len(existing))
	for k, v := range existing {
		result[k] = v
	}
	for _, a := range changes {
		ns, perm, ok := strings.Cut(a, "=")
		if !ok {
			return nil, fmt.Errorf("invalid namespace-access %q: must be in the format 'namespace=permission'", a)
		}
		if perm == "" {
			delete(result, ns)
		} else {
			p, ok := namespacePermissionNames[perm]
			if !ok {
				return nil, fmt.Errorf("invalid permission %q in namespace-access %q: must be one of admin, write, read", perm, a)
			}
			result[ns] = &identityv1.NamespaceAccess{Permission: p}
		}
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

func validateUserIdentification(identification UserIdentificationOptions) error {
	if identification.UserId == "" && identification.UserEmail == "" {
		return errors.New("must provide either --user-id or --user-email")
	}
	if identification.UserId != "" && identification.UserEmail != "" {
		return errors.New("cannot provide both --user-id and --user-email")
	}
	return nil
}
