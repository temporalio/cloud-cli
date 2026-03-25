package temporalcloudcli_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	cliext "github.com/temporalio/cli/cliext"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

// --- ListApiKeys ---

func TestListApiKeys(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudApikeyListCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
	}{
		{
			name: "NoFilter",
			cmd:  temporalcloudcli.CloudApikeyListCommand{},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetApiKeys(mock.Anything, &cloudservice.GetApiKeysRequest{}, mock.Anything).
					Return(&cloudservice.GetApiKeysResponse{
						ApiKeys: []*identityv1.ApiKey{
							{Id: "key-1", Spec: &identityv1.ApiKeySpec{DisplayName: "Key One"}},
						},
					}, nil)
			},
		},
		{
			name: "FilterByUserId",
			cmd:  temporalcloudcli.CloudApikeyListCommand{UserId: "user-123"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetApiKeys(mock.Anything, &cloudservice.GetApiKeysRequest{
						OwnerId:   "user-123",
						OwnerType: identityv1.OwnerType_OWNER_TYPE_USER,
					}, mock.Anything).
					Return(&cloudservice.GetApiKeysResponse{}, nil)
			},
		},
		{
			name: "FilterByUserEmail",
			cmd:  temporalcloudcli.CloudApikeyListCommand{UserEmail: "alice@example.com"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "alice@example.com"}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{
						Users: []*identityv1.User{{Id: "user-123"}},
					}, nil)
				c.EXPECT().
					GetApiKeys(mock.Anything, &cloudservice.GetApiKeysRequest{
						OwnerId:   "user-123",
						OwnerType: identityv1.OwnerType_OWNER_TYPE_USER,
					}, mock.Anything).
					Return(&cloudservice.GetApiKeysResponse{}, nil)
			},
		},
		{
			name: "FilterByUserEmail_NotFound",
			cmd:  temporalcloudcli.CloudApikeyListCommand{UserEmail: "nobody@example.com"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetUsers(mock.Anything, &cloudservice.GetUsersRequest{Email: "nobody@example.com"}, mock.Anything).
					Return(&cloudservice.GetUsersResponse{}, nil)
			},
			expectedErr: "nobody@example.com",
		},
		{
			name: "FilterByServiceAccountId",
			cmd:  temporalcloudcli.CloudApikeyListCommand{ServiceAccountId: "sa-456"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetApiKeys(mock.Anything, &cloudservice.GetApiKeysRequest{
						OwnerId:   "sa-456",
						OwnerType: identityv1.OwnerType_OWNER_TYPE_SERVICE_ACCOUNT,
					}, mock.Anything).
					Return(&cloudservice.GetApiKeysResponse{}, nil)
			},
		},
		{
			name:        "MutuallyExclusiveFilters",
			cmd:         temporalcloudcli.CloudApikeyListCommand{UserId: "user-123", UserEmail: "alice@example.com"},
			expectedErr: "only one of",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, context.Background(), &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- GetApiKey ---

func TestGetApiKey(t *testing.T) {
	testApiKey := &identityv1.ApiKey{
		Id: "key-1",
		Spec: &identityv1.ApiKeySpec{
			DisplayName: "Key One",
			Description: "A description",
			OwnerId:     "user-123",
			OwnerType:   identityv1.OwnerType_OWNER_TYPE_USER,
			ExpiryTime:  timestamppb.New(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)),
		},
	}

	tests := []struct {
		name                  string
		cmd                   temporalcloudcli.CloudApikeyGetCommand
		setClientExpectations func(cloudClient *cloudmock.MockCloudServiceClient)
		expectedErr           string
		expectedJsonOutput    any
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudApikeyGetCommand{KeyId: "key-1"},
			setClientExpectations: func(cloudClient *cloudmock.MockCloudServiceClient) {
				cloudClient.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{ApiKey: testApiKey}, nil)
			},
			expectedJsonOutput: testApiKey,
		},
		{
			name: "GetApiKeyError",
			cmd:  temporalcloudcli.CloudApikeyGetCommand{KeyId: "key-1"},
			setClientExpectations: func(cloudClient *cloudmock.MockCloudServiceClient) {
				cloudClient.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(nil, errors.New("api key not found"))
			},
			expectedErr: "api key not found",
		},
		{
			name: "DisabledKey",
			cmd:  temporalcloudcli.CloudApikeyGetCommand{KeyId: "key-disabled"},
			setClientExpectations: func(cloudClient *cloudmock.MockCloudServiceClient) {
				cloudClient.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-disabled"}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{
						ApiKey: &identityv1.ApiKey{
							Id: "key-disabled",
							Spec: &identityv1.ApiKeySpec{
								DisplayName: "Disabled Key",
								OwnerId:     "user-123",
								OwnerType:   identityv1.OwnerType_OWNER_TYPE_USER,
								Disabled:    true,
							},
						},
					}, nil)
			},
			expectedJsonOutput: &identityv1.ApiKey{
				Id: "key-disabled",
				Spec: &identityv1.ApiKeySpec{
					DisplayName: "Disabled Key",
					OwnerId:     "user-123",
					OwnerType:   identityv1.OwnerType_OWNER_TYPE_USER,
					Disabled:    true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, context.Background(), &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.setClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedJsonOutput,
			})
		})
	}
}

// --- CreateApiKeyForServiceAccount ---

func TestCreateApiKeyForServiceAccount(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudApikeyCreateForServiceAccountCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudApikeyCreateForServiceAccountCommand{
				ServiceAccountId: "sa-456",
				DisplayName:      "My Key",
				Description:      "a description",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateApiKey(mock.Anything, &cloudservice.CreateApiKeyRequest{
						Spec: &identityv1.ApiKeySpec{
							OwnerId:     "sa-456",
							OwnerType:   identityv1.OwnerType_OWNER_TYPE_SERVICE_ACCOUNT,
							DisplayName: "My Key",
							Description: "a description",
						},
					}, mock.Anything).
					Return(&cloudservice.CreateApiKeyResponse{
						KeyId:          "key-new",
						Token:          "tok-secret",
						AsyncOperation: &operation.AsyncOperation{Id: "op-1"},
					}, nil)
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-1"},
			expectedJsonOutput: &cloudservice.CreateApiKeyResponse{
				KeyId: "key-new",
				Token: "tok-secret",
				AsyncOperation: &operation.AsyncOperation{
					Id:    "op-1",
					State: operation.AsyncOperation_STATE_FULFILLED,
				},
			},
		},
		{
			name: "ExpiryDuration",
			cmd: temporalcloudcli.CloudApikeyCreateForServiceAccountCommand{
				ServiceAccountId: "sa-456",
				DisplayName:      "My Key",
				ExpiryDuration:   cliext.MustParseFlagDuration("30d"),
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateApiKey(mock.Anything, mock.MatchedBy(func(req *cloudservice.CreateApiKeyRequest) bool {
						return req.Spec.OwnerId == "sa-456" &&
							req.Spec.DisplayName == "My Key" &&
							req.Spec.ExpiryTime != nil &&
							req.Spec.ExpiryTime.AsTime().After(time.Now())
					}), mock.Anything).
					Return(&cloudservice.CreateApiKeyResponse{
						KeyId:          "key-dur",
						Token:          "tok-dur",
						AsyncOperation: &operation.AsyncOperation{Id: "op-dur"},
					}, nil)
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-dur"},
		},
		{
			name: "MutuallyExclusiveExpiry",
			cmd: temporalcloudcli.CloudApikeyCreateForServiceAccountCommand{
				ServiceAccountId: "sa-456",
				DisplayName:      "My Key",
				ExpiryTime:       cliext.FlagTimestamp(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)),
				ExpiryDuration:   cliext.MustParseFlagDuration("30d"),
			},
			expectedErr: "mutually exclusive",
		},
		{
			name: "ApiError",
			cmd: temporalcloudcli.CloudApikeyCreateForServiceAccountCommand{
				ServiceAccountId: "sa-456",
				DisplayName:      "My Key",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateApiKey(mock.Anything, &cloudservice.CreateApiKeyRequest{
						Spec: &identityv1.ApiKeySpec{
							OwnerId:     "sa-456",
							OwnerType:   identityv1.OwnerType_OWNER_TYPE_SERVICE_ACCOUNT,
							DisplayName: "My Key",
						},
					}, mock.Anything).
					Return(nil, errors.New("create failed"))
			},
			expectedErr: "create operation failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, context.Background(), &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedJsonOutput,
			})
		})
	}
}

// --- CreateApiKeyForMe ---

func TestCreateApiKeyForMe(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name                  string
		cmd                   temporalcloudcli.CloudApikeyCreateForMeCommand
		setClientExpectations func(cloudClient *cloudmock.MockCloudServiceClient)
		asyncPollerOptions    temporalcloudcli.TestAsyncPollerOptions
		expectedErr           string
		expectedJsonOutput    any
	}{
		{
			name: "GetCurrentIdentityError",
			cmd:  temporalcloudcli.CloudApikeyCreateForMeCommand{DisplayName: "My Key"},
			setClientExpectations: func(cloudClient *cloudmock.MockCloudServiceClient) {
				cloudClient.EXPECT().GetCurrentIdentity(
					mock.Anything,
					&cloudservice.GetCurrentIdentityRequest{},
					mock.Anything,
				).Return(nil, errors.New("identity service unavailable")).Once()
			},
			expectedErr: "identity service unavailable",
		},
		{
			name: "PrincipalNotUser",
			cmd:  temporalcloudcli.CloudApikeyCreateForMeCommand{DisplayName: "My Key"},
			setClientExpectations: func(cloudClient *cloudmock.MockCloudServiceClient) {
				cloudClient.EXPECT().GetCurrentIdentity(
					mock.Anything,
					&cloudservice.GetCurrentIdentityRequest{},
					mock.Anything,
				).Return(&cloudservice.GetCurrentIdentityResponse{}, nil).Once()
			},
			expectedErr: "current principal is not a user",
		},
		{
			name: "MutuallyExclusiveExpiry",
			cmd: temporalcloudcli.CloudApikeyCreateForMeCommand{
				DisplayName:    "My Key",
				ExpiryTime:     cliext.FlagTimestamp(now.Add(24 * time.Hour)),
				ExpiryDuration: cliext.MustParseFlagDuration("30d"),
			},
			setClientExpectations: func(cloudClient *cloudmock.MockCloudServiceClient) {
				cloudClient.EXPECT().GetCurrentIdentity(
					mock.Anything,
					&cloudservice.GetCurrentIdentityRequest{},
					mock.Anything,
				).Return(
					&cloudservice.GetCurrentIdentityResponse{
						Principal: &cloudservice.GetCurrentIdentityResponse_User{
							User: &identityv1.User{Id: "user-123"},
						},
					},
					nil,
				).Once()
			},
			expectedErr: "mutually exclusive",
		},
		{
			name: "CreateApiKeyError",
			cmd:  temporalcloudcli.CloudApikeyCreateForMeCommand{DisplayName: "My Key"},
			setClientExpectations: func(cloudClient *cloudmock.MockCloudServiceClient) {
				cloudClient.EXPECT().GetCurrentIdentity(
					mock.Anything,
					&cloudservice.GetCurrentIdentityRequest{},
					mock.Anything,
				).Return(
					&cloudservice.GetCurrentIdentityResponse{
						Principal: &cloudservice.GetCurrentIdentityResponse_User{
							User: &identityv1.User{Id: "user-123"},
						},
					},
					nil,
				).Once()
				cloudClient.EXPECT().CreateApiKey(
					mock.Anything,
					&cloudservice.CreateApiKeyRequest{
						Spec: &identityv1.ApiKeySpec{
							OwnerId:     "user-123",
							OwnerType:   identityv1.OwnerType_OWNER_TYPE_USER,
							DisplayName: "My Key",
						},
					},
					mock.Anything,
				).Return(nil, errors.New("create failed")).Once()
			},
			expectedErr: "create operation failed",
		},
		{
			name: "AsyncPollerError",
			cmd:  temporalcloudcli.CloudApikeyCreateForMeCommand{DisplayName: "My Key"},
			setClientExpectations: func(cloudClient *cloudmock.MockCloudServiceClient) {
				cloudClient.EXPECT().GetCurrentIdentity(
					mock.Anything,
					&cloudservice.GetCurrentIdentityRequest{},
					mock.Anything,
				).Return(
					&cloudservice.GetCurrentIdentityResponse{
						Principal: &cloudservice.GetCurrentIdentityResponse_User{
							User: &identityv1.User{Id: "user-123"},
						},
					},
					nil,
				).Once()
				cloudClient.EXPECT().CreateApiKey(
					mock.Anything,
					&cloudservice.CreateApiKeyRequest{
						Spec: &identityv1.ApiKeySpec{
							OwnerId:     "user-123",
							OwnerType:   identityv1.OwnerType_OWNER_TYPE_USER,
							DisplayName: "My Key",
						},
					},
					mock.Anything,
				).Return(
					&cloudservice.CreateApiKeyResponse{
						KeyId:          "key-abc",
						Token:          "secret",
						AsyncOperation: &operation.AsyncOperation{Id: "op-poll-fail"},
					},
					nil,
				).Once()
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{
				AsyncOperationID: "op-poll-fail",
				ErrorToReturn:    errors.New("poll failed"),
			},
			expectedErr: "failed to get async operation status",
		},
		{
			name: "Success",
			cmd: temporalcloudcli.CloudApikeyCreateForMeCommand{
				DisplayName:    "My Key",
				Description:    "a description",
				ExpiryTime:     cliext.FlagTimestamp(now.Add(24 * time.Hour)),
			},
			setClientExpectations: func(cloudClient *cloudmock.MockCloudServiceClient) {
				cloudClient.EXPECT().GetCurrentIdentity(
					mock.Anything,
					&cloudservice.GetCurrentIdentityRequest{},
					mock.Anything,
				).Return(
					&cloudservice.GetCurrentIdentityResponse{
						Principal: &cloudservice.GetCurrentIdentityResponse_User{
							User: &identityv1.User{Id: "user-123"},
						},
					},
					nil,
				).Once()
				cloudClient.EXPECT().CreateApiKey(
					mock.Anything,
					&cloudservice.CreateApiKeyRequest{
						Spec: &identityv1.ApiKeySpec{
							OwnerId:     "user-123",
							OwnerType:   identityv1.OwnerType_OWNER_TYPE_USER,
							DisplayName: "My Key",
							Description: "a description",
							ExpiryTime:  timestamppb.New(now.Add(24 * time.Hour)),
						},
					},
					mock.Anything,
				).Return(
					&cloudservice.CreateApiKeyResponse{
						KeyId:          "key-abc",
						Token:          "secret",
						AsyncOperation: &operation.AsyncOperation{Id: "op-123"},
					},
					nil,
				).Once()
			},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-123"},
			expectedJsonOutput: &cloudservice.CreateApiKeyResponse{
				KeyId: "key-abc",
				Token: "secret",
				AsyncOperation: &operation.AsyncOperation{
					Id:    "op-123",
					State: operation.AsyncOperation_STATE_FULFILLED,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, context.Background(), &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.setClientExpectations,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedJsonOutput,
			})
		})
	}
}

// --- DeleteApiKey ---

func TestDeleteApiKey(t *testing.T) {
	testApiKey := &identityv1.ApiKey{
		Id:              "key-1",
		ResourceVersion: "rv-1",
		Spec: &identityv1.ApiKeySpec{
			DisplayName: "Key One",
			Description: "A description",
			OwnerId:     "user-123",
			OwnerType:   identityv1.OwnerType_OWNER_TYPE_USER,
			ExpiryTime:  timestamppb.New(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)),
		},
	}

	tt := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudApikeyDeleteCommand
		cloudClientExpectations func(cloudClient *cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		pollerOptions           temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudApikeyDeleteCommand{KeyId: "key-1"},
			cloudClientExpectations: func(cloudClient *cloudmock.MockCloudServiceClient) {
				cloudClient.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: testApiKey.Id}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{ApiKey: testApiKey}, nil)
				cloudClient.EXPECT().
					DeleteApiKey(mock.Anything, &cloudservice.DeleteApiKeyRequest{KeyId: testApiKey.Id, ResourceVersion: "rv-1"}, mock.Anything).
					Return(&cloudservice.DeleteApiKeyResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
					}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptYesResult: true},
			pollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name: "GetApiKeyError",
			cmd:  temporalcloudcli.CloudApikeyDeleteCommand{KeyId: "key-1"},
			cloudClientExpectations: func(cloudClient *cloudmock.MockCloudServiceClient) {
				cloudClient.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(nil, errors.New("api key not found"))
			},
			expectedErr: "api key not found",
		},
		{
			name: "PromptDeclined",
			cmd:  temporalcloudcli.CloudApikeyDeleteCommand{KeyId: "key-1"},
			cloudClientExpectations: func(cloudClient *cloudmock.MockCloudServiceClient) {
				cloudClient.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: testApiKey.Id}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{ApiKey: testApiKey}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptYesResult: false},
			expectedErr:   "Aborting delete.",
		},
		{
			name: "ResourceVersionOverride",
			cmd:  temporalcloudcli.CloudApikeyDeleteCommand{KeyId: "key-1", ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-override"}},
			cloudClientExpectations: func(cloudClient *cloudmock.MockCloudServiceClient) {
				cloudClient.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: testApiKey.Id}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{ApiKey: testApiKey}, nil)
				cloudClient.EXPECT().
					DeleteApiKey(mock.Anything, &cloudservice.DeleteApiKeyRequest{KeyId: testApiKey.Id, ResourceVersion: "rv-override"}, mock.Anything).
					Return(&cloudservice.DeleteApiKeyResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
					}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptYesResult: true},
			pollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name: "DeleteApiKeyError",
			cmd:  temporalcloudcli.CloudApikeyDeleteCommand{KeyId: "key-1"},
			cloudClientExpectations: func(cloudClient *cloudmock.MockCloudServiceClient) {
				cloudClient.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: testApiKey.Id}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{ApiKey: testApiKey}, nil)
				cloudClient.EXPECT().
					DeleteApiKey(mock.Anything, &cloudservice.DeleteApiKeyRequest{KeyId: testApiKey.Id, ResourceVersion: "rv-1"}, mock.Anything).
					Return(nil, errors.New("delete failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPromptYes: true, PromptYesResult: true},
			expectedErr:   "delete operation failed",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, context.Background(), &tc.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tc.cloudClientExpectations,
				PromptOptions:           tc.promptOptions,
				AsyncPollerOptions:      tc.pollerOptions,
				ExpectedError:           tc.expectedErr,
			})
		})
	}
}

// --- UpdateApiKey ---

// TestUpdateApiKey uses table-driven tests for the update apikey command.
// AIDEV-NOTE: We initialize the cobra FlagSet manually in setupCmd so Changed() returns true for
// explicitly set flags, since TestCommand calls run() directly without going through cobra's flag parsing.
func TestUpdateApiKey(t *testing.T) {
	tests := []struct {
		name                    string
		setupCmd                func(*temporalcloudcli.CloudApikeyUpdateCommand)
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "DisplayName",
			setupCmd: func(cmd *temporalcloudcli.CloudApikeyUpdateCommand) {
				cmd.Command.Flags().StringVar(&cmd.DisplayName, "display-name", "", "")
				require.NoError(t, cmd.Command.Flags().Set("display-name", "New Name"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				oldSpec := &identityv1.ApiKeySpec{DisplayName: "Old Name", Description: "desc"}
				newSpec := &identityv1.ApiKeySpec{DisplayName: "New Name", Description: "desc"}
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{
						ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateApiKey(mock.Anything, &cloudservice.UpdateApiKeyRequest{
						KeyId: "key-1", Spec: newSpec, ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateApiKeyResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-upd"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-upd"},
		},
		{
			name: "DisabledFlag",
			setupCmd: func(cmd *temporalcloudcli.CloudApikeyUpdateCommand) {
				cmd.Command.Flags().BoolVar(&cmd.Disabled, "disabled", false, "")
				require.NoError(t, cmd.Command.Flags().Set("disabled", "true"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				oldSpec := &identityv1.ApiKeySpec{DisplayName: "Key", Disabled: false}
				newSpec := &identityv1.ApiKeySpec{DisplayName: "Key", Disabled: true}
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{
						ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateApiKey(mock.Anything, &cloudservice.UpdateApiKeyRequest{
						KeyId: "key-1", Spec: newSpec, ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateApiKeyResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-upd"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-upd"},
		},
		{
			name: "NoChanges",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				spec := &identityv1.ApiKeySpec{DisplayName: "Key"}
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{
						ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: spec},
					}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{
				ExpectPrompApply: true,
				PromptApplyError: errors.New("Aborting apply."),
			},
			expectedErr: "Aborting apply.",
		},
		{
			name: "GetError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(nil, errors.New("get error"))
			},
			expectedErr: "get error",
		},
		{
			name: "ResourceVersionOverride",
			setupCmd: func(cmd *temporalcloudcli.CloudApikeyUpdateCommand) {
				cmd.ResourceVersion = "rv-override"
				cmd.Command.Flags().StringVar(&cmd.DisplayName, "display-name", "", "")
				require.NoError(t, cmd.Command.Flags().Set("display-name", "New Name"))
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				spec := &identityv1.ApiKeySpec{DisplayName: "Old Name"}
				newSpec := &identityv1.ApiKeySpec{DisplayName: "New Name"}
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{
						ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: spec},
					}, nil)
				c.EXPECT().
					UpdateApiKey(mock.Anything, &cloudservice.UpdateApiKeyRequest{
						KeyId: "key-1", Spec: newSpec, ResourceVersion: "rv-override",
					}, mock.Anything).
					Return(&cloudservice.UpdateApiKeyResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-upd"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-upd"},
		},
		{
			name: "UpdateError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				spec := &identityv1.ApiKeySpec{DisplayName: "Key"}
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{
						ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: spec},
					}, nil)
				c.EXPECT().
					UpdateApiKey(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("update failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true},
			expectedErr:   "update operation failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := temporalcloudcli.CloudApikeyUpdateCommand{KeyId: "key-1"}
			if tt.setupCmd != nil {
				tt.setupCmd(&cmd)
			}
			temporalcloudcli.TestCommand(t, context.Background(), &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- EditApiKey ---

func TestEditApiKey(t *testing.T) {
	oldSpec := &identityv1.ApiKeySpec{DisplayName: "Old"}
	editedSpec := &identityv1.ApiKeySpec{DisplayName: "Edited"}

	tests := []struct {
		name                    string
		resourceVersion         string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		editorOptions           temporalcloudcli.TestEditorOptions
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{
						ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateApiKey(mock.Anything, &cloudservice.UpdateApiKeyRequest{
						KeyId: "key-1", Spec: editedSpec, ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateApiKeyResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-edit"},
					}, nil)
			},
			editorOptions:      temporalcloudcli.TestEditorOptions{Modified: editedSpec},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-edit"},
		},
		{
			name: "GetApiKeyError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
		},
		{
			name: "EditorError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{
						ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
			},
			editorOptions: temporalcloudcli.TestEditorOptions{EditorError: errors.New("editor cancelled")},
			expectedErr:   "editor cancelled",
		},
		{
			name: "ResourceVersionOverride",
			resourceVersion: "rv-override",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{
						ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateApiKey(mock.Anything, &cloudservice.UpdateApiKeyRequest{
						KeyId: "key-1", Spec: editedSpec, ResourceVersion: "rv-override",
					}, mock.Anything).
					Return(&cloudservice.UpdateApiKeyResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-edit"},
					}, nil)
			},
			editorOptions:      temporalcloudcli.TestEditorOptions{Modified: editedSpec},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-edit"},
		},
		{
			name: "PromptDeclined",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{
						ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
			},
			editorOptions: temporalcloudcli.TestEditorOptions{Modified: editedSpec},
			promptOptions: temporalcloudcli.TestPromptOptions{
				ExpectPrompApply: true,
				PromptApplyError: errors.New("Aborting apply."),
			},
			expectedErr: "Aborting apply.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &temporalcloudcli.CloudApikeyEditCommand{KeyId: "key-1"}
			cmd.ResourceVersion = tt.resourceVersion
			temporalcloudcli.TestCommand(t, context.Background(), cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				EditorOptions:           tt.editorOptions,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- DisableApiKey ---

func TestDisableApiKey(t *testing.T) {
	oldSpec := &identityv1.ApiKeySpec{DisplayName: "Key", Disabled: false}
	newSpec := &identityv1.ApiKeySpec{DisplayName: "Key", Disabled: true}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudApikeyDisableCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudApikeyDisableCommand{KeyId: "key-1"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{
						ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateApiKey(mock.Anything, &cloudservice.UpdateApiKeyRequest{
						KeyId: "key-1", Spec: newSpec, ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateApiKeyResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-dis"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-dis"},
		},
		{
			name: "GetApiKeyError",
			cmd:  temporalcloudcli.CloudApikeyDisableCommand{KeyId: "key-1"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(nil, errors.New("get error"))
			},
			expectedErr: "get error",
		},
		{
			name: "PromptDeclined",
			cmd:  temporalcloudcli.CloudApikeyDisableCommand{KeyId: "key-1"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{
						ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{
				ExpectPrompApply: true,
				PromptApplyError: errors.New("Aborting apply."),
			},
			expectedErr: "Aborting apply.",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, context.Background(), &tc.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tc.cloudClientExpectations,
				PromptOptions:           tc.promptOptions,
				AsyncPollerOptions:      tc.asyncPollerOptions,
				ExpectedError:           tc.expectedErr,
			})
		})
	}
}

// --- EnableApiKey ---

func TestEnableApiKey(t *testing.T) {
	oldSpec := &identityv1.ApiKeySpec{DisplayName: "Key", Disabled: true}
	newSpec := &identityv1.ApiKeySpec{DisplayName: "Key", Disabled: false}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudApikeyEnableCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudApikeyEnableCommand{KeyId: "key-1"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{
						ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
				c.EXPECT().
					UpdateApiKey(mock.Anything, &cloudservice.UpdateApiKeyRequest{
						KeyId: "key-1", Spec: newSpec, ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateApiKeyResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-en"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-en"},
		},
		{
			name: "GetApiKeyError",
			cmd:  temporalcloudcli.CloudApikeyEnableCommand{KeyId: "key-1"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(nil, errors.New("get error"))
			},
			expectedErr: "get error",
		},
		{
			name: "PromptDeclined",
			cmd:  temporalcloudcli.CloudApikeyEnableCommand{KeyId: "key-1"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetApiKey(mock.Anything, &cloudservice.GetApiKeyRequest{KeyId: "key-1"}, mock.Anything).
					Return(&cloudservice.GetApiKeyResponse{
						ApiKey: &identityv1.ApiKey{Id: "key-1", ResourceVersion: "rv-1", Spec: oldSpec},
					}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{
				ExpectPrompApply: true,
				PromptApplyError: errors.New("Aborting apply."),
			},
			expectedErr: "Aborting apply.",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, context.Background(), &tc.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tc.cloudClientExpectations,
				PromptOptions:           tc.promptOptions,
				AsyncPollerOptions:      tc.asyncPollerOptions,
				ExpectedError:           tc.expectedErr,
			})
		})
	}
}
