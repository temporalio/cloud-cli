package temporalcloudcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"go.temporal.io/cloud-sdk/api/operation/v1"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
)

func TestCloudNamespaceSearchAttributeListCommand_Success(t *testing.T) {
	ctx := context.Background()
	mockClient := cmdmock.NewMockNamespaceClient(t)

	expectedAttrs := []namespace.SearchAttribute{
		{
			Name: "CustomField",
			Type: namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_TEXT,
		},
		{
			Name: "Priority",
			Type: namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_INT,
		},
	}

	mockClient.EXPECT().
		ListSearchAttributes(ctx, "test-namespace.test-account").
		Return(expectedAttrs, nil)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         ctx,
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}

	var capturedErr error
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeListCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"

	cmd.Command.Run(&cmd.Command, []string{})
	require.NoError(t, capturedErr)

	var result []namespace.SearchAttribute
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, expectedAttrs, result)
}

func TestCloudNamespaceSearchAttributeListCommand_Error(t *testing.T) {
	ctx := context.Background()
	mockClient := cmdmock.NewMockNamespaceClient(t)

	expectedErr := errors.New("namespace not found")
	mockClient.EXPECT().
		ListSearchAttributes(ctx, "test-namespace.test-account").
		Return(nil, expectedErr)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         ctx,
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}

	var capturedErr error
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeListCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceSearchAttributeCreateCommand_Success(t *testing.T) {
	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient := &cmdmock.MockNamespaceClient{}
	mockClient.On(
		"CreateSearchAttribute",
		mock.Anything,
		namespace.CreateSearchAttributeParams{
			Namespace:        "test-namespace.test-account",
			Name:             "CustomField",
			Type:             namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_TEXT,
			ResourceVersion:  "test-version",
			AsyncOperationID: "custom-operation-id",
		},
	).Return(expectedOp, nil)

	mockClient.On(
		"CreateSearchAttribute",
		mock.Anything,
		namespace.CreateSearchAttributeParams{
			Namespace:        "test-namespace.test-account",
			Name:             "CustomField",
			Type:             namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_TEXT,
			ResourceVersion:  "",
			AsyncOperationID: "",
		},
	).Return(expectedOp, nil)

	mockPoller := &cmdmock.MockPoller{}
	mockPoller.On("Poll", mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceSearchAttributeCreateCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceSearchAttributeCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Name = "CustomField"
				cmd.Type = "Text"
				cmd.ResourceVersion = "test-version"
				cmd.AsyncOperationId = "custom-operation-id"
				cmd.Async = true
			},
			assertResult: func(t *testing.T, buf bytes.Buffer) {
				var result temporalcloudcli.MutationResult
				err := json.Unmarshal(buf.Bytes(), &result)
				require.NoError(t, err)
				expected := temporalcloudcli.MutationResult{
					AsyncOp: &operation.AsyncOperation{
						Id: "test-operation-id",
					},
					ID: "test-namespace.test-account",
				}
				assert.Equal(t, expected, result)
			},
		},
		{
			name: "sync",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceSearchAttributeCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Name = "CustomField"
				cmd.Type = "Text"
				cmd.Async = false
			},
			assertResult: func(t *testing.T, buf bytes.Buffer) {
				// Sync mode just polls and returns without printing anything
				assert.Empty(t, buf.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cctx := &temporalcloudcli.CommandContext{
				Context:         context.Background(),
				Printer:         &printer.Printer{Output: &buf, JSON: true},
				NamespaceClient: mockClient,
				Poller:          mockPoller,
				RootCommand: &temporalcloudcli.CloudCommand{
					AutoConfirm: true,
				},
			}

			var capturedErr error
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeCreateCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)

			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceSearchAttributeCreateCommand_InvalidInput(t *testing.T) {
	tests := []struct {
		name          string
		setupCmd      func(*temporalcloudcli.CloudNamespaceSearchAttributeCreateCommand)
		expectedError string
	}{
		{
			name: "missing name",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceSearchAttributeCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Type = "Text"
			},
			expectedError: "name is required",
		},
		{
			name: "missing type",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceSearchAttributeCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Name = "CustomField"
			},
			expectedError: "type is required",
		},
		{
			name: "invalid type",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceSearchAttributeCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Name = "CustomField"
				cmd.Type = "InvalidType"
			},
			expectedError: "invalid search attribute type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := cmdmock.NewMockNamespaceClient(t)

			var buf bytes.Buffer
			cctx := &temporalcloudcli.CommandContext{
				Context:         context.Background(),
				Printer:         &printer.Printer{Output: &buf, JSON: true},
				NamespaceClient: mockClient,
				RootCommand: &temporalcloudcli.CloudCommand{
					AutoConfirm: true,
				},
			}

			var capturedErr error
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeCreateCommand(cctx, parent)
			cmd.Async = true
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.Error(t, capturedErr)
			assert.Contains(t, capturedErr.Error(), tt.expectedError)
		})
	}
}

func TestCloudNamespaceSearchAttributeRenameCommand_Success(t *testing.T) {
	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient := &cmdmock.MockNamespaceClient{}
	mockClient.On(
		"RenameSearchAttribute",
		mock.Anything,
		namespace.RenameSearchAttributeParams{
			Namespace:                         "test-namespace.test-account",
			ExistingCustomSearchAttributeName: "OldField",
			NewCustomSearchAttributeName:      "NewField",
			ResourceVersion:                   "test-version",
			AsyncOperationID:                  "custom-operation-id",
		},
	).Return(expectedOp, nil)

	mockClient.On(
		"RenameSearchAttribute",
		mock.Anything,
		namespace.RenameSearchAttributeParams{
			Namespace:                         "test-namespace.test-account",
			ExistingCustomSearchAttributeName: "OldField",
			NewCustomSearchAttributeName:      "NewField",
			ResourceVersion:                   "",
			AsyncOperationID:                  "",
		},
	).Return(expectedOp, nil)

	mockPoller := &cmdmock.MockPoller{}
	mockPoller.On("Poll", mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceSearchAttributeRenameCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceSearchAttributeRenameCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.ExistingName = "OldField"
				cmd.NewName = "NewField"
				cmd.ResourceVersion = "test-version"
				cmd.AsyncOperationId = "custom-operation-id"
				cmd.Async = true
			},
			assertResult: func(t *testing.T, buf bytes.Buffer) {
				var result temporalcloudcli.MutationResult
				err := json.Unmarshal(buf.Bytes(), &result)
				require.NoError(t, err)
				expected := temporalcloudcli.MutationResult{
					AsyncOp: &operation.AsyncOperation{
						Id: "test-operation-id",
					},
					ID: "test-namespace.test-account",
				}
				assert.Equal(t, expected, result)
			},
		},
		{
			name: "sync",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceSearchAttributeRenameCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.ExistingName = "OldField"
				cmd.NewName = "NewField"
				cmd.Async = false
			},
			assertResult: func(t *testing.T, buf bytes.Buffer) {
				// Sync mode just polls and returns without printing anything
				assert.Empty(t, buf.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cctx := &temporalcloudcli.CommandContext{
				Context:         context.Background(),
				Printer:         &printer.Printer{Output: &buf, JSON: true},
				NamespaceClient: mockClient,
				Poller:          mockPoller,
				RootCommand: &temporalcloudcli.CloudCommand{
					AutoConfirm: true,
				},
			}

			var capturedErr error
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeRenameCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)

			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceSearchAttributeRenameCommand_InvalidInput(t *testing.T) {
	tests := []struct {
		name          string
		setupCmd      func(*temporalcloudcli.CloudNamespaceSearchAttributeRenameCommand)
		expectedError string
	}{
		{
			name: "missing existing name",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceSearchAttributeRenameCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.NewName = "NewField"
			},
			expectedError: "existing-name is required",
		},
		{
			name: "missing new name",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceSearchAttributeRenameCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.ExistingName = "OldField"
			},
			expectedError: "new-name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := cmdmock.NewMockNamespaceClient(t)

			var buf bytes.Buffer
			cctx := &temporalcloudcli.CommandContext{
				Context:         context.Background(),
				Printer:         &printer.Printer{Output: &buf, JSON: true},
				NamespaceClient: mockClient,
				RootCommand: &temporalcloudcli.CloudCommand{
					AutoConfirm: true,
				},
			}

			var capturedErr error
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeRenameCommand(cctx, parent)
			cmd.Async = true
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.Error(t, capturedErr)
			assert.Contains(t, capturedErr.Error(), tt.expectedError)
		})
	}
}

func TestCloudNamespaceSearchAttributeDeleteCommand_Success(t *testing.T) {
	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient := &cmdmock.MockNamespaceClient{}
	mockClient.On(
		"DeleteSearchAttribute",
		mock.Anything,
		namespace.DeleteSearchAttributeParams{
			Namespace:        "test-namespace.test-account",
			Name:             "CustomField",
			ResourceVersion:  "test-version",
			AsyncOperationID: "custom-operation-id",
		},
	).Return(expectedOp, nil)

	mockClient.On(
		"DeleteSearchAttribute",
		mock.Anything,
		namespace.DeleteSearchAttributeParams{
			Namespace:        "test-namespace.test-account",
			Name:             "CustomField",
			ResourceVersion:  "",
			AsyncOperationID: "",
		},
	).Return(expectedOp, nil)

	mockPoller := &cmdmock.MockPoller{}
	mockPoller.On("Poll", mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceSearchAttributeDeleteCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceSearchAttributeDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Name = "CustomField"
				cmd.ResourceVersion = "test-version"
				cmd.AsyncOperationId = "custom-operation-id"
				cmd.Async = true
			},
			assertResult: func(t *testing.T, buf bytes.Buffer) {
				var result temporalcloudcli.MutationResult
				err := json.Unmarshal(buf.Bytes(), &result)
				require.NoError(t, err)
				expected := temporalcloudcli.MutationResult{
					AsyncOp: &operation.AsyncOperation{
						Id: "test-operation-id",
					},
					ID: "test-namespace.test-account",
				}
				assert.Equal(t, expected, result)
			},
		},
		{
			name: "sync",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceSearchAttributeDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Name = "CustomField"
				cmd.Async = false
			},
			assertResult: func(t *testing.T, buf bytes.Buffer) {
				// Sync mode just polls and returns without printing anything
				assert.Empty(t, buf.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cctx := &temporalcloudcli.CommandContext{
				Context:         context.Background(),
				Printer:         &printer.Printer{Output: &buf, JSON: true},
				NamespaceClient: mockClient,
				Poller:          mockPoller,
				RootCommand: &temporalcloudcli.CloudCommand{
					AutoConfirm: true,
				},
			}

			var capturedErr error
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeDeleteCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)

			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceSearchAttributeDeleteCommand_InvalidInput(t *testing.T) {
	tests := []struct {
		name          string
		setupCmd      func(*temporalcloudcli.CloudNamespaceSearchAttributeDeleteCommand)
		expectedError string
	}{
		{
			name: "missing name",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceSearchAttributeDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
			},
			expectedError: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := cmdmock.NewMockNamespaceClient(t)

			var buf bytes.Buffer
			cctx := &temporalcloudcli.CommandContext{
				Context:         context.Background(),
				Printer:         &printer.Printer{Output: &buf, JSON: true},
				NamespaceClient: mockClient,
				RootCommand: &temporalcloudcli.CloudCommand{
					AutoConfirm: true,
				},
			}

			var capturedErr error
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeDeleteCommand(cctx, parent)
			cmd.Async = true
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.Error(t, capturedErr)
			assert.Contains(t, capturedErr.Error(), tt.expectedError)
		})
	}
}
