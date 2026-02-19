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
	"go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
)

func TestCloudNamespaceTagListCommand_Success(t *testing.T) {
	ctx := context.Background()
	mockClient := cmdmock.NewMockNamespaceClient(t)

	expectedTags := []namespace.Tag{
		{
			Key:   "environment",
			Value: "production",
		},
		{
			Key:   "team",
			Value: "platform",
		},
	}

	mockClient.EXPECT().
		ListTags(ctx, "test-namespace.test-account").
		Return(expectedTags, nil)

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

	parent := &temporalcloudcli.CloudNamespaceTagCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceTagListCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"

	cmd.Command.Run(&cmd.Command, []string{})
	require.NoError(t, capturedErr)

	var result []namespace.Tag
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, expectedTags, result)
}

func TestCloudNamespaceTagListCommand_Error(t *testing.T) {
	ctx := context.Background()
	mockClient := cmdmock.NewMockNamespaceClient(t)

	expectedErr := errors.New("namespace not found")
	mockClient.EXPECT().
		ListTags(ctx, "test-namespace.test-account").
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

	parent := &temporalcloudcli.CloudNamespaceTagCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceTagListCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceTagCreateCommand_Success(t *testing.T) {
	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient := &cmdmock.MockNamespaceClient{}

	// Mock ListTags to return no existing tags
	mockClient.On(
		"ListTags",
		mock.Anything,
		"test-namespace.test-account",
	).Return([]namespace.Tag{}, nil)

	mockClient.On(
		"SetTags",
		mock.Anything,
		namespace.SetTagsParams{
			Namespace: "test-namespace.test-account",
			Tags: []namespace.Tag{
				{
					Key:   "environment",
					Value: "production",
				},
			},
			AsyncOperationID: "custom-operation-id",
		},
	).Return(expectedOp, nil)

	mockClient.On(
		"SetTags",
		mock.Anything,
		namespace.SetTagsParams{
			Namespace: "test-namespace.test-account",
			Tags: []namespace.Tag{
				{
					Key:   "environment",
					Value: "production",
				},
			},
			AsyncOperationID: "",
		},
	).Return(expectedOp, nil)

	mockPoller := &cmdmock.MockPoller{}
	mockPoller.On("Poll", mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceTagCreateCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceTagCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Key = "environment"
				cmd.Value = "production"
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
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceTagCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Key = "environment"
				cmd.Value = "production"
				cmd.Async = false
			},
			assertResult: func(t *testing.T, buf bytes.Buffer) {
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

			parent := &temporalcloudcli.CloudNamespaceTagCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceTagCreateCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)

			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceTagCreateCommand_InvalidInput(t *testing.T) {
	tests := []struct {
		name          string
		setupCmd      func(*temporalcloudcli.CloudNamespaceTagCreateCommand)
		expectedError string
	}{
		{
			name: "missing key",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceTagCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Value = "production"
			},
			expectedError: "key is required",
		},
		{
			name: "missing value",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceTagCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Key = "environment"
			},
			expectedError: "value is required",
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

			parent := &temporalcloudcli.CloudNamespaceTagCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceTagCreateCommand(cctx, parent)
			cmd.Async = true
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.Error(t, capturedErr)
			assert.Contains(t, capturedErr.Error(), tt.expectedError)
		})
	}
}

func TestCloudNamespaceTagCreateCommand_TagAlreadyExists(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	existingTags := []namespace.Tag{
		{
			Key:   "environment",
			Value: "staging",
		},
	}

	mockClient.EXPECT().
		ListTags(mock.Anything, "test-namespace.test-account").
		Return(existingTags, nil)

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

	parent := &temporalcloudcli.CloudNamespaceTagCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceTagCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Key = "environment"
	cmd.Value = "production"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "tag with key \"environment\" already exists")
}

func TestCloudNamespaceTagCreateCommand_SetTagsError(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	mockClient.EXPECT().
		ListTags(mock.Anything, "test-namespace.test-account").
		Return([]namespace.Tag{}, nil)

	expectedErr := errors.New("failed to set tags")
	mockClient.EXPECT().
		SetTags(
			mock.Anything,
			namespace.SetTagsParams{
				Namespace: "test-namespace.test-account",
				Tags: []namespace.Tag{
					{
						Key:   "environment",
						Value: "production",
					},
				},
				AsyncOperationID: "",
			},
		).
		Return(nil, expectedErr)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
		RootCommand: &temporalcloudcli.CloudCommand{
			AutoConfirm: true,
		},
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceTagCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceTagCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Key = "environment"
	cmd.Value = "production"
	cmd.Async = true
	cmd.Idempotent = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceTagCreateCommand_NothingToChange(t *testing.T) {
	nothingToChangeErr := status.Error(codes.InvalidArgument, "nothing to change")

	tests := []struct {
		name         string
		idempotent   bool
		assertResult func(*testing.T, error, bytes.Buffer)
	}{
		{
			name:       "idempotent",
			idempotent: true,
			assertResult: func(t *testing.T, capturedErr error, buf bytes.Buffer) {
				require.NoError(t, capturedErr)
				var result struct {
					Status    string `json:"Status"`
					Namespace string `json:"Namespace"`
				}
				err := json.Unmarshal(buf.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, "unchanged", result.Status)
				assert.Equal(t, "test-namespace.test-account", result.Namespace)
			},
		},
		{
			name:       "not idempotent",
			idempotent: false,
			assertResult: func(t *testing.T, capturedErr error, buf bytes.Buffer) {
				require.Error(t, capturedErr)
				assert.Equal(t, nothingToChangeErr, capturedErr)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := cmdmock.NewMockNamespaceClient(t)

			mockClient.EXPECT().
				ListTags(mock.Anything, "test-namespace.test-account").
				Return([]namespace.Tag{}, nil)

			mockClient.EXPECT().
				SetTags(
					mock.Anything,
					namespace.SetTagsParams{
						Namespace: "test-namespace.test-account",
						Tags: []namespace.Tag{
							{
								Key:   "environment",
								Value: "production",
							},
						},
						AsyncOperationID: "",
					},
				).
				Return(nil, nothingToChangeErr)

			var buf bytes.Buffer
			var capturedErr error
			cctx := &temporalcloudcli.CommandContext{
				Context:         context.Background(),
				Printer:         &printer.Printer{Output: &buf, JSON: true},
				NamespaceClient: mockClient,
				RootCommand: &temporalcloudcli.CloudCommand{
					AutoConfirm: true,
				},
			}
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceTagCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceTagCreateCommand(cctx, parent)
			cmd.Namespace = "test-namespace.test-account"
			cmd.Key = "environment"
			cmd.Value = "production"
			cmd.Async = true
			cmd.Idempotent = tt.idempotent

			cmd.Command.Run(&cmd.Command, []string{})

			tt.assertResult(t, capturedErr, buf)
		})
	}
}

func TestCloudNamespaceTagCreateCommand_PollingError(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockPoller := cmdmock.NewMockPoller(t)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient.EXPECT().
		ListTags(mock.Anything, "test-namespace.test-account").
		Return([]namespace.Tag{}, nil)

	mockClient.EXPECT().
		SetTags(
			mock.Anything,
			namespace.SetTagsParams{
				Namespace: "test-namespace.test-account",
				Tags: []namespace.Tag{
					{
						Key:   "environment",
						Value: "production",
					},
				},
				AsyncOperationID: "",
			},
		).
		Return(expectedOp, nil)

	pollErr := errors.New("polling failed")
	mockPoller.EXPECT().
		Poll(mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(pollErr)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
		Poller:          mockPoller,
		RootCommand: &temporalcloudcli.CloudCommand{
			AutoConfirm: true,
		},
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceTagCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceTagCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Key = "environment"
	cmd.Value = "production"
	cmd.Async = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, pollErr, capturedErr)
}

func TestCloudNamespaceTagCreateCommand_UserDeclinesPrompt(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	mockClient.EXPECT().
		ListTags(mock.Anything, "test-namespace.test-account").
		Return([]namespace.Tag{}, nil)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: false},
		NamespaceClient: mockClient,
		RootCommand: &temporalcloudcli.CloudCommand{
			AutoConfirm: false,
		},
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}
	cctx.Options.Stdin = bytes.NewBufferString("n\n")

	parent := &temporalcloudcli.CloudNamespaceTagCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceTagCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Key = "environment"
	cmd.Value = "production"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "Aborting create")
}

func TestCloudNamespaceTagCreateCommand_JSONOutputWithoutAutoConfirm(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	mockClient.EXPECT().
		ListTags(mock.Anything, "test-namespace.test-account").
		Return([]namespace.Tag{}, nil)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		JSONOutput:      true,
		NamespaceClient: mockClient,
		RootCommand: &temporalcloudcli.CloudCommand{
			AutoConfirm: false,
		},
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceTagCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceTagCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Key = "environment"
	cmd.Value = "production"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "must bypass prompts when using JSON output")
}

func TestCloudNamespaceTagUpdateCommand_Success(t *testing.T) {
	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient := &cmdmock.MockNamespaceClient{}

	// Mock ListTags to return an existing tag
	mockClient.On(
		"ListTags",
		mock.Anything,
		"test-namespace.test-account",
	).Return([]namespace.Tag{
		{
			Key:   "environment",
			Value: "staging",
		},
	}, nil)

	mockClient.On(
		"SetTags",
		mock.Anything,
		namespace.SetTagsParams{
			Namespace: "test-namespace.test-account",
			Tags: []namespace.Tag{
				{
					Key:   "environment",
					Value: "production",
				},
			},
			AsyncOperationID: "custom-operation-id",
		},
	).Return(expectedOp, nil)

	mockClient.On(
		"SetTags",
		mock.Anything,
		namespace.SetTagsParams{
			Namespace: "test-namespace.test-account",
			Tags: []namespace.Tag{
				{
					Key:   "environment",
					Value: "production",
				},
			},
			AsyncOperationID: "",
		},
	).Return(expectedOp, nil)

	mockPoller := &cmdmock.MockPoller{}
	mockPoller.On("Poll", mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceTagUpdateCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceTagUpdateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Key = "environment"
				cmd.Value = "production"
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
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceTagUpdateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Key = "environment"
				cmd.Value = "production"
				cmd.Async = false
			},
			assertResult: func(t *testing.T, buf bytes.Buffer) {
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

			parent := &temporalcloudcli.CloudNamespaceTagCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceTagUpdateCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)

			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceTagUpdateCommand_InvalidInput(t *testing.T) {
	tests := []struct {
		name          string
		setupCmd      func(*temporalcloudcli.CloudNamespaceTagUpdateCommand)
		expectedError string
	}{
		{
			name: "missing key",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceTagUpdateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Value = "production"
			},
			expectedError: "key is required",
		},
		{
			name: "missing value",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceTagUpdateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Key = "environment"
			},
			expectedError: "value is required",
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

			parent := &temporalcloudcli.CloudNamespaceTagCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceTagUpdateCommand(cctx, parent)
			cmd.Async = true
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.Error(t, capturedErr)
			assert.Contains(t, capturedErr.Error(), tt.expectedError)
		})
	}
}

func TestCloudNamespaceTagUpdateCommand_TagDoesNotExist(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	mockClient.EXPECT().
		ListTags(mock.Anything, "test-namespace.test-account").
		Return([]namespace.Tag{}, nil)

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

	parent := &temporalcloudcli.CloudNamespaceTagCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceTagUpdateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Key = "environment"
	cmd.Value = "production"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "tag with key \"environment\" does not exist")
}

func TestCloudNamespaceTagUpdateCommand_UserDeclinesPrompt(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	mockClient.EXPECT().
		ListTags(mock.Anything, "test-namespace.test-account").
		Return([]namespace.Tag{
			{
				Key:   "environment",
				Value: "staging",
			},
		}, nil)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: false},
		NamespaceClient: mockClient,
		RootCommand: &temporalcloudcli.CloudCommand{
			AutoConfirm: false,
		},
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}
	cctx.Options.Stdin = bytes.NewBufferString("n\n")

	parent := &temporalcloudcli.CloudNamespaceTagCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceTagUpdateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Key = "environment"
	cmd.Value = "production"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "Aborting update")
}

func TestCloudNamespaceTagDeleteCommand_Success(t *testing.T) {
	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient := &cmdmock.MockNamespaceClient{}

	mockClient.On(
		"DeleteTags",
		mock.Anything,
		namespace.DeleteTagsParams{
			Namespace:        "test-namespace.test-account",
			Keys:             []string{"environment"},
			AsyncOperationID: "custom-operation-id",
		},
	).Return(expectedOp, nil)

	mockClient.On(
		"DeleteTags",
		mock.Anything,
		namespace.DeleteTagsParams{
			Namespace:        "test-namespace.test-account",
			Keys:             []string{"environment"},
			AsyncOperationID: "",
		},
	).Return(expectedOp, nil)

	mockPoller := &cmdmock.MockPoller{}
	mockPoller.On("Poll", mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceTagDeleteCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceTagDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Key = "environment"
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
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceTagDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Key = "environment"
				cmd.Async = false
			},
			assertResult: func(t *testing.T, buf bytes.Buffer) {
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

			parent := &temporalcloudcli.CloudNamespaceTagCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceTagDeleteCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)

			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceTagDeleteCommand_InvalidInput(t *testing.T) {
	tests := []struct {
		name          string
		setupCmd      func(*temporalcloudcli.CloudNamespaceTagDeleteCommand)
		expectedError string
	}{
		{
			name: "missing key",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceTagDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
			},
			expectedError: "key is required",
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

			parent := &temporalcloudcli.CloudNamespaceTagCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceTagDeleteCommand(cctx, parent)
			cmd.Async = true
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.Error(t, capturedErr)
			assert.Contains(t, capturedErr.Error(), tt.expectedError)
		})
	}
}

func TestCloudNamespaceTagDeleteCommand_DeleteTagsError(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	expectedErr := errors.New("failed to delete tags")
	mockClient.EXPECT().
		DeleteTags(
			mock.Anything,
			namespace.DeleteTagsParams{
				Namespace:        "test-namespace.test-account",
				Keys:             []string{"environment"},
				AsyncOperationID: "",
			},
		).
		Return(nil, expectedErr)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
		RootCommand: &temporalcloudcli.CloudCommand{
			AutoConfirm: true,
		},
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceTagCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceTagDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Key = "environment"
	cmd.Async = true
	cmd.Idempotent = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceTagDeleteCommand_NothingToChange(t *testing.T) {
	nothingToChangeErr := status.Error(codes.InvalidArgument, "nothing to change")

	tests := []struct {
		name         string
		idempotent   bool
		assertResult func(*testing.T, error, bytes.Buffer)
	}{
		{
			name:       "idempotent",
			idempotent: true,
			assertResult: func(t *testing.T, capturedErr error, buf bytes.Buffer) {
				require.NoError(t, capturedErr)
				var result struct {
					Status    string `json:"Status"`
					Namespace string `json:"Namespace"`
				}
				err := json.Unmarshal(buf.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, "unchanged", result.Status)
				assert.Equal(t, "test-namespace.test-account", result.Namespace)
			},
		},
		{
			name:       "not idempotent",
			idempotent: false,
			assertResult: func(t *testing.T, capturedErr error, buf bytes.Buffer) {
				require.Error(t, capturedErr)
				assert.Equal(t, nothingToChangeErr, capturedErr)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := cmdmock.NewMockNamespaceClient(t)

			mockClient.EXPECT().
				DeleteTags(
					mock.Anything,
					namespace.DeleteTagsParams{
						Namespace:        "test-namespace.test-account",
						Keys:             []string{"environment"},
						AsyncOperationID: "",
					},
				).
				Return(nil, nothingToChangeErr)

			var buf bytes.Buffer
			var capturedErr error
			cctx := &temporalcloudcli.CommandContext{
				Context:         context.Background(),
				Printer:         &printer.Printer{Output: &buf, JSON: true},
				NamespaceClient: mockClient,
				RootCommand: &temporalcloudcli.CloudCommand{
					AutoConfirm: true,
				},
			}
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceTagCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceTagDeleteCommand(cctx, parent)
			cmd.Namespace = "test-namespace.test-account"
			cmd.Key = "environment"
			cmd.Async = true
			cmd.Idempotent = tt.idempotent

			cmd.Command.Run(&cmd.Command, []string{})

			tt.assertResult(t, capturedErr, buf)
		})
	}
}

func TestCloudNamespaceTagDeleteCommand_PollingError(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockPoller := cmdmock.NewMockPoller(t)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient.EXPECT().
		DeleteTags(
			mock.Anything,
			namespace.DeleteTagsParams{
				Namespace:        "test-namespace.test-account",
				Keys:             []string{"environment"},
				AsyncOperationID: "",
			},
		).
		Return(expectedOp, nil)

	pollErr := errors.New("polling failed")
	mockPoller.EXPECT().
		Poll(mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(pollErr)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
		Poller:          mockPoller,
		RootCommand: &temporalcloudcli.CloudCommand{
			AutoConfirm: true,
		},
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceTagCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceTagDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Key = "environment"
	cmd.Async = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, pollErr, capturedErr)
}

func TestCloudNamespaceTagDeleteCommand_UserDeclinesPrompt(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: false},
		NamespaceClient: mockClient,
		RootCommand: &temporalcloudcli.CloudCommand{
			AutoConfirm: false,
		},
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}
	cctx.Options.Stdin = bytes.NewBufferString("n\n")

	parent := &temporalcloudcli.CloudNamespaceTagCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceTagDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Key = "environment"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "Aborting delete")
}

func TestCloudNamespaceTagDeleteCommand_JSONOutputWithoutAutoConfirm(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		JSONOutput:      true,
		NamespaceClient: mockClient,
		RootCommand: &temporalcloudcli.CloudCommand{
			AutoConfirm: false,
		},
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceTagCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceTagDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Key = "environment"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "must bypass prompts when using JSON output")
}
