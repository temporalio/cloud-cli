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

func (c *CloudUserGetCommand) run(cctx *CommandContext, _ []string) error {
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
	return cctx.Printer.PrintResource(user, printer.PrintResourceOptions{})
}

func (c *CloudUserListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetUsers(cctx, &cloudservice.GetUsersRequest{
		PageSize:  int32(c.PageSize),
		PageToken: c.PageToken,
		Email:     c.Email,
		Namespace: c.Namespace,
	})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResourceList(
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

func (c *CloudUserInviteCommand) run(cctx *CommandContext, _ []string) error {
	accountAccess, err := parseAccountRole(c.AccountRole)
	if err != nil {
		return err
	}
	namespaceAccesses, err := parseNamespaceAccesses(c.NamespaceAccess)
	if err != nil {
		return err
	}
	customRoleProvided := c.Command.Flags().Changed("custom-role")
	if customRoleProvided {
		if accountAccess == nil {
			return errors.New("--custom-role requires --account-role; a principal must have a built-in role")
		}
		accountAccess.CustomRoles = dedupeStrings(c.CustomRole)
	}
	yes, err := cctx.GetPrompter().PromptYes("Invite")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting invite.")
	}
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	resp, err := client.CreateUser(cctx, &cloudservice.CreateUserRequest{
		Spec: &identityv1.UserSpec{
			Email: c.Email,
			Access: &identityv1.Access{
				AccountAccess:     accountAccess,
				NamespaceAccesses: namespaceAccesses,
			},
		},
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudUserDeleteCommand) run(cctx *CommandContext, _ []string) error {
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
	yes, err := cctx.GetPrompter().PromptYes("Delete")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting delete.")
	}
	rv := user.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.DeleteUser(cctx, &cloudservice.DeleteUserRequest{
		UserId:           user.Id,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleDeleteOperation(cctx, resp, err)
}

func (c *CloudUserSetNamespacePermissionsCommand) run(cctx *CommandContext, _ []string) error {
	if err := validateUserIdentification(c.UserIdentificationOptions); err != nil {
		return err
	}
	// Validate namespace access changes before any API calls.
	if _, err := applyNamespaceAccessChanges(nil, c.NamespaceAccess); err != nil {
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
	newSpec := proto.Clone(user.Spec).(*identityv1.UserSpec)
	if newSpec.Access == nil {
		newSpec.Access = &identityv1.Access{}
	}
	namespaceAccesses, err := applyNamespaceAccessChanges(newSpec.Access.NamespaceAccesses, c.NamespaceAccess)
	if err != nil {
		return err
	}
	newSpec.Access.NamespaceAccesses = namespaceAccesses

	yes, err := cctx.GetPrompter().PromptYes("Set namespace permissions")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting set.")
	}
	rv := user.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateUser(cctx, &cloudservice.UpdateUserRequest{
		UserId:           user.Id,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudUserSetAccountRoleCommand) run(cctx *CommandContext, _ []string) error {
	if err := validateUserIdentification(c.UserIdentificationOptions); err != nil {
		return err
	}
	accountAccess, err := parseAccountRole(c.AccountRole)
	if err != nil {
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
	newSpec := proto.Clone(user.Spec).(*identityv1.UserSpec)
	if newSpec.Access == nil {
		newSpec.Access = &identityv1.Access{}
	}
	if newSpec.Access.AccountAccess != nil {
		accountAccess.CustomRoles = newSpec.Access.AccountAccess.CustomRoles
	}
	newSpec.Access.AccountAccess = accountAccess

	yes, err := cctx.GetPrompter().PromptYes("Set account role")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting set.")
	}
	rv := user.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateUser(cctx, &cloudservice.UpdateUserRequest{
		UserId:           user.Id,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudUserEditCommand) run(cctx *CommandContext, _ []string) error {
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
	edited, err := cctx.GetEditor().EditProto(user.Spec)
	if err != nil {
		return err
	}
	newSpec := edited.(*identityv1.UserSpec)

	yes, err := cctx.GetPrompter().PromptApply(user.Spec, newSpec, c.VerboseDiff)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting edit.")
	}
	rv := user.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateUser(cctx, &cloudservice.UpdateUserRequest{
		UserId:           user.Id,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
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
	if spec.Email == "" {
		return errors.New("spec must include an email address")
	}
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetUsers(cctx, &cloudservice.GetUsersRequest{Email: spec.Email})
	if err != nil {
		return err
	}
	var existingSpec *identityv1.UserSpec
	var existingUserID string
	rv := c.ResourceVersion
	if len(res.Users) > 0 {
		existing := res.Users[0]
		existingSpec = existing.Spec
		existingUserID = existing.Id
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
	if existingUserID != "" {
		resp, err := client.UpdateUser(cctx, &cloudservice.UpdateUserRequest{
			UserId:           existingUserID,
			Spec:             spec,
			ResourceVersion:  rv,
			AsyncOperationId: c.AsyncOperationId,
		})
		return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
	}
	resp, err := client.CreateUser(cctx, &cloudservice.CreateUserRequest{
		Spec:             spec,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

// resolveUser fetches a user by ID or by email (via GetUsers).
// Exactly one of UserId or UserEmail should be non-empty.
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

// applyCustomRoleChanges returns the new CustomRoles slice given the existing
// list and flag values. When neither --custom-role nor --clear-custom-roles
// is set, the existing list is returned unchanged. When --custom-role is set,
// the list is replaced (de-duplicated). When --clear-custom-roles is set, the
// list is cleared.
//
// AIDEV-NOTE: customRoleProvided / clearProvided come from
// cobra.Command.Flags().Changed(...) so the caller can distinguish "leave
// untouched" (no flag) from "clear" (--clear-custom-roles).
func applyCustomRoleChanges(
	existing []string,
	customRoles []string,
	customRoleProvided bool,
	clearProvided bool,
) ([]string, error) {
	if customRoleProvided && clearProvided {
		return nil, errors.New("--custom-role and --clear-custom-roles are mutually exclusive")
	}
	if clearProvided {
		return nil, nil
	}
	if !customRoleProvided {
		return existing, nil
	}
	return dedupeStrings(customRoles), nil
}

// dedupeStrings preserves first-occurrence order so test output is deterministic.
func dedupeStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
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

func (c *CloudUserSetCustomRolesCommand) run(cctx *CommandContext, _ []string) error {
	if err := validateUserIdentification(c.UserIdentificationOptions); err != nil {
		return err
	}
	customRoleProvided := c.Command.Flags().Changed("custom-role")
	clearProvided := c.Command.Flags().Changed("clear-custom-roles")
	if !customRoleProvided && !clearProvided {
		return errors.New("must provide --custom-role or --clear-custom-roles")
	}
	if _, err := applyCustomRoleChanges(nil, c.CustomRole, customRoleProvided, clearProvided); err != nil {
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
	newSpec := proto.Clone(user.Spec).(*identityv1.UserSpec)
	if newSpec.Access == nil || newSpec.Access.AccountAccess == nil {
		return errors.New("user has no account access; assign a built-in role with set-account-role first")
	}
	roles, err := applyCustomRoleChanges(
		newSpec.Access.AccountAccess.CustomRoles,
		c.CustomRole, customRoleProvided, clearProvided,
	)
	if err != nil {
		return err
	}
	newSpec.Access.AccountAccess.CustomRoles = roles

	yes, err := cctx.GetPrompter().PromptApply(user.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting set.")
	}
	rv := user.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateUser(cctx, &cloudservice.UpdateUserRequest{
		UserId:           user.Id,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
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
