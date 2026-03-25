package temporalcloudcli

import (
	"errors"
	"fmt"
	"time"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/temporalio/cli/cliext"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

// resolveApiKeyExpiry resolves the expiry timestamp from --expiry-time or --expiry-duration.
// The two flags are mutually exclusive. Returns nil if neither is set.
func resolveApiKeyExpiry(expiryTime cliext.FlagTimestamp, expiryDuration cliext.FlagDuration) (*timestamppb.Timestamp, error) {
	if !time.Time(expiryTime).IsZero() && expiryDuration != 0 {
		return nil, errors.New("--expiry-time and --expiry-duration are mutually exclusive")
	}
	if !time.Time(expiryTime).IsZero() {
		return timestamppb.New(time.Time(expiryTime)), nil
	}
	if expiryDuration != 0 {
		return timestamppb.New(time.Now().Add(time.Duration(expiryDuration))), nil
	}
	return nil, nil
}

func (c *CloudApikeyListCommand) run(cctx *CommandContext, _ []string) error {
	filterCount := 0
	if c.UserId != "" {
		filterCount++
	}
	if c.UserEmail != "" {
		filterCount++
	}
	if c.ServiceAccountId != "" {
		filterCount++
	}
	if filterCount > 1 {
		return errors.New("only one of --user-id, --user-email, --service-account-id may be specified")
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	var ownerID string
	var ownerType identityv1.OwnerType
	switch {
	case c.UserId != "":
		ownerID = c.UserId
		ownerType = identityv1.OwnerType_OWNER_TYPE_USER
	case c.UserEmail != "":
		usersRes, err := client.GetUsers(cctx, &cloudservice.GetUsersRequest{Email: c.UserEmail})
		if err != nil {
			return err
		}
		if len(usersRes.Users) == 0 {
			return fmt.Errorf("no user found with email %q", c.UserEmail)
		}
		ownerID = usersRes.Users[0].Id
		ownerType = identityv1.OwnerType_OWNER_TYPE_USER
	case c.ServiceAccountId != "":
		ownerID = c.ServiceAccountId
		ownerType = identityv1.OwnerType_OWNER_TYPE_SERVICE_ACCOUNT
	}

	res, err := client.GetApiKeys(cctx, &cloudservice.GetApiKeysRequest{
		OwnerId:   ownerID,
		OwnerType: ownerType,
		PageSize:  int32(c.PageSize),
		PageToken: c.PageToken,
	})
	if err != nil {
		return err
	}

	return cctx.Printer.PrintResourceList(
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

func (c *CloudApikeyGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetApiKey(cctx, &cloudservice.GetApiKeyRequest{KeyId: c.KeyId})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResource(res.ApiKey, printer.PrintResourceOptions{})
}

func (c *CloudApikeyCreateForServiceAccountCommand) run(cctx *CommandContext, _ []string) error {
	expiry, err := resolveApiKeyExpiry(c.ExpiryTime, c.ExpiryDuration)
	if err != nil {
		return err
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	defer func() {
		cctx.Printer.Println("Make sure to copy or store the ApiKey token as you will not be able to see this secret again.")
	}()

	resp, err := client.CreateApiKey(cctx, &cloudservice.CreateApiKeyRequest{
		Spec: &identityv1.ApiKeySpec{
			OwnerId:     c.ServiceAccountId,
			OwnerType:   identityv1.OwnerType_OWNER_TYPE_SERVICE_ACCOUNT,
			DisplayName: c.DisplayName,
			Description: c.Description,
			ExpiryTime:  expiry,
		},
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudApikeyCreateForMeCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	identityRes, err := client.GetCurrentIdentity(cctx, &cloudservice.GetCurrentIdentityRequest{})
	if err != nil {
		return err
	}

	user := identityRes.GetUser()
	if user == nil {
		return errors.New("current principal is not a user: use create-for-service-account for service accounts")
	}

	expiry, err := resolveApiKeyExpiry(c.ExpiryTime, c.ExpiryDuration)
	if err != nil {
		return err
	}

	defer func() {
		cctx.Printer.Println("Make sure to copy or store the ApiKey token as you will not be able to see this secret again.")
	}()

	resp, err := client.CreateApiKey(cctx, &cloudservice.CreateApiKeyRequest{
		Spec: &identityv1.ApiKeySpec{
			OwnerId:     user.Id,
			OwnerType:   identityv1.OwnerType_OWNER_TYPE_USER,
			DisplayName: c.DisplayName,
			Description: c.Description,
			ExpiryTime:  expiry,
		},
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudApikeyDeleteCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetApiKey(cctx, &cloudservice.GetApiKeyRequest{KeyId: c.KeyId})
	if err != nil {
		return err
	}

	yes, err := cctx.GetPrompter().PromptYes("Delete (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting delete.")
	}

	rv := res.ApiKey.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.DeleteApiKey(cctx, &cloudservice.DeleteApiKeyRequest{
		KeyId:            c.KeyId,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleDeleteOperation(cctx, resp, err)
}

func (c *CloudApikeyUpdateCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetApiKey(cctx, &cloudservice.GetApiKeyRequest{KeyId: c.KeyId})
	if err != nil {
		return err
	}
	key := res.ApiKey
	newSpec := proto.Clone(key.Spec).(*identityv1.ApiKeySpec)

	// Only apply fields that were explicitly provided — use Changed() to distinguish
	// "not set" from zero value (especially important for --disabled=false).
	if c.Command.Flags().Changed("display-name") {
		newSpec.DisplayName = c.DisplayName
	}
	if c.Command.Flags().Changed("description") {
		newSpec.Description = c.Description
	}
	if c.Command.Flags().Changed("disabled") {
		newSpec.Disabled = c.Disabled
	}

	if err := cctx.GetPrompter().PromptApply(key.Spec, newSpec, false); err != nil {
		return err
	}

	rv := key.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateApiKey(cctx, &cloudservice.UpdateApiKeyRequest{
		KeyId:            c.KeyId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudApikeyEditCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetApiKey(cctx, &cloudservice.GetApiKeyRequest{KeyId: c.KeyId})
	if err != nil {
		return err
	}
	key := res.ApiKey

	edited, err := cctx.GetEditor().EditProto(key.Spec)
	if err != nil {
		return err
	}
	newSpec := edited.(*identityv1.ApiKeySpec)

	if err := cctx.GetPrompter().PromptApply(key.Spec, newSpec, c.VerboseDiff); err != nil {
		return err
	}

	rv := key.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateApiKey(cctx, &cloudservice.UpdateApiKeyRequest{
		KeyId:            c.KeyId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudApikeyDisableCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetApiKey(cctx, &cloudservice.GetApiKeyRequest{KeyId: c.KeyId})
	if err != nil {
		return err
	}
	key := res.ApiKey
	newSpec := proto.Clone(key.Spec).(*identityv1.ApiKeySpec)
	newSpec.Disabled = true

	if err := cctx.GetPrompter().PromptApply(key.Spec, newSpec, false); err != nil {
		return err
	}

	rv := key.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateApiKey(cctx, &cloudservice.UpdateApiKeyRequest{
		KeyId:            c.KeyId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudApikeyEnableCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetApiKey(cctx, &cloudservice.GetApiKeyRequest{KeyId: c.KeyId})
	if err != nil {
		return err
	}
	key := res.ApiKey
	newSpec := proto.Clone(key.Spec).(*identityv1.ApiKeySpec)
	newSpec.Disabled = false

	if err := cctx.GetPrompter().PromptApply(key.Spec, newSpec, false); err != nil {
		return err
	}

	rv := key.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateApiKey(cctx, &cloudservice.UpdateApiKeyRequest{
		KeyId:            c.KeyId,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}
