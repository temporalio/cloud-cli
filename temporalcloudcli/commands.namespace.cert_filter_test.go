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
)

func TestCloudNamespaceCertFilterListCommand_Success(t *testing.T) {
	expectedFilters := []*namespacev1.CertificateFilterSpec{
		{
			CommonName:   "test.temporal.io",
			Organization: "Temporal",
		},
		{
			SubjectAlternativeName: "*.temporal.io",
		},
	}

	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockClient.EXPECT().
		ListCertFilters(mock.Anything, "test-namespace.test-account").
		Return(expectedFilters, nil)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}

	var capturedErr error
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterListCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"

	cmd.Command.Run(&cmd.Command, []string{})
	require.NoError(t, capturedErr)

	var result struct {
		CertificateFilters []*namespacev1.CertificateFilterSpec `json:"certificateFilters"`
	}
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	expected := struct {
		CertificateFilters []*namespacev1.CertificateFilterSpec `json:"certificateFilters"`
	}{CertificateFilters: expectedFilters}
	assert.Equal(t, expected, result)
}

func TestCloudNamespaceMtlsCertFilterListCommand_EmptyList(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockClient.EXPECT().
		ListCertFilters(mock.Anything, "test-namespace.test-account").
		Return([]*namespacev1.CertificateFilterSpec{}, nil)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}

	var capturedErr error
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterListCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"

	cmd.Command.Run(&cmd.Command, []string{})
	require.NoError(t, capturedErr)

	var result struct {
		CertificateFilters []*namespacev1.CertificateFilterSpec `json:"certificateFilters"`
	}
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Empty(t, result.CertificateFilters)
}

func TestCloudNamespaceMtlsCertFilterListCommand_ListError(t *testing.T) {
	expectedErr := errors.New("failed to list filters")

	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockClient.EXPECT().
		ListCertFilters(mock.Anything, "test-namespace.test-account").
		Return(nil, expectedErr)

	var buf bytes.Buffer
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterListCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceMtlsCertFilterCreateCommand_Success(t *testing.T) {
	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient := &cmdmock.MockNamespaceClient{}
	mockClient.On(
		"AddCertFilters",
		mock.Anything,
		namespace.AddCertFiltersParams{
			Namespace: "test-namespace.test-account",
			Filters: []*namespacev1.CertificateFilterSpec{
				{
					CommonName:   "test.temporal.io",
					Organization: "Temporal",
				},
			},
			ResourceVersion:  "test-version",
			AsyncOperationID: "custom-operation-id",
		},
	).Return(expectedOp, nil)

	mockClient.On(
		"AddCertFilters",
		mock.Anything,
		namespace.AddCertFiltersParams{
			Namespace: "test-namespace.test-account",
			Filters: []*namespacev1.CertificateFilterSpec{
				{
					CommonName:   "test.temporal.io",
					Organization: "Temporal",
				},
			},
			ResourceVersion:  "",
			AsyncOperationID: "",
		},
	).Return(expectedOp, nil)

	mockPoller := &cmdmock.MockPoller{}
	mockPoller.On("PollAsyncOperation", mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceMtlsCertFilterCreateCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceMtlsCertFilterCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CommonName = "test.temporal.io"
				cmd.Organization = "Temporal"
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
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceMtlsCertFilterCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CommonName = "test.temporal.io"
				cmd.Organization = "Temporal"
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

			parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterCreateCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)

			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceMtlsCertFilterCreateCommand_InvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		setupCmd    func(*temporalcloudcli.CloudNamespaceMtlsCertFilterCreateCommand)
		assertError func(*testing.T, error)
	}{
		{
			name: "no fields provided",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceMtlsCertFilterCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Async = true
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "at least one certificate filter field must be specified")
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
				RootCommand: &temporalcloudcli.CloudCommand{
					AutoConfirm: true,
				},
			}
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterCreateCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})

			require.Error(t, capturedErr)
			tt.assertError(t, capturedErr)
		})
	}
}

func TestCloudNamespaceMtlsCertFilterCreateCommand_AddCertFiltersError(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	expectedErr := errors.New("failed to add filters")
	mockClient.EXPECT().
		AddCertFilters(
			mock.Anything,
			namespace.AddCertFiltersParams{
				Namespace: "test-namespace.test-account",
				Filters: []*namespacev1.CertificateFilterSpec{
					{
						CommonName: "test.temporal.io",
					},
				},
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
		RootCommand: &temporalcloudcli.CloudCommand{
			AutoConfirm: true,
		},
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CommonName = "test.temporal.io"
	cmd.Async = true
	cmd.Idempotent = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceCertFilterCreateCommand_NothingToChange(t *testing.T) {
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
					Status string `json:"status"`
				}
				err := json.Unmarshal(buf.Bytes(), &result)
				require.NoError(t, err)
				expected := struct {
					Status string `json:"status"`
				}{
					Status: "unchanged",
				}
				assert.Equal(t, expected, result)
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
				AddCertFilters(
					mock.Anything,
					namespace.AddCertFiltersParams{
						Namespace: "test-namespace.test-account",
						Filters: []*namespacev1.CertificateFilterSpec{
							{
								CommonName: "test.temporal.io",
							},
						},
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
				RootCommand: &temporalcloudcli.CloudCommand{
					AutoConfirm: true,
				},
			}
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterCreateCommand(cctx, parent)
			cmd.Namespace = "test-namespace.test-account"
			cmd.CommonName = "test.temporal.io"
			cmd.Async = true
			cmd.Idempotent = tt.idempotent

			cmd.Command.Run(&cmd.Command, []string{})

			tt.assertResult(t, capturedErr, buf)
		})
	}
}

func TestCloudNamespaceMtlsCertFilterCreateCommand_IdempotentWithOtherError(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	otherErr := status.Error(codes.PermissionDenied, "permission denied")
	mockClient.EXPECT().
		AddCertFilters(
			mock.Anything,
			namespace.AddCertFiltersParams{
				Namespace: "test-namespace.test-account",
				Filters: []*namespacev1.CertificateFilterSpec{
					{
						CommonName: "test.temporal.io",
					},
				},
				ResourceVersion:  "",
				AsyncOperationID: "",
			},
		).
		Return(nil, otherErr)

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

	parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CommonName = "test.temporal.io"
	cmd.Async = true
	cmd.Idempotent = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, otherErr, capturedErr)
}

func TestCloudNamespaceMtlsCertFilterCreateCommand_PollingError(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockPoller := cmdmock.NewMockPoller(t)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient.EXPECT().
		AddCertFilters(
			mock.Anything,
			namespace.AddCertFiltersParams{
				Namespace: "test-namespace.test-account",
				Filters: []*namespacev1.CertificateFilterSpec{
					{
						CommonName: "test.temporal.io",
					},
				},
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
		RootCommand: &temporalcloudcli.CloudCommand{
			AutoConfirm: true,
		},
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CommonName = "test.temporal.io"
	cmd.Async = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, pollErr, capturedErr)
}

func TestCloudNamespaceMtlsCertFilterCreateCommand_UserDeclinesPrompt(t *testing.T) {
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
	// Simulate user declining the prompt by providing "n" as input
	cctx.Options.Stdin = bytes.NewBufferString("n\n")

	parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CommonName = "test.temporal.io"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "Aborting create")
}

func TestCloudNamespaceMtlsCertFilterCreateCommand_JSONOutputWithoutAutoConfirm(t *testing.T) {
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

	parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CommonName = "test.temporal.io"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "must bypass prompts when using JSON output")
}

func TestCloudNamespaceMtlsCertFilterDeleteCommand_Success(t *testing.T) {
	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient := &cmdmock.MockNamespaceClient{}
	mockClient.On(
		"DeleteCertFilters",
		mock.Anything,
		namespace.DeleteCertFiltersParams{
			Namespace: "test-namespace.test-account",
			Filters: []*namespacev1.CertificateFilterSpec{
				{
					CommonName:   "test.temporal.io",
					Organization: "Temporal",
				},
			},
			ResourceVersion:  "test-version",
			AsyncOperationID: "custom-operation-id",
		},
	).Return(expectedOp, nil)

	mockClient.On(
		"DeleteCertFilters",
		mock.Anything,
		namespace.DeleteCertFiltersParams{
			Namespace: "test-namespace.test-account",
			Filters: []*namespacev1.CertificateFilterSpec{
				{
					CommonName:   "test.temporal.io",
					Organization: "Temporal",
				},
			},
			ResourceVersion:  "",
			AsyncOperationID: "",
		},
	).Return(expectedOp, nil)

	mockPoller := &cmdmock.MockPoller{}
	mockPoller.On("PollAsyncOperation", mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceMtlsCertFilterDeleteCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceMtlsCertFilterDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CommonName = "test.temporal.io"
				cmd.Organization = "Temporal"
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
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceMtlsCertFilterDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CommonName = "test.temporal.io"
				cmd.Organization = "Temporal"
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

			parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterDeleteCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)

			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceMtlsCertFilterDeleteCommand_InvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		setupCmd    func(*temporalcloudcli.CloudNamespaceMtlsCertFilterDeleteCommand)
		assertError func(*testing.T, error)
	}{
		{
			name: "no fields provided",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceMtlsCertFilterDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Async = true
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "at least one certificate filter field must be specified")
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
				RootCommand: &temporalcloudcli.CloudCommand{
					AutoConfirm: true,
				},
			}
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterDeleteCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})

			require.Error(t, capturedErr)
			tt.assertError(t, capturedErr)
		})
	}
}

func TestCloudNamespaceMtlsCertFilterDeleteCommand_DeleteCertFiltersError(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	expectedErr := errors.New("failed to remove filters")
	mockClient.EXPECT().
		DeleteCertFilters(
			mock.Anything,
			namespace.DeleteCertFiltersParams{
				Namespace: "test-namespace.test-account",
				Filters: []*namespacev1.CertificateFilterSpec{
					{
						CommonName: "test.temporal.io",
					},
				},
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
		RootCommand: &temporalcloudcli.CloudCommand{
			AutoConfirm: true,
		},
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CommonName = "test.temporal.io"
	cmd.Async = true
	cmd.Idempotent = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceMtlsCertFilterDeleteCommand_NothingToChange(t *testing.T) {
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
					Status string `json:"status"`
				}
				err := json.Unmarshal(buf.Bytes(), &result)
				require.NoError(t, err)
				expected := struct {
					Status string `json:"status"`
				}{
					Status: "unchanged",
				}
				assert.Equal(t, expected, result)
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
				DeleteCertFilters(
					mock.Anything,
					namespace.DeleteCertFiltersParams{
						Namespace: "test-namespace.test-account",
						Filters: []*namespacev1.CertificateFilterSpec{
							{
								CommonName: "test.temporal.io",
							},
						},
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
				RootCommand: &temporalcloudcli.CloudCommand{
					AutoConfirm: true,
				},
			}
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterDeleteCommand(cctx, parent)
			cmd.Namespace = "test-namespace.test-account"
			cmd.CommonName = "test.temporal.io"
			cmd.Async = true
			cmd.Idempotent = tt.idempotent

			cmd.Command.Run(&cmd.Command, []string{})

			tt.assertResult(t, capturedErr, buf)
		})
	}
}

func TestCloudNamespaceMtlsCertFilterDeleteCommand_IdempotentWithOtherError(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	otherErr := status.Error(codes.PermissionDenied, "permission denied")
	mockClient.EXPECT().
		DeleteCertFilters(
			mock.Anything,
			namespace.DeleteCertFiltersParams{
				Namespace: "test-namespace.test-account",
				Filters: []*namespacev1.CertificateFilterSpec{
					{
						CommonName: "test.temporal.io",
					},
				},
				ResourceVersion:  "",
				AsyncOperationID: "",
			},
		).
		Return(nil, otherErr)

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

	parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CommonName = "test.temporal.io"
	cmd.Async = true
	cmd.Idempotent = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, otherErr, capturedErr)
}

func TestCloudNamespaceCertFilterDeleteCommand_PollingError(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockPoller := cmdmock.NewMockPoller(t)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient.EXPECT().
		DeleteCertFilters(
			mock.Anything,
			namespace.DeleteCertFiltersParams{
				Namespace: "test-namespace.test-account",
				Filters: []*namespacev1.CertificateFilterSpec{
					{
						CommonName: "test.temporal.io",
					},
				},
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
		RootCommand: &temporalcloudcli.CloudCommand{
			AutoConfirm: true,
		},
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CommonName = "test.temporal.io"
	cmd.Async = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, pollErr, capturedErr)
}

func TestCloudNamespaceMtlsCertFilterDeleteCommand_UserDeclinesPrompt(t *testing.T) {
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
	// Simulate user declining the prompt by providing "n" as input
	cctx.Options.Stdin = bytes.NewBufferString("n\n")

	parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CommonName = "test.temporal.io"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "Aborting delete")
}

func TestCloudNamespaceMtlsCertFilterDeleteCommand_JSONOutputWithoutAutoConfirm(t *testing.T) {
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

	parent := &temporalcloudcli.CloudNamespaceMtlsCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceMtlsCertFilterDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CommonName = "test.temporal.io"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "must bypass prompts when using JSON output")
}
