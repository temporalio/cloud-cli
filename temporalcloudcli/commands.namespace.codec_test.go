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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"go.temporal.io/cloud-sdk/api/operation/v1"
)

func TestCloudNamespaceCodecGetCommand_NoCodecServer(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockClient.EXPECT().
		GetCodecServer(mock.Anything, "test-namespace.test-account").
		Return(nil, nil)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}
	var capturedErr error
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceCodecCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCodecGetCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"

	cmd.Command.Run(&cmd.Command, []string{})
	require.NoError(t, capturedErr)

	var result struct {
		Namespace string          `json:"Namespace"`
		Spec      json.RawMessage `json:"Spec"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "test-namespace.test-account", result.Namespace)
	assert.Equal(t, "null", string(result.Spec))
}

func TestCloudNamespaceCodecGetCommand_WithCodecServer(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockClient.EXPECT().
		GetCodecServer(mock.Anything, "test-namespace.test-account").
		Return(&namespacev1.CodecServerSpec{
			Endpoint:        "https://codec.example.com",
			PassAccessToken: true,
		}, nil)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}
	var capturedErr error
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceCodecCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCodecGetCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"

	cmd.Command.Run(&cmd.Command, []string{})
	require.NoError(t, capturedErr)
	assert.Contains(t, buf.String(), "codec.example.com")
}

func TestCloudNamespaceCodecGetCommand_Error(t *testing.T) {
	expectedErr := errors.New("API error")

	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockClient.EXPECT().
		GetCodecServer(mock.Anything, "test-namespace.test-account").
		Return(nil, expectedErr)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceCodecCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCodecGetCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceCodecSetCommand_Success(t *testing.T) {
	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}

	mockClient := &cmdmock.MockNamespaceClient{}
	mockClient.On("SetCodec", mock.Anything, mock.MatchedBy(func(p namespace.SetCodecParams) bool {
		return p.Namespace == "test-namespace.test-account" &&
			p.Endpoint == "https://codec.example.com" &&
			p.PassAccessToken == true
	})).Return(expectedOp, nil)

	mockPoller := &cmdmock.MockPoller{}
	mockPoller.On("PollAsyncOperation", mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceCodecSetCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCodecSetCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Endpoint = "https://codec.example.com"
				cmd.PassAccessToken = true
				cmd.AsyncOperationId = "test-operation-id"
				cmd.Async = true
			},
			assertResult: func(t *testing.T, buf bytes.Buffer) {
				var result temporalcloudcli.MutationResult
				require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
				assert.Equal(t, temporalcloudcli.MutationResult{
					AsyncOp: &operation.AsyncOperation{Id: "test-operation-id"},
					ID:      "test-namespace.test-account",
				}, result)
			},
		},
		{
			name: "sync",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCodecSetCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Endpoint = "https://codec.example.com"
				cmd.PassAccessToken = true
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
			}
			var capturedErr error
			cctx.Options.Fail = func(err error) { capturedErr = err }

			parent := &temporalcloudcli.CloudNamespaceCodecCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCodecSetCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)
			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceCodecSetCommand_WithCustomErrorMessage(t *testing.T) {
	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}

	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockClient.EXPECT().
		SetCodec(mock.Anything, mock.MatchedBy(func(p namespace.SetCodecParams) bool {
			return p.CustomErrorMessageDefaultMessage == "Codec unavailable" &&
				p.CustomErrorMessageDefaultLink == "https://docs.example.com"
		})).
		Return(expectedOp, nil)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}
	var capturedErr error
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceCodecCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCodecSetCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Endpoint = "https://codec.example.com"
	cmd.CustomErrorMessageDefaultMessage = "Codec unavailable"
	cmd.CustomErrorMessageDefaultLink = "https://docs.example.com"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.NoError(t, capturedErr)

	var result temporalcloudcli.MutationResult
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "test-operation-id", result.AsyncOp.Id)
}

func TestCloudNamespaceCodecSetCommand_Error(t *testing.T) {
	expectedErr := errors.New("set codec failed")

	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockClient.EXPECT().
		SetCodec(mock.Anything, mock.Anything).
		Return(nil, expectedErr)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceCodecCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCodecSetCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Endpoint = "https://codec.example.com"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceCodecSetCommand_NothingToChange(t *testing.T) {
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
				var result temporalcloudcli.Result
				require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
				assert.Equal(t, temporalcloudcli.Result{Status: "unchanged", ID: "test-namespace.test-account"}, result)
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
				SetCodec(mock.Anything, mock.Anything).
				Return(nil, nothingToChangeErr)

			var buf bytes.Buffer
			var capturedErr error
			cctx := &temporalcloudcli.CommandContext{
				Context:         context.Background(),
				Printer:         &printer.Printer{Output: &buf, JSON: true},
				NamespaceClient: mockClient,
			}
			cctx.Options.Fail = func(err error) { capturedErr = err }

			parent := &temporalcloudcli.CloudNamespaceCodecCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCodecSetCommand(cctx, parent)
			cmd.Namespace = "test-namespace.test-account"
			cmd.Endpoint = "https://codec.example.com"
			cmd.Async = true
			cmd.Idempotent = tt.idempotent

			cmd.Command.Run(&cmd.Command, []string{})
			tt.assertResult(t, capturedErr, buf)
		})
	}
}

func TestCloudNamespaceCodecSetCommand_PollingError(t *testing.T) {
	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}
	pollErr := errors.New("polling failed")

	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockClient.EXPECT().
		SetCodec(mock.Anything, mock.Anything).
		Return(expectedOp, nil)

	mockPoller := cmdmock.NewMockPoller(t)
	mockPoller.EXPECT().
		PollAsyncOperation(mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(pollErr)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
		Poller:          mockPoller,
	}
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceCodecCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCodecSetCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Endpoint = "https://codec.example.com"
	cmd.Async = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, pollErr, capturedErr)
}

func TestCloudNamespaceCodecDeleteCommand_Success(t *testing.T) {
	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}

	mockClient := &cmdmock.MockNamespaceClient{}
	mockClient.On("DeleteCodec", mock.Anything, mock.MatchedBy(func(p namespace.DeleteCodecParams) bool {
		return p.Namespace == "test-namespace.test-account"
	})).Return(expectedOp, nil)

	mockPoller := &cmdmock.MockPoller{}
	mockPoller.On("PollAsyncOperation", mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceCodecDeleteCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCodecDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.AsyncOperationId = "test-operation-id"
				cmd.Async = true
			},
			assertResult: func(t *testing.T, buf bytes.Buffer) {
				var result temporalcloudcli.MutationResult
				require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
				assert.Equal(t, temporalcloudcli.MutationResult{
					AsyncOp: &operation.AsyncOperation{Id: "test-operation-id"},
					ID:      "test-namespace.test-account",
				}, result)
			},
		},
		{
			name: "sync",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCodecDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
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
				RootCommand:     &temporalcloudcli.CloudCommand{AutoConfirm: true},
			}
			var capturedErr error
			cctx.Options.Fail = func(err error) { capturedErr = err }

			parent := &temporalcloudcli.CloudNamespaceCodecCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCodecDeleteCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)
			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceCodecDeleteCommand_Error(t *testing.T) {
	expectedErr := errors.New("delete codec failed")

	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockClient.EXPECT().
		DeleteCodec(mock.Anything, mock.Anything).
		Return(nil, expectedErr)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
		RootCommand:     &temporalcloudcli.CloudCommand{AutoConfirm: true},
	}
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceCodecCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCodecDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceCodecDeleteCommand_UserDeclinesPrompt(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	// No client methods should be called when user declines the prompt.

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: false},
		NamespaceClient: mockClient,
		RootCommand:     &temporalcloudcli.CloudCommand{AutoConfirm: false},
	}
	cctx.Options.Fail = func(err error) { capturedErr = err }
	cctx.Options.Stdin = bytes.NewBufferString("n\n")

	parent := &temporalcloudcli.CloudNamespaceCodecCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCodecDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "Aborting delete")
}

func TestCloudNamespaceCodecDeleteCommand_JSONOutputWithoutAutoConfirm(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	// No client methods should be called when the prompt check fails.

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		JSONOutput:      true,
		NamespaceClient: mockClient,
		RootCommand:     &temporalcloudcli.CloudCommand{AutoConfirm: false},
	}
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceCodecCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCodecDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "must bypass prompts when using JSON output")
}

func TestCloudNamespaceCodecDeleteCommand_NothingToChange(t *testing.T) {
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
				var result temporalcloudcli.Result
				require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
				assert.Equal(t, temporalcloudcli.Result{Status: "unchanged", ID: "test-namespace.test-account"}, result)
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
				DeleteCodec(mock.Anything, mock.Anything).
				Return(nil, nothingToChangeErr)

			var buf bytes.Buffer
			var capturedErr error
			cctx := &temporalcloudcli.CommandContext{
				Context:         context.Background(),
				Printer:         &printer.Printer{Output: &buf, JSON: true},
				NamespaceClient: mockClient,
				RootCommand:     &temporalcloudcli.CloudCommand{AutoConfirm: true},
			}
			cctx.Options.Fail = func(err error) { capturedErr = err }

			parent := &temporalcloudcli.CloudNamespaceCodecCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCodecDeleteCommand(cctx, parent)
			cmd.Namespace = "test-namespace.test-account"
			cmd.Async = true
			cmd.Idempotent = tt.idempotent

			cmd.Command.Run(&cmd.Command, []string{})
			tt.assertResult(t, capturedErr, buf)
		})
	}
}

func TestCloudNamespaceCodecDeleteCommand_PollingError(t *testing.T) {
	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}
	pollErr := errors.New("polling failed")

	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockPoller := cmdmock.NewMockPoller(t)

	mockClient.EXPECT().
		DeleteCodec(mock.Anything, mock.Anything).
		Return(expectedOp, nil)
	mockPoller.EXPECT().
		PollAsyncOperation(mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(pollErr)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
		Poller:          mockPoller,
		RootCommand:     &temporalcloudcli.CloudCommand{AutoConfirm: true},
	}
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceCodecCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCodecDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Async = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, pollErr, capturedErr)
}
