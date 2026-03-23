package temporalcloudcli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type (
	ListApiKeysParams struct {
		UserId           string
		UserEmail        string
		ServiceAccountId string
		PageSize         int32
		PageToken        string

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
	}

	GetApiKeyParams struct {
		KeyId string

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
	}

	CreateApiKeyForServiceAccountParams struct {
		ServiceAccountId string
		DisplayName      string
		Description      string
		ExpiryTime       string
		ExpiryDuration   string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Printer          *printer.Printer
		OperationHandler AsyncOperationHandler
	}

	CreateApiKeyForMeParams struct {
		DisplayName      string
		Description      string
		ExpiryTime       string
		ExpiryDuration   string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Printer          *printer.Printer
		OperationHandler AsyncOperationHandler
	}

	DeleteApiKeyParams struct {
		KeyId            string
		ResourceVersion  string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		OperationHandler AsyncOperationHandler
	}

	UpdateApiKeyParams struct {
		KeyId            string
		DisplayName      *string // nil = not changed
		Description      *string // nil = not changed
		Disabled         *bool   // nil = not changed
		ResourceVersion  string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	EditApiKeyParams struct {
		KeyId            string
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

	DisableApiKeyParams struct {
		KeyId            string
		ResourceVersion  string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	EnableApiKeyParams struct {
		KeyId            string
		ResourceVersion  string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}
)

func ListApiKeys(ctx context.Context, params ListApiKeysParams) error {
	filterCount := 0
	if params.UserId != "" {
		filterCount++
	}
	if params.UserEmail != "" {
		filterCount++
	}
	if params.ServiceAccountId != "" {
		filterCount++
	}
	if filterCount > 1 {
		return errors.New("only one of --user-id, --user-email, --service-account-id may be specified")
	}

	var ownerID string
	var ownerType identityv1.OwnerType
	switch {
	case params.UserId != "":
		ownerID = params.UserId
		ownerType = identityv1.OwnerType_OWNER_TYPE_USER
	case params.UserEmail != "":
		usersRes, err := params.Cloud.GetUsers(ctx, &cloudservice.GetUsersRequest{Email: params.UserEmail})
		if err != nil {
			return err
		}
		if len(usersRes.Users) == 0 {
			return fmt.Errorf("no user found with email %q", params.UserEmail)
		}
		ownerID = usersRes.Users[0].Id
		ownerType = identityv1.OwnerType_OWNER_TYPE_USER
	case params.ServiceAccountId != "":
		ownerID = params.ServiceAccountId
		ownerType = identityv1.OwnerType_OWNER_TYPE_SERVICE_ACCOUNT
	}

	res, err := params.Cloud.GetApiKeys(ctx, &cloudservice.GetApiKeysRequest{
		OwnerId:   ownerID,
		OwnerType: ownerType,
		PageSize:  params.PageSize,
		PageToken: params.PageToken,
	})
	if err != nil {
		return err
	}

	return params.Printer.PrintResourceList(
		struct {
			ApiKeys       []*identityv1.ApiKey
			NextPageToken string
		}{
			ApiKeys:       res.ApiKeys,
			NextPageToken: res.NextPageToken,
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Id", "State", "CreatedTime"},
			SpecFields: []string{"DisplayName", "OwnerId", "OwnerType", "Disabled"},
		},
		printer.TableOptions{},
	)
}

func GetApiKey(ctx context.Context, params GetApiKeyParams) error {
	res, err := params.Cloud.GetApiKey(ctx, &cloudservice.GetApiKeyRequest{KeyId: params.KeyId})
	if err != nil {
		return err
	}
	return params.Printer.PrintResource(res.ApiKey, printer.PrintResourceOptions{})
}

// resolveApiKeyExpiry resolves the expiry timestamp from --expiry-time or --expiry-duration.
// The two flags are mutually exclusive. Returns nil if neither is set.
func resolveApiKeyExpiry(expiryTime, expiryDuration string) (*timestamppb.Timestamp, error) {
	if expiryTime != "" && expiryDuration != "" {
		return nil, errors.New("--expiry-time and --expiry-duration are mutually exclusive")
	}
	if expiryTime != "" {
		t, err := time.Parse(time.RFC3339, expiryTime)
		if err != nil {
			return nil, fmt.Errorf("invalid --expiry-time %q: must be RFC3339 format (e.g. 2025-12-31T00:00:00Z): %w", expiryTime, err)
		}
		return timestamppb.New(t), nil
	}
	if expiryDuration != "" {
		d, err := parseExpiryDuration(expiryDuration)
		if err != nil {
			return nil, err
		}
		return timestamppb.New(time.Now().Add(d)), nil
	}
	return nil, nil
}

// parseExpiryDuration parses a duration string, extending Go's standard format with
// a "d" suffix for days (e.g. "30d" = 720h).
func parseExpiryDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil || days <= 0 {
			return 0, fmt.Errorf("invalid --expiry-duration %q: day count must be a positive integer (e.g. 30d)", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid --expiry-duration %q: use Go duration format or days (e.g. 30d, 24h, 90m): %w", s, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("invalid --expiry-duration %q: duration must be positive", s)
	}
	return d, nil
}

func CreateApiKeyForServiceAccount(ctx context.Context, params CreateApiKeyForServiceAccountParams) error {
	expiry, err := resolveApiKeyExpiry(params.ExpiryTime, params.ExpiryDuration)
	if err != nil {
		return err
	}
	res, err := params.Cloud.CreateApiKey(ctx, &cloudservice.CreateApiKeyRequest{
		Spec: &identityv1.ApiKeySpec{
			OwnerId:     params.ServiceAccountId,
			OwnerType:   identityv1.OwnerType_OWNER_TYPE_SERVICE_ACCOUNT,
			DisplayName: params.DisplayName,
			Description: params.Description,
			ExpiryTime:  expiry,
		},
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return params.OperationHandler.HandleErr(err)
	}
	// AIDEV-NOTE: The token is only returned on creation and cannot be retrieved again.
	// Print it immediately before handling the async operation.
	_ = params.Printer.PrintStructured(struct {
		KeyId string `json:"keyId"`
		Token string `json:"token"`
	}{KeyId: res.KeyId, Token: res.Token}, printer.StructuredOptions{})
	return params.OperationHandler.Handle(res.GetAsyncOperation())
}

func CreateApiKeyForMe(ctx context.Context, params CreateApiKeyForMeParams) error {
	identityRes, err := params.Cloud.GetCurrentIdentity(ctx, &cloudservice.GetCurrentIdentityRequest{})
	if err != nil {
		return err
	}
	user := identityRes.GetUser()
	if user == nil {
		return errors.New("current principal is not a user: use create-for-service-account for service accounts")
	}

	expiry, err := resolveApiKeyExpiry(params.ExpiryTime, params.ExpiryDuration)
	if err != nil {
		return err
	}
	res, err := params.Cloud.CreateApiKey(ctx, &cloudservice.CreateApiKeyRequest{
		Spec: &identityv1.ApiKeySpec{
			OwnerId:     user.Id,
			OwnerType:   identityv1.OwnerType_OWNER_TYPE_USER,
			DisplayName: params.DisplayName,
			Description: params.Description,
			ExpiryTime:  expiry,
		},
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return params.OperationHandler.HandleErr(err)
	}
	// AIDEV-NOTE: The token is only returned on creation and cannot be retrieved again.
	// Print it immediately before handling the async operation.
	_ = params.Printer.PrintStructured(struct {
		KeyId string `json:"keyId"`
		Token string `json:"token"`
	}{KeyId: res.KeyId, Token: res.Token}, printer.StructuredOptions{})
	return params.OperationHandler.Handle(res.GetAsyncOperation())
}

func DeleteApiKey(ctx context.Context, params DeleteApiKeyParams) error {
	res, err := params.Cloud.GetApiKey(ctx, &cloudservice.GetApiKeyRequest{KeyId: params.KeyId})
	if err != nil {
		return err
	}
	rv := res.ApiKey.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}
	deleteApiKey := runAsyncOperation(params.Cloud.DeleteApiKey, params.OperationHandler)
	return deleteApiKey(ctx, &cloudservice.DeleteApiKeyRequest{
		KeyId:            params.KeyId,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func UpdateApiKey(ctx context.Context, params UpdateApiKeyParams) error {
	res, err := params.Cloud.GetApiKey(ctx, &cloudservice.GetApiKeyRequest{KeyId: params.KeyId})
	if err != nil {
		return err
	}
	key := res.ApiKey
	newSpec := proto.Clone(key.Spec).(*identityv1.ApiKeySpec)

	if params.DisplayName != nil {
		newSpec.DisplayName = *params.DisplayName
	}
	if params.Description != nil {
		newSpec.Description = *params.Description
	}
	if params.Disabled != nil {
		newSpec.Disabled = *params.Disabled
	}

	if err := params.Prompter.PromptApply(key.Spec, newSpec, false); err != nil {
		return err
	}

	rv := key.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}

	updateApiKey := runAsyncOperation(params.Cloud.UpdateApiKey, params.OperationHandler)
	return updateApiKey(ctx, &cloudservice.UpdateApiKeyRequest{
		KeyId:            params.KeyId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func EditApiKey(ctx context.Context, params EditApiKeyParams) error {
	res, err := params.Cloud.GetApiKey(ctx, &cloudservice.GetApiKeyRequest{KeyId: params.KeyId})
	if err != nil {
		return err
	}
	key := res.ApiKey
	newSpec := &identityv1.ApiKeySpec{}
	if err := params.RunEditor(key.Spec, newSpec); err != nil {
		return err
	}

	if err := params.Prompter.PromptApply(key.Spec, newSpec, params.VerboseDiff); err != nil {
		return err
	}

	rv := key.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}

	updateApiKey := runAsyncOperation(params.Cloud.UpdateApiKey, params.OperationHandler)
	return updateApiKey(ctx, &cloudservice.UpdateApiKeyRequest{
		KeyId:            params.KeyId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func DisableApiKey(ctx context.Context, params DisableApiKeyParams) error {
	res, err := params.Cloud.GetApiKey(ctx, &cloudservice.GetApiKeyRequest{KeyId: params.KeyId})
	if err != nil {
		return err
	}
	key := res.ApiKey
	newSpec := proto.Clone(key.Spec).(*identityv1.ApiKeySpec)
	newSpec.Disabled = true

	if err := params.Prompter.PromptApply(key.Spec, newSpec, false); err != nil {
		return err
	}

	rv := key.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}

	updateApiKey := runAsyncOperation(params.Cloud.UpdateApiKey, params.OperationHandler)
	return updateApiKey(ctx, &cloudservice.UpdateApiKeyRequest{
		KeyId:            params.KeyId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func EnableApiKey(ctx context.Context, params EnableApiKeyParams) error {
	res, err := params.Cloud.GetApiKey(ctx, &cloudservice.GetApiKeyRequest{KeyId: params.KeyId})
	if err != nil {
		return err
	}
	key := res.ApiKey
	newSpec := proto.Clone(key.Spec).(*identityv1.ApiKeySpec)
	newSpec.Disabled = false

	if err := params.Prompter.PromptApply(key.Spec, newSpec, false); err != nil {
		return err
	}

	rv := key.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}

	updateApiKey := runAsyncOperation(params.Cloud.UpdateApiKey, params.OperationHandler)
	return updateApiKey(ctx, &cloudservice.UpdateApiKeyRequest{
		KeyId:            params.KeyId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func (c *CloudApikeyListCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return ListApiKeys(cctx.Context, ListApiKeysParams{
		UserId:           c.UserId,
		UserEmail:        c.UserEmail,
		ServiceAccountId: c.ServiceAccountId,
		PageSize:         int32(c.PageSize),
		PageToken:        c.PageToken,
		Cloud:            cloudClient.CloudService(),
		Printer:          cctx.Printer,
	})
}

func (c *CloudApikeyGetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return GetApiKey(cctx.Context, GetApiKeyParams{
		KeyId:   c.KeyId,
		Cloud:   cloudClient.CloudService(),
		Printer: cctx.Printer,
	})
}

func (c *CloudApikeyCreateForServiceAccountCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return CreateApiKeyForServiceAccount(cctx.Context, CreateApiKeyForServiceAccountParams{
		ServiceAccountId: c.ServiceAccountId,
		DisplayName:      c.DisplayName,
		Description:      c.Description,
		ExpiryTime:       c.ExpiryTime,
		ExpiryDuration:   c.ExpiryDuration,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Printer:          cctx.Printer,
		OperationHandler: NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.ServiceAccountId, c.ClientOptions),
	})
}

func (c *CloudApikeyCreateForMeCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return CreateApiKeyForMe(cctx.Context, CreateApiKeyForMeParams{
		DisplayName:      c.DisplayName,
		Description:      c.Description,
		ExpiryTime:       c.ExpiryTime,
		ExpiryDuration:   c.ExpiryDuration,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Printer:          cctx.Printer,
		OperationHandler: NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, "", c.ClientOptions),
	})
}

func (c *CloudApikeyDeleteCommand) run(cctx *CommandContext, _ []string) error {
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
	return DeleteApiKey(cctx.Context, DeleteApiKeyParams{
		KeyId:            c.KeyId,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		OperationHandler: NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.KeyId, c.ClientOptions),
	})
}

func (c *CloudApikeyUpdateCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	// Only pass fields that were explicitly set — use Changed() to distinguish
	// "not provided" from the zero value (especially for --disabled=false).
	var displayName *string
	if c.Command.Flags().Changed("display-name") {
		displayName = &c.DisplayName
	}
	var description *string
	if c.Command.Flags().Changed("description") {
		description = &c.Description
	}
	var disabled *bool
	if c.Command.Flags().Changed("disabled") {
		disabled = &c.Disabled
	}
	return UpdateApiKey(cctx.Context, UpdateApiKeyParams{
		KeyId:            c.KeyId,
		DisplayName:      displayName,
		Description:      description,
		Disabled:         disabled,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.KeyId, c.ClientOptions),
	})
}

func (c *CloudApikeyEditCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return EditApiKey(cctx.Context, EditApiKeyParams{
		KeyId:            c.KeyId,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		VerboseDiff:      c.VerboseDiff,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.KeyId, c.ClientOptions),
		RunEditor:        runEditorForJSONEditForProtos,
	})
}

func (c *CloudApikeyDisableCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return DisableApiKey(cctx.Context, DisableApiKeyParams{
		KeyId:            c.KeyId,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.KeyId, c.ClientOptions),
	})
}

func (c *CloudApikeyEnableCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return EnableApiKey(cctx.Context, EnableApiKeyParams{
		KeyId:            c.KeyId,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.KeyId, c.ClientOptions),
	})
}
