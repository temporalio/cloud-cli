package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
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
					"mode":            2,
					"metadata_key":    "encoding",
					"metadata_values": []any{"binary/encrypted"},
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
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "ReplaceSpec",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				Mode:             "warn",
				MetadataKey:      "encoding",
				MetadataValue:    []string{"binary/encrypted", "legacy-value"},
				InspectHeader:    true,
				InspectFailure:   true,
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
			name: "InvalidMode",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				Mode:             "oops",
			},
			expectedErr: `invalid encryption validation mode "oops": must be disabled, warn, or deny`,
		},
		{
			name: "ResourceVersionOverride",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationSetCommand{
				NamespaceOptions:       temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-user"},
				Mode:                   "deny",
			},
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
				Mode:                  "warn",
			},
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
				Mode:             "warn",
			},
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
				Mode:             "warn",
			},
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
				Mode:             "warn",
			},
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
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

func TestNamespaceEncryptionValidationEnable(t *testing.T) {
	disabledWithCustomMetadata := &namespacev1.EncryptionValidationSpec{
		Mode:           namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DISABLED,
		MetadataKey:    "custom-key",
		MetadataValues: []string{"custom-value"},
		InspectHeader:  true,
	}
	warn := &namespacev1.EncryptionValidationSpec{
		Mode:           namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_WARN,
		MetadataKey:    "encoding",
		MetadataValues: []string{"binary/encrypted"},
	}
	deny := &namespacev1.EncryptionValidationSpec{
		Mode:           namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DENY,
		MetadataKey:    "encoding",
		MetadataValues: []string{"binary/encrypted"},
	}

	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceEncryptionValidationEnableCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "NilToWarnDefaults",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationEnableCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
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
								MetadataValues: []string{"binary/encrypted"},
							})
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-enable"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-enable"},
		},
		{
			name: "DisabledKeepsMetadata",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationEnableCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(disabledWithCustomMetadata)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return proto.Equal(req.Spec.EncryptionValidation, &namespacev1.EncryptionValidationSpec{
							Mode:           namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_WARN,
							MetadataKey:    "custom-key",
							MetadataValues: []string{"custom-value"},
							InspectHeader:  true,
						})
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-enable"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-enable"},
		},
		{
			name: "DenyFlag",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationEnableCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				Deny:             true,
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(disabledWithCustomMetadata)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return proto.Equal(req.Spec.EncryptionValidation, &namespacev1.EncryptionValidationSpec{
							Mode:           namespacev1.EncryptionValidationSpec_ENCRYPTION_VALIDATION_MODE_DENY,
							MetadataKey:    "custom-key",
							MetadataValues: []string{"custom-value"},
							InspectHeader:  true,
						})
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-enable"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-enable"},
		},
		{
			name: "AlreadyWarn",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationEnableCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(warn)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return proto.Equal(req.Spec.EncryptionValidation, warn)
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-enable"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-enable"},
		},
		{
			name: "AlreadyWarnWithDeny",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationEnableCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
				Deny:             true,
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(warn)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return proto.Equal(req.Spec.EncryptionValidation, warn)
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-enable"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-enable"},
		},
		{
			name: "AlreadyDeny",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationEnableCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(deny)}, nil)
				c.EXPECT().
					UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
						return proto.Equal(req.Spec.EncryptionValidation, deny)
					}), mock.Anything).
					Return(&cloudservice.UpdateNamespaceResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-enable"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-enable"},
		},
		{
			name: "GetNamespaceError",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationEnableCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("namespace not found"))
			},
			expectedErr: "namespace not found",
		},
		{
			name: "PromptDeclined",
			cmd: temporalcloudcli.CloudNamespaceEncryptionValidationEnableCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.my-acct"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespace(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceResponse{Namespace: namespaceWithEncryptionValidation(nil)}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting enable.",
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
			})
		})
	}
}
