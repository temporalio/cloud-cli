package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/protobuf/proto"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

func namespaceWithEncryptionValidation(spec *namespacev1.EncryptionValidationSpec) *namespacev1.Namespace {
	return &namespacev1.Namespace{
		Namespace:       "my-ns.my-acct",
		ResourceVersion: "rv-fetched",
		Spec: &namespacev1.NamespaceSpec{
			Name:                 "my-ns",
			Regions:              []string{"aws-us-east-1"},
			RetentionDays:        30,
			EncryptionValidation: spec,
		},
	}
}

func TestNamespaceEncryptionValidationGet(t *testing.T) {
	configured := &namespacev1.EncryptionValidationSpec{
		Mode:           namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_WARN,
		MetadataKey:    "encoding",
		MetadataValues: []string{"binary/encrypted"},
	}
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceEncryptionValidationGetCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "Configured",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationGetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-acct"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(configured)}, nil)
			},
			expectedJsonOutput: map[string]any{
				"namespace": "my-ns.my-acct",
				"spec": map[string]any{
					"mode":           "ENCRYPTION_VALIDATION_MODE_WARN",
					"metadataKey":    "encoding",
					"metadataValues": []any{"binary/encrypted"},
				},
			},
		},
		{
			// The server applies defaults for an empty key or values; get shows them.
			name: "DefaultsShownWhenUnset",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationGetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-acct"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(&namespacev1.EncryptionValidationSpec{
						Mode: namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DENY,
					})}, nil)
			},
			expectedJsonOutput: map[string]any{
				"namespace": "my-ns.my-acct",
				"spec": map[string]any{
					"mode":           "ENCRYPTION_VALIDATION_MODE_DENY",
					"metadataKey":    "encoding",
					"metadataValues": []any{"binary/encrypted"},
				},
			},
		},
		{
			// Only the missing key is defaulted; configured values are shown as-is.
			name: "DefaultKeyWithCustomValues",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationGetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-acct"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(&namespacev1.EncryptionValidationSpec{
						Mode:           namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DENY,
						MetadataValues: []string{"binary/custom", "json/plain"},
					})}, nil)
			},
			expectedJsonOutput: map[string]any{
				"namespace": "my-ns.my-acct",
				"spec": map[string]any{
					"mode":           "ENCRYPTION_VALIDATION_MODE_DENY",
					"metadataKey":    "encoding",
					"metadataValues": []any{"binary/custom", "json/plain"},
				},
			},
		},
		{
			name: "NilSpec",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationGetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-acct"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(nil)}, nil)
			},
			expectedJsonOutput: map[string]any{
				"namespace": "my-ns.my-acct",
				"spec":      nil,
			},
		},
		{
			name: "GetNamespaceError",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationGetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("namespace not found"))
			},
			expectedErr: "namespace not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedJsonOutput,
			})
		})
	}
}

func TestNamespaceEncryptionValidationSet(t *testing.T) {
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand
		args                    []string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "AllFlags",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			args: []string{
				"--mode=warn",
				"--metadata-key=encoding",
				"--metadata-value=binary/encrypted",
				"--metadata-value=legacy-value",
				"--inspect-header",
				"--inspect-failure",
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-acct"}, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(nil)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return req.Namespace == "my-ns.my-acct" &&
							req.ResourceVersion == "rv-fetched" &&
							proto.Equal(req.Spec.EncryptionValidation, &namespacev1.EncryptionValidationSpec{
								Mode:           namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_WARN,
								MetadataKey:    "encoding",
								MetadataValues: []string{"binary/encrypted", "legacy-value"},
								InspectHeader:  true,
								InspectFailure: true,
							})
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-set"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-set"},
		},
		{
			name: "OmittedFlagsPreserved",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			args: []string{"--mode=deny"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(&namespacev1.EncryptionValidationSpec{
						Mode:           namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_WARN,
						MetadataKey:    "custom-key",
						MetadataValues: []string{"custom-value"},
						InspectHeader:  true,
						InspectFailure: true,
					})}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return proto.Equal(req.Spec.EncryptionValidation, &namespacev1.EncryptionValidationSpec{
							Mode:           namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DENY,
							MetadataKey:    "custom-key",
							MetadataValues: []string{"custom-value"},
							InspectHeader:  true,
							InspectFailure: true,
						})
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-set"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-set"},
		},
		{
			name: "ExplicitFalseOverrides",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			args: []string{"--inspect-failure=false"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(&namespacev1.EncryptionValidationSpec{
						Mode:           namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_WARN,
						InspectFailure: true,
					})}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return proto.Equal(req.Spec.EncryptionValidation, &namespacev1.EncryptionValidationSpec{
							Mode: namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_WARN,
						})
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-set"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-set"},
		},
		{
			name: "InvalidMode",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			args:        []string{"--mode=oops"},
			expectedErr: `invalid encryption validation mode "oops": must be disabled, warn, or deny`,
		},
		{
			name: "NoFlags",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			expectedErr: "at least one of --mode, --metadata-key, --metadata-value, --inspect-header, " +
				"or --inspect-failure must be set",
		},
		{
			// No prompt and no UpdateNamespace call are expected.
			name: "NoChange",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			args: []string{"--mode=warn", "--inspect-header=false"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(&namespacev1.EncryptionValidationSpec{
						Mode: namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_WARN,
					})}, nil)
			},
			expectedErr: "nothing to change: encryption validation is already in the requested state " +
				"(use --idempotent to succeed without changes)",
		},
		{
			name: "NoChangeIdempotent",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions:      temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				AsyncOperationOptions: temporalcloudcli.AsyncOperationOptions{Idempotent: true},
			},
			args: []string{"--mode=warn"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(&namespacev1.EncryptionValidationSpec{
						Mode: namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_WARN,
					})}, nil)
			},
			expectedJsonOutput: map[string]any{"status": "unchanged"},
		},
		{
			// Rejected before any API call.
			name: "BlankMetadataValue",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			args:        []string{"--metadata-value=binary/encrypted", "--metadata-value= "},
			expectedErr: "--metadata-value cannot be blank",
		},
		{
			name: "EmptyMetadataValue",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			args:        []string{"--metadata-value="},
			expectedErr: "--metadata-value cannot be blank",
		},
		{
			name: "ModeRequiredWhenUnconfigured",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			args: []string{"--inspect-failure"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(nil)}, nil)
			},
			expectedErr: "--mode is required when encryption validation is not configured",
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions:       temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-user"},
			},
			args: []string{"--mode=deny"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(nil)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return req.ResourceVersion == "rv-user" &&
							proto.Equal(req.Spec.EncryptionValidation, &namespacev1.EncryptionValidationSpec{
								Mode: namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DENY,
							})
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-rv"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-rv"},
		},
		{
			name: "AsyncOperationIdOverride",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions:      temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				AsyncOperationOptions: temporalcloudcli.AsyncOperationOptions{AsyncOperationId: "op-custom"},
			},
			args: []string{"--mode=warn"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(nil)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return req.AsyncOperationId == "op-custom"
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-custom"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-custom"},
		},
		{
			name: "GetNamespaceError",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			args: []string{"--mode=warn"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("namespace not found"))
			},
			expectedErr: "namespace not found",
		},
		{
			name: "UpdateNamespaceError",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			args: []string{"--mode=warn"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(nil)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("update failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			expectedErr:   "update failed",
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			args: []string{"--mode=warn"},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(nil)}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting set.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := tt.cmd.Command.Flags()
			fs.StringVar(&tt.cmd.Mode, "mode", "", "")
			fs.StringVar(&tt.cmd.MetadataKey, "metadata-key", "", "")
			fs.StringArrayVar(&tt.cmd.MetadataValue, "metadata-value", nil, "")
			fs.BoolVar(&tt.cmd.InspectHeader, "inspect-header", false, "")
			fs.BoolVar(&tt.cmd.InspectFailure, "inspect-failure", false, "")
			require.NoError(t, fs.Parse(tt.args))

			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedJsonOutput,
			})
		})
	}
}

func TestNamespaceEncryptionValidationDisable(t *testing.T) {
	existing := &namespacev1.EncryptionValidationSpec{
		Mode:           namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_WARN,
		MetadataKey:    "encoding",
		MetadataValues: []string{"binary/encrypted"},
		InspectFailure: true,
	}
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceEncryptionValidationDisableCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
		expectedJsonOutput      any
	}{
		{
			name: "DisableFromEnabled",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationDisableCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(existing)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return proto.Equal(req.Spec.EncryptionValidation, &namespacev1.EncryptionValidationSpec{
							Mode:           namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DISABLED,
							MetadataKey:    "encoding",
							MetadataValues: []string{"binary/encrypted"},
							InspectFailure: true,
						})
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-disable"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-disable"},
		},
		{
			// No-op cases: no prompt and no UpdateNamespace call are expected.
			name: "AlreadyDisabled",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationDisableCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(&namespacev1.EncryptionValidationSpec{
						Mode: namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DISABLED,
					})}, nil)
			},
			expectedErr: "nothing to change: encryption validation is already in the requested state " +
				"(use --idempotent to succeed without changes)",
		},
		{
			name: "AlreadyDisabledIdempotent",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationDisableCommand{
				NamespaceOptions:      temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				AsyncOperationOptions: temporalcloudcli.AsyncOperationOptions{Idempotent: true},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(&namespacev1.EncryptionValidationSpec{
						Mode: namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DISABLED,
					})}, nil)
			},
			expectedJsonOutput: map[string]any{"status": "unchanged"},
		},
		{
			name: "Unconfigured",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationDisableCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(nil)}, nil)
			},
			expectedErr: "nothing to change: encryption validation is already in the requested state " +
				"(use --idempotent to succeed without changes)",
		},
		{
			name: "UnconfiguredIdempotent",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationDisableCommand{
				NamespaceOptions:      temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				AsyncOperationOptions: temporalcloudcli.AsyncOperationOptions{Idempotent: true},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(nil)}, nil)
			},
			expectedJsonOutput: map[string]any{"status": "unchanged"},
		},
		{
			name: "UpdateNamespaceError",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationDisableCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(existing)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("update failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			expectedErr:   "update failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedJsonOutput,
			})
		})
	}
}
