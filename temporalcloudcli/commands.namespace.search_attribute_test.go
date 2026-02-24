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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
)

func TestCloudNamespaceSearchAttributeListCommand_Success(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockClient.EXPECT().
		ListSearchAttributes(mock.Anything, "test-namespace.test-account").
		Return([]namespace.SearchAttribute{
			{Name: "MyField", Type: namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD},
		}, nil)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}
	cctx.Options.Fail = func(err error) { t.Fatalf("unexpected error: %v", err) }

	parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeListCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"

	cmd.Command.Run(&cmd.Command, []string{})

	var result struct {
		SearchAttributes []temporalcloudcli.SearchAttributeOutput `json:"SearchAttributes"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, []temporalcloudcli.SearchAttributeOutput{{Name: "MyField", Type: "Keyword"}}, result.SearchAttributes)
}

func TestCloudNamespaceSearchAttributeListCommand_Error(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	expectedErr := errors.New("API error")
	mockClient.EXPECT().
		ListSearchAttributes(mock.Anything, "test-namespace.test-account").
		Return(nil, expectedErr)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeListCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceSearchAttributeCreateCommand_Success(t *testing.T) {
	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}

	mockClient := &cmdmock.MockNamespaceClient{}
	mockClient.On(
		"CreateSearchAttribute",
		mock.Anything,
		namespace.CreateSearchAttributeParams{
			Namespace:        "test-namespace.test-account",
			Name:             "MyField",
			Type:             namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
			ResourceVersion:  "test-version",
			AsyncOperationID: "custom-operation-id",
		},
	).Return(expectedOp, nil)

	mockClient.On(
		"CreateSearchAttribute",
		mock.Anything,
		namespace.CreateSearchAttributeParams{
			Namespace:        "test-namespace.test-account",
			Name:             "MyField",
			Type:             namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
			ResourceVersion:  "",
			AsyncOperationID: "",
		},
	).Return(expectedOp, nil)

	mockPoller := &cmdmock.MockPoller{}
	mockPoller.On("PollAsyncOperation", mock.Anything, "test-operation-id", "test-namespace.test-account").
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
				cmd.Name = "MyField"
				cmd.Type = "Keyword"
				cmd.ResourceVersion = "test-version"
				cmd.AsyncOperationId = "custom-operation-id"
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
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceSearchAttributeCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Name = "MyField"
				cmd.Type = "Keyword"
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

			parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeCreateCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)
			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceSearchAttributeCreateCommand_InvalidType(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Name = "MyField"
	cmd.Type = "InvalidType"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "invalid search attribute type")
}

func TestCloudNamespaceSearchAttributeCreateCommand_APIError(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	expectedErr := errors.New("API error")
	mockClient.EXPECT().
		CreateSearchAttribute(
			mock.Anything,
			namespace.CreateSearchAttributeParams{
				Namespace:        "test-namespace.test-account",
				Name:             "MyField",
				Type:             namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
				ResourceVersion:  "",
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
	}
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Name = "MyField"
	cmd.Type = "Keyword"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceSearchAttributeCreateCommand_NothingToChange(t *testing.T) {
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
				CreateSearchAttribute(
					mock.Anything,
					namespace.CreateSearchAttributeParams{
						Namespace:        "test-namespace.test-account",
						Name:             "MyField",
						Type:             namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
						ResourceVersion:  "",
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
			}
			cctx.Options.Fail = func(err error) { capturedErr = err }

			parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeCreateCommand(cctx, parent)
			cmd.Namespace = "test-namespace.test-account"
			cmd.Name = "MyField"
			cmd.Type = "Keyword"
			cmd.Async = true
			cmd.Idempotent = tt.idempotent

			cmd.Command.Run(&cmd.Command, []string{})
			tt.assertResult(t, capturedErr, buf)
		})
	}
}

func TestCloudNamespaceSearchAttributeCreateCommand_PollingError(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockPoller := cmdmock.NewMockPoller(t)

	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}
	mockClient.EXPECT().
		CreateSearchAttribute(
			mock.Anything,
			namespace.CreateSearchAttributeParams{
				Namespace:        "test-namespace.test-account",
				Name:             "MyField",
				Type:             namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
				ResourceVersion:  "",
				AsyncOperationID: "",
			},
		).
		Return(expectedOp, nil)

	pollErr := errors.New("polling failed")
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

	parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Name = "MyField"
	cmd.Type = "Keyword"
	cmd.Async = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, pollErr, capturedErr)
}

func TestCloudNamespaceSearchAttributeRenameCommand_Success(t *testing.T) {
	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}

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
	mockPoller.On("PollAsyncOperation", mock.Anything, "test-operation-id", "test-namespace.test-account").
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
				require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
				assert.Equal(t, temporalcloudcli.MutationResult{
					AsyncOp: &operation.AsyncOperation{Id: "test-operation-id"},
					ID:      "test-namespace.test-account",
				}, result)
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

			parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeRenameCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)
			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceSearchAttributeRenameCommand_APIError(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	expectedErr := errors.New("API error")
	mockClient.EXPECT().
		RenameSearchAttribute(
			mock.Anything,
			namespace.RenameSearchAttributeParams{
				Namespace:                         "test-namespace.test-account",
				ExistingCustomSearchAttributeName: "OldField",
				NewCustomSearchAttributeName:      "NewField",
				ResourceVersion:                   "",
				AsyncOperationID:                  "",
			},
		).
		Return(nil, expectedErr)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeRenameCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.ExistingName = "OldField"
	cmd.NewName = "NewField"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceSearchAttributeRenameCommand_PollingError(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockPoller := cmdmock.NewMockPoller(t)

	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}
	mockClient.EXPECT().
		RenameSearchAttribute(
			mock.Anything,
			namespace.RenameSearchAttributeParams{
				Namespace:                         "test-namespace.test-account",
				ExistingCustomSearchAttributeName: "OldField",
				NewCustomSearchAttributeName:      "NewField",
				ResourceVersion:                   "",
				AsyncOperationID:                  "",
			},
		).
		Return(expectedOp, nil)

	pollErr := errors.New("polling failed")
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

	parent := &temporalcloudcli.CloudNamespaceSearchAttributeCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceSearchAttributeRenameCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.ExistingName = "OldField"
	cmd.NewName = "NewField"
	cmd.Async = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, pollErr, capturedErr)
}
