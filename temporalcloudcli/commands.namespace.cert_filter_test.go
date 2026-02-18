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
	"go.temporal.io/cloud-sdk/api/operation/v1"
)

func TestCloudNamespaceCertFilterListCommand_Success(t *testing.T) {
	ctx := context.Background()
	mockClient := cmdmock.NewMockNamespaceClient(t)

	expectedFilters := []namespace.CertFilter{
		{
			CommonName:             "test.example.com",
			Organization:           "Example Corp",
			OrganizationalUnit:     "Engineering",
			SubjectAlternativeName: "*.example.com",
		},
		{
			CommonName:   "test2.example.com",
			Organization: "Another Corp",
		},
	}

	mockClient.EXPECT().
		ListCertFilters(ctx, "test-namespace.test-account").
		Return(expectedFilters, nil)

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

	parent := &temporalcloudcli.CloudNamespaceCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertFilterListCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"

	cmd.Command.Run(&cmd.Command, []string{})
	require.NoError(t, capturedErr)

	var result []namespace.CertFilter
	err := json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, expectedFilters, result)
}

func TestCloudNamespaceCertFilterListCommand_Error(t *testing.T) {
	ctx := context.Background()
	mockClient := cmdmock.NewMockNamespaceClient(t)

	expectedErr := errors.New("namespace not found")
	mockClient.EXPECT().
		ListCertFilters(ctx, "test-namespace.test-account").
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

	parent := &temporalcloudcli.CloudNamespaceCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertFilterListCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceCertFilterCreateCommand_Success(t *testing.T) {
	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient := &cmdmock.MockNamespaceClient{}
	mockClient.On(
		"AddCertFilters",
		mock.Anything,
		namespace.AddCertFiltersParams{
			Namespace: "test-namespace.test-account",
			Filters: []namespace.CertFilter{
				{
					CommonName:             "test.example.com",
					Organization:           "Example Corp",
					OrganizationalUnit:     "Engineering",
					SubjectAlternativeName: "*.example.com",
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
			Filters: []namespace.CertFilter{
				{
					CommonName:             "test.example.com",
					Organization:           "Example Corp",
					OrganizationalUnit:     "Engineering",
					SubjectAlternativeName: "*.example.com",
				},
			},
			ResourceVersion:  "",
			AsyncOperationID: "",
		},
	).Return(expectedOp, nil)

	mockPoller := &cmdmock.MockPoller{}
	mockPoller.On("Poll", mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceCertFilterCreateCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertFilterCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CommonName = "test.example.com"
				cmd.Organization = "Example Corp"
				cmd.OrganizationalUnit = "Engineering"
				cmd.SubjectAlternativeName = "*.example.com"
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
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertFilterCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CommonName = "test.example.com"
				cmd.Organization = "Example Corp"
				cmd.OrganizationalUnit = "Engineering"
				cmd.SubjectAlternativeName = "*.example.com"
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

			parent := &temporalcloudcli.CloudNamespaceCertFilterCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCertFilterCreateCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)

			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceCertFilterCreateCommand_NoFieldsSpecified(t *testing.T) {
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

	parent := &temporalcloudcli.CloudNamespaceCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertFilterCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "at least one filter field must be specified")
}

func TestCloudNamespaceCertFilterDeleteCommand_Success(t *testing.T) {
	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient := &cmdmock.MockNamespaceClient{}
	mockClient.On(
		"DeleteCertFilters",
		mock.Anything,
		namespace.DeleteCertFiltersParams{
			Namespace: "test-namespace.test-account",
			Filters: []namespace.CertFilter{
				{
					CommonName:   "test.example.com",
					Organization: "Example Corp",
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
			Filters: []namespace.CertFilter{
				{
					CommonName:   "test.example.com",
					Organization: "Example Corp",
				},
			},
			ResourceVersion:  "",
			AsyncOperationID: "",
		},
	).Return(expectedOp, nil)

	mockPoller := &cmdmock.MockPoller{}
	mockPoller.On("Poll", mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceCertFilterDeleteCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertFilterDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CommonName = "test.example.com"
				cmd.Organization = "Example Corp"
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
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertFilterDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CommonName = "test.example.com"
				cmd.Organization = "Example Corp"
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

			parent := &temporalcloudcli.CloudNamespaceCertFilterCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCertFilterDeleteCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)

			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceCertFilterDeleteCommand_NoFieldsSpecified(t *testing.T) {
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

	parent := &temporalcloudcli.CloudNamespaceCertFilterCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertFilterDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "at least one filter field must be specified")
}
