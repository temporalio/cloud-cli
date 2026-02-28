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
	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var defaultCreateResult = namespace.CreateNamespaceResult{
	NamespaceID: "my-namespace.my-account",
	AsyncOp:     &operation.AsyncOperation{Id: "test-operation-id"},
}

func TestCloudNamespaceCreateCommand_Success(t *testing.T) {
	mockClient := &cmdmock.MockNamespaceClient{}
	mockClient.On("CreateNamespace", mock.Anything, mock.Anything).
		Return(defaultCreateResult, nil)

	mockPoller := &cmdmock.MockPoller{}
	mockPoller.On("PollAsyncOperation", mock.Anything, "test-operation-id", "my-namespace.my-account").
		Return(nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceCreateCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCreateCommand) {
				cmd.Async = true
				cmd.AsyncOperationId = "custom-op-id"
			},
			assertResult: func(t *testing.T, buf bytes.Buffer) {
				var result temporalcloudcli.MutationResult
				err := json.Unmarshal(buf.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, temporalcloudcli.MutationResult{
					AsyncOp: &operation.AsyncOperation{Id: "test-operation-id"},
					ID:      "my-namespace.my-account",
				}, result)
			},
		},
		{
			name: "sync",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCreateCommand) {
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

			parent := &temporalcloudcli.CloudNamespaceCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCreateCommand(cctx, parent)
			cmd.Name = "my-namespace"
			cmd.Region = []string{"aws-us-east-1"}
			cmd.RetentionDays = 30
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)
			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceCreateCommand_BuildsSpecCorrectly(t *testing.T) {
	expectedSpec := &namespacev1.NamespaceSpec{
		Name:          "my-namespace",
		Regions:       []string{"aws-us-east-1"},
		RetentionDays: 30,
		ApiKeyAuth:    &namespacev1.ApiKeyAuthSpec{Enabled: false},
		Lifecycle:     &namespacev1.LifecycleSpec{EnableDeleteProtection: false},
		MtlsAuth: &namespacev1.MtlsAuthSpec{
			CertificateFilters: []*namespacev1.CertificateFilterSpec{
				{CommonName: "test.temporal.io"},
				{SubjectAlternativeName: "*.temporal.io"},
			},
		},
		SearchAttributes: map[string]namespacev1.NamespaceSpec_SearchAttributeType{
			"MyText":    namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_TEXT,
			"MyKeyword": namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
		},
	}

	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockClient.EXPECT().
		CreateNamespace(mock.Anything, mock.MatchedBy(func(p namespace.CreateNamespaceParams) bool {
			return proto.Equal(p.Spec, expectedSpec)
		})).
		Return(defaultCreateResult, nil)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
		RootCommand:     &temporalcloudcli.CloudCommand{AutoConfirm: true},
	}
	var capturedErr error
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCreateCommand(cctx, parent)
	cmd.Name = "my-namespace"
	cmd.Region = []string{"aws-us-east-1"}
	cmd.RetentionDays = 30
	cmd.SearchAttribute = []string{"MyText=Text", "MyKeyword=Keyword"}
	cmd.CertificateFilter = []string{
		`{"commonName":"test.temporal.io"}`,
		`{"subjectAlternativeName":"*.temporal.io"}`,
	}
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.NoError(t, capturedErr)
}

func TestCloudNamespaceCreateCommand_CreateNamespaceError(t *testing.T) {
	expectedErr := errors.New("create failed")

	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockClient.EXPECT().
		CreateNamespace(mock.Anything, mock.Anything).
		Return(namespace.CreateNamespaceResult{}, expectedErr)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
		RootCommand:     &temporalcloudcli.CloudCommand{AutoConfirm: true},
	}
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCreateCommand(cctx, parent)
	cmd.Name = "my-namespace"
	cmd.Region = []string{"aws-us-east-1"}
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceCreateCommand_IdempotentAlreadyExists(t *testing.T) {
	alreadyExistsErr := status.Error(codes.AlreadyExists, "namespace already exists")

	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockClient.EXPECT().
		CreateNamespace(mock.Anything, mock.Anything).
		Return(namespace.CreateNamespaceResult{}, alreadyExistsErr)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
		RootCommand:     &temporalcloudcli.CloudCommand{AutoConfirm: true},
	}
	cctx.Options.Fail = func(err error) { capturedErr = err }

	parent := &temporalcloudcli.CloudNamespaceCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCreateCommand(cctx, parent)
	cmd.Name = "my-namespace"
	cmd.Region = []string{"aws-us-east-1"}
	cmd.Idempotent = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.NoError(t, capturedErr)

	var result temporalcloudcli.Result
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "unchanged", result.Status)
	assert.Equal(t, "my-namespace", result.ID)
}

func TestCloudNamespaceCreateCommand_PollingError(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockClient.EXPECT().
		CreateNamespace(mock.Anything, mock.Anything).
		Return(defaultCreateResult, nil)

	pollErr := errors.New("polling failed")
	mockPoller := cmdmock.NewMockPoller(t)
	mockPoller.EXPECT().
		PollAsyncOperation(mock.Anything, "test-operation-id", "my-namespace.my-account").
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

	parent := &temporalcloudcli.CloudNamespaceCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCreateCommand(cctx, parent)
	cmd.Name = "my-namespace"
	cmd.Region = []string{"aws-us-east-1"}
	cmd.Async = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, pollErr, capturedErr)
}

func TestCloudNamespaceCreateCommand_UserDeclinesPrompt(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

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

	parent := &temporalcloudcli.CloudNamespaceCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCreateCommand(cctx, parent)
	cmd.Name = "my-namespace"
	cmd.Region = []string{"aws-us-east-1"}

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "Aborting create")
}

func TestCloudNamespaceCreateCommand_JSONOutputWithoutAutoConfirm(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

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

	parent := &temporalcloudcli.CloudNamespaceCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCreateCommand(cctx, parent)
	cmd.Name = "my-namespace"
	cmd.Region = []string{"aws-us-east-1"}

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "must bypass prompts when using JSON output")
}

func TestCloudNamespaceCreateCommand_InvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		setupCmd    func(*temporalcloudcli.CloudNamespaceCreateCommand)
		assertError func(*testing.T, error)
	}{
		{
			name: "invalid search attribute format",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCreateCommand) {
				cmd.SearchAttribute = []string{"MissingEquals"}
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "invalid search attribute format")
			},
		},
		{
			name: "invalid search attribute type",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCreateCommand) {
				cmd.SearchAttribute = []string{"MyAttr=NotAType"}
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "invalid search attribute type")
			},
		},
		{
			name: "invalid certificate filter JSON",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCreateCommand) {
				cmd.CertificateFilter = []string{"not-valid-json"}
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "failed to parse certificate filter")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := cmdmock.NewMockNamespaceClient(t)

			var buf bytes.Buffer
			var capturedErr error
			cctx := &temporalcloudcli.CommandContext{
				Context:         context.Background(),
				Printer:         &printer.Printer{Output: &buf, JSON: true},
				NamespaceClient: mockClient,
				RootCommand:     &temporalcloudcli.CloudCommand{AutoConfirm: true},
			}
			cctx.Options.Fail = func(err error) { capturedErr = err }

			parent := &temporalcloudcli.CloudNamespaceCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCreateCommand(cctx, parent)
			cmd.Name = "my-namespace"
			cmd.Region = []string{"aws-us-east-1"}
			cmd.Async = true
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.Error(t, capturedErr)
			tt.assertError(t, capturedErr)
		})
	}
}
