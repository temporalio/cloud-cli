package temporalcloudcli_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	cliext "github.com/temporalio/cli/cliext"
	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/protobuf/proto"
)

// --- ListApiKeys ---

func TestListApiKeys_NoFilter(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	var buf bytes.Buffer

	mockCloud.EXPECT().
		GetApiKeys(context.Background(), &cloudservice.GetApiKeysRequest{}).
		Return(&cloudservice.GetApiKeysResponse{
			ApiKeys: []*identityv1.ApiKey{
				{Id: "key-1", Spec: &identityv1.ApiKeySpec{DisplayName: "Key One"}},
			},
		}, nil)

	err := temporalcloudcli.ListApiKeys(context.Background(), temporalcloudcli.ListApiKeysParams{
		Cloud:   mockCloud,
		Printer: newTestPrinter(&buf),
	})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "key-1")
}

func TestListApiKeys_FilterByUserId(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	var buf bytes.Buffer

	mockCloud.EXPECT().
		GetApiKeys(context.Background(), &cloudservice.GetApiKeysRequest{
			OwnerId:   "user-123",
			OwnerType: identityv1.OwnerType_OWNER_TYPE_USER,
		}).
		Return(&cloudservice.GetApiKeysResponse{ApiKeys: nil}, nil)

	err := temporalcloudcli.ListApiKeys(context.Background(), temporalcloudcli.ListApiKeysParams{
		UserId:  "user-123",
		Cloud:   mockCloud,
		Printer: newTestPrinter(&buf),
	})
	require.NoError(t, err)
}

func TestListApiKeys_FilterByUserEmail(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	var buf bytes.Buffer

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "alice@example.com"}).
		Return(&cloudservice.GetUsersResponse{
			Users: []*identityv1.User{{Id: "user-123"}},
		}, nil)
	mockCloud.EXPECT().
		GetApiKeys(context.Background(), &cloudservice.GetApiKeysRequest{
			OwnerId:   "user-123",
			OwnerType: identityv1.OwnerType_OWNER_TYPE_USER,
		}).
		Return(&cloudservice.GetApiKeysResponse{ApiKeys: nil}, nil)

	err := temporalcloudcli.ListApiKeys(context.Background(), temporalcloudcli.ListApiKeysParams{
		UserEmail: "alice@example.com",
		Cloud:     mockCloud,
		Printer:   newTestPrinter(&buf),
	})
	require.NoError(t, err)
}

func TestListApiKeys_FilterByUserEmail_NotFound(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	var buf bytes.Buffer

	mockCloud.EXPECT().
		GetUsers(context.Background(), &cloudservice.GetUsersRequest{Email: "nobody@example.com"}).
		Return(&cloudservice.GetUsersResponse{Users: nil}, nil)

	err := temporalcloudcli.ListApiKeys(context.Background(), temporalcloudcli.ListApiKeysParams{
		UserEmail: "nobody@example.com",
		Cloud:     mockCloud,
		Printer:   newTestPrinter(&buf),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nobody@example.com")
}

func TestListApiKeys_FilterByServiceAccountId(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	var buf bytes.Buffer

	mockCloud.EXPECT().
		GetApiKeys(context.Background(), &cloudservice.GetApiKeysRequest{
			OwnerId:   "sa-456",
			OwnerType: identityv1.OwnerType_OWNER_TYPE_SERVICE_ACCOUNT,
		}).
		Return(&cloudservice.GetApiKeysResponse{ApiKeys: nil}, nil)

	err := temporalcloudcli.ListApiKeys(context.Background(), temporalcloudcli.ListApiKeysParams{
		ServiceAccountId: "sa-456",
		Cloud:            mockCloud,
		Printer:          newTestPrinter(&buf),
	})
	require.NoError(t, err)
}

func TestListApiKeys_MutuallyExclusiveFilters(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	var buf bytes.Buffer

	err := temporalcloudcli.ListApiKeys(context.Background(), temporalcloudcli.ListApiKeysParams{
		UserId:    "user-123",
		UserEmail: "alice@example.com",
		Cloud:     mockCloud,
		Printer:   newTestPrinter(&buf),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only one of")
}

// --- GetApiKey ---

func TestGetApiKey_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	var buf bytes.Buffer

	mockCloud.EXPECT().
		GetApiKey(context.Background(), &cloudservice.GetApiKeyRequest{KeyId: "key-1"}).
		Return(&cloudservice.GetApiKeyResponse{
			ApiKey: &identityv1.ApiKey{
				Id:   "key-1",
				Spec: &identityv1.ApiKeySpec{DisplayName: "My Key"},
			},
		}, nil)

	err := temporalcloudcli.GetApiKey(context.Background(), temporalcloudcli.GetApiKeyParams{
		KeyId:   "key-1",
		Cloud:   mockCloud,
		Printer: newTestPrinter(&buf),
	})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "key-1")
}

func TestGetApiKey_Error(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	var buf bytes.Buffer
	apiErr := errors.New("not found")

	mockCloud.EXPECT().
		GetApiKey(context.Background(), &cloudservice.GetApiKeyRequest{KeyId: "key-1"}).
		Return(nil, apiErr)

	err := temporalcloudcli.GetApiKey(context.Background(), temporalcloudcli.GetApiKeyParams{
		KeyId:   "key-1",
		Cloud:   mockCloud,
		Printer: newTestPrinter(&buf),
	})
	require.ErrorIs(t, err, apiErr)
}

// --- CreateApiKeyForServiceAccount ---

func TestCreateApiKeyForServiceAccount_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	var buf bytes.Buffer

	op := &operation.AsyncOperation{Id: "op-1"}
	mockCloud.EXPECT().
		CreateApiKey(context.Background(), &cloudservice.CreateApiKeyRequest{
			Spec: &identityv1.ApiKeySpec{
				OwnerId:     "sa-456",
				OwnerType:   identityv1.OwnerType_OWNER_TYPE_SERVICE_ACCOUNT,
				DisplayName: "My Key",
				Description: "a description",
			},
		}).
		Return(&cloudservice.CreateApiKeyResponse{
			KeyId:          "key-new",
			Token:          "tok-secret",
			AsyncOperation: op,
		}, nil)
	mockRunner.EXPECT().HandleOperation(op, "key-new").Return(nil)

	err := temporalcloudcli.CreateApiKeyForServiceAccount(context.Background(), temporalcloudcli.CreateApiKeyForServiceAccountParams{
		ServiceAccountId: "sa-456",
		DisplayName:      "My Key",
		Description:      "a description",
		Cloud:            mockCloud,
		Printer:          newTestPrinter(&buf),
		OperationHandler: mockRunner,
	})
	require.NoError(t, err)
}

func TestCreateApiKeyForServiceAccount_ExpiryDuration(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	var buf bytes.Buffer

	op := &operation.AsyncOperation{Id: "op-dur"}
	// We can't predict the exact expiry timestamp, so use mock.MatchedBy to verify
	// the spec fields and that ExpiryTime is set (non-nil and in the future).
	mockCloud.EXPECT().
		CreateApiKey(context.Background(), mock.MatchedBy(func(req *cloudservice.CreateApiKeyRequest) bool {
			return req.Spec.OwnerId == "sa-456" &&
				req.Spec.DisplayName == "My Key" &&
				req.Spec.ExpiryTime != nil &&
				req.Spec.ExpiryTime.AsTime().After(time.Now())
		})).
		Return(&cloudservice.CreateApiKeyResponse{
			KeyId:          "key-dur",
			Token:          "tok-dur",
			AsyncOperation: op,
		}, nil)
	mockRunner.EXPECT().HandleOperation(op, "key-dur").Return(nil)

	err := temporalcloudcli.CreateApiKeyForServiceAccount(context.Background(), temporalcloudcli.CreateApiKeyForServiceAccountParams{
		ServiceAccountId: "sa-456",
		DisplayName:      "My Key",
		ExpiryDuration:   cliext.MustParseFlagDuration("30d"),
		Cloud:            mockCloud,
		Printer:          newTestPrinter(&buf),
		OperationHandler: mockRunner,
	})
	require.NoError(t, err)
}

func TestCreateApiKeyForServiceAccount_MutuallyExclusiveExpiry(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	var buf bytes.Buffer

	err := temporalcloudcli.CreateApiKeyForServiceAccount(context.Background(), temporalcloudcli.CreateApiKeyForServiceAccountParams{
		ServiceAccountId: "sa-456",
		DisplayName:      "My Key",
		ExpiryTime:       cliext.FlagTimestamp(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)),
		ExpiryDuration:   cliext.MustParseFlagDuration("30d"),
		Cloud:            mockCloud,
		Printer:          newTestPrinter(&buf),
		OperationHandler: mockRunner,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestCreateApiKeyForServiceAccount_ApiError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	var buf bytes.Buffer
	apiErr := errors.New("create failed")

	mockCloud.EXPECT().
		CreateApiKey(context.Background(), &cloudservice.CreateApiKeyRequest{
			Spec: &identityv1.ApiKeySpec{
				OwnerId:     "sa-456",
				OwnerType:   identityv1.OwnerType_OWNER_TYPE_SERVICE_ACCOUNT,
				DisplayName: "My Key",
			},
		}).
		Return(nil, apiErr)
	mockRunner.EXPECT().HandleCreateErr(apiErr).Return(apiErr)

	err := temporalcloudcli.CreateApiKeyForServiceAccount(context.Background(), temporalcloudcli.CreateApiKeyForServiceAccountParams{
		ServiceAccountId: "sa-456",
		DisplayName:      "My Key",
		Cloud:            mockCloud,
		Printer:          newTestPrinter(&buf),
		OperationHandler: mockRunner,
	})
	require.ErrorIs(t, err, apiErr)
}

// --- CreateApiKeyForMe ---

func TestCreateApiKeyForMe_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	var buf bytes.Buffer

	mockCloud.EXPECT().
		GetCurrentIdentity(context.Background(), &cloudservice.GetCurrentIdentityRequest{}).
		Return(&cloudservice.GetCurrentIdentityResponse{
			Principal: &cloudservice.GetCurrentIdentityResponse_User{
				User: &identityv1.User{Id: "user-me"},
			},
		}, nil)

	op := &operation.AsyncOperation{Id: "op-2"}
	mockCloud.EXPECT().
		CreateApiKey(context.Background(), &cloudservice.CreateApiKeyRequest{
			Spec: &identityv1.ApiKeySpec{
				OwnerId:     "user-me",
				OwnerType:   identityv1.OwnerType_OWNER_TYPE_USER,
				DisplayName: "Personal Key",
			},
		}).
		Return(&cloudservice.CreateApiKeyResponse{
			KeyId:          "key-me",
			Token:          "tok-me",
			AsyncOperation: op,
		}, nil)
	mockRunner.EXPECT().HandleOperation(op, "key-me").Return(nil)

	err := temporalcloudcli.CreateApiKeyForMe(context.Background(), temporalcloudcli.CreateApiKeyForMeParams{
		DisplayName:      "Personal Key",
		Cloud:            mockCloud,
		Printer:          newTestPrinter(&buf),
		OperationHandler: mockRunner,
	})
	require.NoError(t, err)
}

func TestCreateApiKeyForMe_NotUser(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	var buf bytes.Buffer

	mockCloud.EXPECT().
		GetCurrentIdentity(context.Background(), &cloudservice.GetCurrentIdentityRequest{}).
		Return(&cloudservice.GetCurrentIdentityResponse{
			Principal: &cloudservice.GetCurrentIdentityResponse_ServiceAccount{},
		}, nil)

	err := temporalcloudcli.CreateApiKeyForMe(context.Background(), temporalcloudcli.CreateApiKeyForMeParams{
		DisplayName:      "Key",
		Cloud:            mockCloud,
		Printer:          newTestPrinter(&buf),
		OperationHandler: mockRunner,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a user")
}

func TestCreateApiKeyForMe_GetIdentityError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	var buf bytes.Buffer
	apiErr := errors.New("identity error")

	mockCloud.EXPECT().
		GetCurrentIdentity(context.Background(), &cloudservice.GetCurrentIdentityRequest{}).
		Return(nil, apiErr)

	err := temporalcloudcli.CreateApiKeyForMe(context.Background(), temporalcloudcli.CreateApiKeyForMeParams{
		DisplayName:      "Key",
		Cloud:            mockCloud,
		Printer:          newTestPrinter(&buf),
		OperationHandler: mockRunner,
	})
	require.ErrorIs(t, err, apiErr)
}

// --- DeleteApiKey ---

func TestDeleteApiKey_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)

	mockCloud.EXPECT().
		GetApiKey(context.Background(), &cloudservice.GetApiKeyRequest{KeyId: "key-1"}).
		Return(&cloudservice.GetApiKeyResponse{
			ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1"},
		}, nil)

	op := &operation.AsyncOperation{Id: "op-del"}
	mockCloud.EXPECT().
		DeleteApiKey(context.Background(), &cloudservice.DeleteApiKeyRequest{
			KeyId:           "key-1",
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.DeleteApiKeyResponse{AsyncOperation: op}, nil)
	mockRunner.EXPECT().HandleOperation(op, "key-1").Return(nil)

	err := temporalcloudcli.DeleteApiKey(context.Background(), temporalcloudcli.DeleteApiKeyParams{
		KeyId:            "key-1",
		Cloud:            mockCloud,
		OperationHandler: mockRunner,
	})
	require.NoError(t, err)
}

func TestDeleteApiKey_ResourceVersionOverride(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)

	mockCloud.EXPECT().
		GetApiKey(context.Background(), &cloudservice.GetApiKeyRequest{KeyId: "key-1"}).
		Return(&cloudservice.GetApiKeyResponse{
			ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1"},
		}, nil)

	op := &operation.AsyncOperation{Id: "op-del"}
	mockCloud.EXPECT().
		DeleteApiKey(context.Background(), &cloudservice.DeleteApiKeyRequest{
			KeyId:           "key-1",
			ResourceVersion: "rv-override",
		}).
		Return(&cloudservice.DeleteApiKeyResponse{AsyncOperation: op}, nil)
	mockRunner.EXPECT().HandleOperation(op, "key-1").Return(nil)

	err := temporalcloudcli.DeleteApiKey(context.Background(), temporalcloudcli.DeleteApiKeyParams{
		KeyId:            "key-1",
		ResourceVersion:  "rv-override",
		Cloud:            mockCloud,
		OperationHandler: mockRunner,
	})
	require.NoError(t, err)
}

func TestDeleteApiKey_GetError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("not found")

	mockCloud.EXPECT().
		GetApiKey(context.Background(), &cloudservice.GetApiKeyRequest{KeyId: "key-1"}).
		Return(nil, apiErr)

	err := temporalcloudcli.DeleteApiKey(context.Background(), temporalcloudcli.DeleteApiKeyParams{
		KeyId:            "key-1",
		Cloud:            mockCloud,
		OperationHandler: mockRunner,
	})
	require.ErrorIs(t, err, apiErr)
}

// --- UpdateApiKey ---

func TestUpdateApiKey_DisplayName(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)

	oldSpec := &identityv1.ApiKeySpec{DisplayName: "Old Name", Description: "desc"}
	newSpec := &identityv1.ApiKeySpec{DisplayName: "New Name", Description: "desc"}

	mockCloud.EXPECT().
		GetApiKey(context.Background(), &cloudservice.GetApiKeyRequest{KeyId: "key-1"}).
		Return(&cloudservice.GetApiKeyResponse{
			ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
		}, nil)
	mockPrompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-upd"}
	mockCloud.EXPECT().
		UpdateApiKey(context.Background(), &cloudservice.UpdateApiKeyRequest{
			KeyId:           "key-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateApiKeyResponse{AsyncOperation: op}, nil)
	mockRunner.EXPECT().HandleOperation(op, "key-1").Return(nil)

	displayName := "New Name"
	err := temporalcloudcli.UpdateApiKey(context.Background(), temporalcloudcli.UpdateApiKeyParams{
		KeyId:            "key-1",
		DisplayName:      &displayName,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.NoError(t, err)
}

func TestUpdateApiKey_DisabledFlag(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)

	oldSpec := &identityv1.ApiKeySpec{DisplayName: "Key", Disabled: false}
	newSpec := &identityv1.ApiKeySpec{DisplayName: "Key", Disabled: true}

	mockCloud.EXPECT().
		GetApiKey(context.Background(), &cloudservice.GetApiKeyRequest{KeyId: "key-1"}).
		Return(&cloudservice.GetApiKeyResponse{
			ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
		}, nil)
	mockPrompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-upd"}
	mockCloud.EXPECT().
		UpdateApiKey(context.Background(), &cloudservice.UpdateApiKeyRequest{
			KeyId:           "key-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateApiKeyResponse{AsyncOperation: op}, nil)
	mockRunner.EXPECT().HandleOperation(op, "key-1").Return(nil)

	disabled := true
	err := temporalcloudcli.UpdateApiKey(context.Background(), temporalcloudcli.UpdateApiKeyParams{
		KeyId:            "key-1",
		Disabled:         &disabled,
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.NoError(t, err)
}

func TestUpdateApiKey_NoChanges(t *testing.T) {
	// When no fields are provided (all nil), the spec is unchanged and prompter decides.
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	spec := &identityv1.ApiKeySpec{DisplayName: "Key"}
	mockCloud.EXPECT().
		GetApiKey(context.Background(), &cloudservice.GetApiKeyRequest{KeyId: "key-1"}).
		Return(&cloudservice.GetApiKeyResponse{
			ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: spec},
		}, nil)
	mockPrompter.EXPECT().PromptApply(spec, spec, false).Return(promptErr)

	err := temporalcloudcli.UpdateApiKey(context.Background(), temporalcloudcli.UpdateApiKeyParams{
		KeyId:            "key-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.ErrorIs(t, err, promptErr)
}

func TestUpdateApiKey_GetError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("get error")

	mockCloud.EXPECT().
		GetApiKey(context.Background(), &cloudservice.GetApiKeyRequest{KeyId: "key-1"}).
		Return(nil, apiErr)

	err := temporalcloudcli.UpdateApiKey(context.Background(), temporalcloudcli.UpdateApiKeyParams{
		KeyId:            "key-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.ErrorIs(t, err, apiErr)
}

// --- EditApiKey ---

func TestEditApiKey_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)

	oldSpec := &identityv1.ApiKeySpec{DisplayName: "Old"}
	editedSpec := &identityv1.ApiKeySpec{DisplayName: "Edited"}

	mockCloud.EXPECT().
		GetApiKey(context.Background(), &cloudservice.GetApiKeyRequest{KeyId: "key-1"}).
		Return(&cloudservice.GetApiKeyResponse{
			ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
		}, nil)
	mockPrompter.EXPECT().PromptApply(oldSpec, editedSpec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-edit"}
	mockCloud.EXPECT().
		UpdateApiKey(context.Background(), &cloudservice.UpdateApiKeyRequest{
			KeyId:           "key-1",
			Spec:            editedSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateApiKeyResponse{AsyncOperation: op}, nil)
	mockRunner.EXPECT().HandleOperation(op, "key-1").Return(nil)

	err := temporalcloudcli.EditApiKey(context.Background(), temporalcloudcli.EditApiKeyParams{
		KeyId:            "key-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
		RunEditor: func(existing, target proto.Message) error {
			proto.Merge(target, editedSpec)
			return nil
		},
	})
	require.NoError(t, err)
}

// --- DisableApiKey ---

func TestDisableApiKey_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)

	oldSpec := &identityv1.ApiKeySpec{DisplayName: "Key", Disabled: false}
	newSpec := &identityv1.ApiKeySpec{DisplayName: "Key", Disabled: true}

	mockCloud.EXPECT().
		GetApiKey(context.Background(), &cloudservice.GetApiKeyRequest{KeyId: "key-1"}).
		Return(&cloudservice.GetApiKeyResponse{
			ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
		}, nil)
	mockPrompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-dis"}
	mockCloud.EXPECT().
		UpdateApiKey(context.Background(), &cloudservice.UpdateApiKeyRequest{
			KeyId:           "key-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateApiKeyResponse{AsyncOperation: op}, nil)
	mockRunner.EXPECT().HandleOperation(op, "key-1").Return(nil)

	err := temporalcloudcli.DisableApiKey(context.Background(), temporalcloudcli.DisableApiKeyParams{
		KeyId:            "key-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.NoError(t, err)
}

func TestDisableApiKey_GetError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("get error")

	mockCloud.EXPECT().
		GetApiKey(context.Background(), &cloudservice.GetApiKeyRequest{KeyId: "key-1"}).
		Return(nil, apiErr)

	err := temporalcloudcli.DisableApiKey(context.Background(), temporalcloudcli.DisableApiKeyParams{
		KeyId:            "key-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.ErrorIs(t, err, apiErr)
}

// --- EnableApiKey ---

func TestEnableApiKey_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)

	oldSpec := &identityv1.ApiKeySpec{DisplayName: "Key", Disabled: true}
	newSpec := &identityv1.ApiKeySpec{DisplayName: "Key", Disabled: false}

	mockCloud.EXPECT().
		GetApiKey(context.Background(), &cloudservice.GetApiKeyRequest{KeyId: "key-1"}).
		Return(&cloudservice.GetApiKeyResponse{
			ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
		}, nil)
	mockPrompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(nil)

	op := &operation.AsyncOperation{Id: "op-en"}
	mockCloud.EXPECT().
		UpdateApiKey(context.Background(), &cloudservice.UpdateApiKeyRequest{
			KeyId:           "key-1",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateApiKeyResponse{AsyncOperation: op}, nil)
	mockRunner.EXPECT().HandleOperation(op, "key-1").Return(nil)

	err := temporalcloudcli.EnableApiKey(context.Background(), temporalcloudcli.EnableApiKeyParams{
		KeyId:            "key-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.NoError(t, err)
}

func TestEnableApiKey_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	oldSpec := &identityv1.ApiKeySpec{DisplayName: "Key", Disabled: true}
	newSpec := &identityv1.ApiKeySpec{DisplayName: "Key", Disabled: false}

	mockCloud.EXPECT().
		GetApiKey(context.Background(), &cloudservice.GetApiKeyRequest{KeyId: "key-1"}).
		Return(&cloudservice.GetApiKeyResponse{
			ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
		}, nil)
	mockPrompter.EXPECT().PromptApply(oldSpec, newSpec, false).Return(promptErr)

	err := temporalcloudcli.EnableApiKey(context.Background(), temporalcloudcli.EnableApiKeyParams{
		KeyId:            "key-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.ErrorIs(t, err, promptErr)
}
