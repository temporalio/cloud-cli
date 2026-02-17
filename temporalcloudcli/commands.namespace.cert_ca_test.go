package temporalcloudcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cloud-cli/internal/cert"
	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCloudNamespaceCertCaAddCommand_Success(t *testing.T) {
	validCertData, err := os.ReadFile("testdata/valid-cert.pem")
	require.NoError(t, err)
	parsedCerts, err := cert.ParseCACerts(validCertData)
	require.NoError(t, err)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient := &cmdmock.MockNamespaceClient{}
	mockClient.On(
		"AddCACerts",
		mock.Anything,
		namespace.AddCACertsParams{
			Namespace:        "test-namespace.test-account",
			Certs:            parsedCerts,
			ResourceVersion:  "test-version",
			AsyncOperationID: "custom-operation-id",
		},
	).Return(expectedOp, nil)

	mockClient.On(
		"AddCACerts",
		mock.Anything,
		namespace.AddCACertsParams{
			Namespace:        "test-namespace.test-account",
			Certs:            parsedCerts,
			ResourceVersion:  "",
			AsyncOperationID: "",
		},
	).Return(expectedOp, nil)

	mockPoller := &cmdmock.MockPoller{}
	mockPoller.On("Poll", mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	expectedCerts := []byte("cert-bundle-data")
	expectedNamespace := &namespacev1.Namespace{
		Namespace: "test-namespace.test-account",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: expectedCerts,
			},
		},
	}
	mockClient.EXPECT().
		GetNamespace(mock.Anything, "test-namespace.test-account").
		Return(expectedNamespace, nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceCertCaAddCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaAddCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificateFile = "testdata/valid-cert.pem"
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
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaAddCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificateFile = "testdata/valid-cert.pem"
				cmd.Async = false
			},
			assertResult: func(t *testing.T, buf bytes.Buffer) {
				expectedCerts := []byte("cert-bundle-data")
				var result []byte
				err := json.Unmarshal(buf.Bytes(), &result)
				require.NoError(t, err)
				assert.Equal(t, expectedCerts, result)
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

			parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, err)

			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceCertCaAddCommand_FileNotFound(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = "testdata/nonexistent-cert.pem"
	cmd.Async = true

	err := cmd.Command.RunE(&cmd.Command, []string{})
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestCloudNamespaceCertCaAddCommand_InvalidCertificate(t *testing.T) {
	mockClient := cmdmock.NewMockNamespaceClient(t)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = "testdata/invalid-cert.pem"
	cmd.Async = true

	err := cmd.Command.RunE(&cmd.Command, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestCloudNamespaceCertCaAddCommand_AddCACertsError(t *testing.T) {
	validCertData, err := os.ReadFile("testdata/valid-cert.pem")
	require.NoError(t, err)
	parsedCerts, err := cert.ParseCACerts(validCertData)
	require.NoError(t, err)

	mockClient := cmdmock.NewMockNamespaceClient(t)

	expectedErr := errors.New("failed to add certs")
	mockClient.EXPECT().
		AddCACerts(
			mock.Anything,
			namespace.AddCACertsParams{
				Namespace:        "test-namespace.test-account",
				Certs:            parsedCerts,
				ResourceVersion:  "",
				AsyncOperationID: "",
			},
		).
		Return(nil, expectedErr)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = "testdata/valid-cert.pem"
	cmd.Async = true
	cmd.Idempotent = false

	err = cmd.Command.RunE(&cmd.Command, []string{})
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestCloudNamespaceCertCaAddCommand_NothingToChange_Idempotent(t *testing.T) {
	validCertData, err := os.ReadFile("testdata/valid-cert.pem")
	require.NoError(t, err)
	parsedCerts, err := cert.ParseCACerts(validCertData)
	require.NoError(t, err)

	mockClient := cmdmock.NewMockNamespaceClient(t)

	nothingToChangeErr := status.Error(codes.InvalidArgument, "nothing to change")
	mockClient.EXPECT().
		AddCACerts(
			mock.Anything,
			namespace.AddCACertsParams{
				Namespace:        "test-namespace.test-account",
				Certs:            parsedCerts,
				ResourceVersion:  "",
				AsyncOperationID: "",
			},
		).
		Return(nil, nothingToChangeErr)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = "testdata/valid-cert.pem"
	cmd.Async = true
	cmd.Idempotent = true

	err = cmd.Command.RunE(&cmd.Command, []string{})
	require.NoError(t, err)

	var result struct {
		Status    string `json:"Status"`
		Namespace string `json:"Namespace"`
	}
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "unchanged", result.Status)
	assert.Equal(t, "test-namespace.test-account", result.Namespace)
}

func TestCloudNamespaceCertCaAddCommand_NothingToChange_NotIdempotent(t *testing.T) {
	validCertData, err := os.ReadFile("testdata/valid-cert.pem")
	require.NoError(t, err)
	parsedCerts, err := cert.ParseCACerts(validCertData)
	require.NoError(t, err)

	mockClient := cmdmock.NewMockNamespaceClient(t)

	nothingToChangeErr := status.Error(codes.InvalidArgument, "nothing to change")
	mockClient.EXPECT().
		AddCACerts(
			mock.Anything,
			namespace.AddCACertsParams{
				Namespace:        "test-namespace.test-account",
				Certs:            parsedCerts,
				ResourceVersion:  "",
				AsyncOperationID: "",
			},
		).
		Return(nil, nothingToChangeErr)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = "testdata/valid-cert.pem"
	cmd.Async = true
	cmd.Idempotent = false

	err = cmd.Command.RunE(&cmd.Command, []string{})
	require.Error(t, err)
	assert.Equal(t, nothingToChangeErr, err)
}

func TestCloudNamespaceCertCaAddCommand_IdempotentWithOtherError(t *testing.T) {
	validCertData, err := os.ReadFile("testdata/valid-cert.pem")
	require.NoError(t, err)
	parsedCerts, err := cert.ParseCACerts(validCertData)
	require.NoError(t, err)

	mockClient := cmdmock.NewMockNamespaceClient(t)

	otherErr := status.Error(codes.PermissionDenied, "permission denied")
	mockClient.EXPECT().
		AddCACerts(
			mock.Anything,
			namespace.AddCACertsParams{
				Namespace:        "test-namespace.test-account",
				Certs:            parsedCerts,
				ResourceVersion:  "",
				AsyncOperationID: "",
			},
		).
		Return(nil, otherErr)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = "testdata/valid-cert.pem"
	cmd.Async = true
	cmd.Idempotent = true

	err = cmd.Command.RunE(&cmd.Command, []string{})
	require.Error(t, err)
	assert.Equal(t, otherErr, err)
}

func TestCloudNamespaceCertCaAddCommand_PollingError(t *testing.T) {
	validCertData, err := os.ReadFile("testdata/valid-cert.pem")
	require.NoError(t, err)
	parsedCerts, err := cert.ParseCACerts(validCertData)
	require.NoError(t, err)

	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockPoller := cmdmock.NewMockPoller(t)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient.EXPECT().
		AddCACerts(
			mock.Anything,
			namespace.AddCACertsParams{
				Namespace:        "test-namespace.test-account",
				Certs:            parsedCerts,
				ResourceVersion:  "",
				AsyncOperationID: "",
			},
		).
		Return(expectedOp, nil)

	pollErr := errors.New("polling failed")
	mockPoller.EXPECT().
		Poll(mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(pollErr)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
		Poller:          mockPoller,
	}

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = "testdata/valid-cert.pem"
	cmd.Async = false

	err = cmd.Command.RunE(&cmd.Command, []string{})
	require.Error(t, err)
	assert.Equal(t, pollErr, err)
}

func TestCloudNamespaceCertCaAddCommand_GetNamespaceError(t *testing.T) {
	validCertData, err := os.ReadFile("testdata/valid-cert.pem")
	require.NoError(t, err)
	parsedCerts, err := cert.ParseCACerts(validCertData)
	require.NoError(t, err)

	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockPoller := cmdmock.NewMockPoller(t)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient.EXPECT().
		AddCACerts(
			mock.Anything,
			namespace.AddCACertsParams{
				Namespace:        "test-namespace.test-account",
				Certs:            parsedCerts,
				ResourceVersion:  "",
				AsyncOperationID: "",
			},
		).
		Return(expectedOp, nil)

	mockPoller.EXPECT().
		Poll(mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	getNamespaceErr := errors.New("failed to get namespace")
	mockClient.EXPECT().
		GetNamespace(mock.Anything, "test-namespace.test-account").
		Return(nil, getNamespaceErr)

	var buf bytes.Buffer
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
		Poller:          mockPoller,
	}

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = "testdata/valid-cert.pem"
	cmd.Async = false

	err = cmd.Command.RunE(&cmd.Command, []string{})
	require.Error(t, err)
	assert.Equal(t, getNamespaceErr, err)
}
