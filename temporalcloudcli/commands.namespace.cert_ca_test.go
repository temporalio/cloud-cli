package temporalcloudcli_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"testing"
	"time"

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
	parsedCerts, certPath := setupTestCertFile(t)

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
				cmd.CaCertificateFile = certPath
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
				cmd.CaCertificateFile = certPath
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

			var capturedErr error
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)

			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceCertCaAddCommand_InvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		setupCmd    func(*temporalcloudcli.CloudNamespaceCertCaAddCommand)
		assertError func(*testing.T, error)
	}{
		{
			name: "file not found",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaAddCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificateFile = "testdata/nonexistent-cert.pem"
				cmd.Async = true
			},
			assertError: func(t *testing.T, err error) {
				assert.True(t, os.IsNotExist(err))
			},
		},
		{
			name: "invalid certificate",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaAddCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificateFile = "testdata/invalid-cert.pem"
				cmd.Async = true
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "invalid certificate")
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
			}
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})

			require.Error(t, capturedErr)
			tt.assertError(t, capturedErr)
		})
	}
}

func TestCloudNamespaceCertCaAddCommand_AddCACertsError(t *testing.T) {
	parsedCerts, certPath := setupTestCertFile(t)

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
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = certPath
	cmd.Async = true
	cmd.Idempotent = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceCertCaAddCommand_NothingToChange(t *testing.T) {
	parsedCerts, certPath := setupTestCertFile(t)

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
			var capturedErr error
			cctx := &temporalcloudcli.CommandContext{
				Context:         context.Background(),
				Printer:         &printer.Printer{Output: &buf, JSON: true},
				NamespaceClient: mockClient,
			}
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
			cmd.Namespace = "test-namespace.test-account"
			cmd.CaCertificateFile = certPath
			cmd.Async = true
			cmd.Idempotent = tt.idempotent

			cmd.Command.Run(&cmd.Command, []string{})

			tt.assertResult(t, capturedErr, buf)
		})
	}
}

func TestCloudNamespaceCertCaAddCommand_IdempotentWithOtherError(t *testing.T) {
	parsedCerts, certPath := setupTestCertFile(t)

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
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = certPath
	cmd.Async = true
	cmd.Idempotent = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, otherErr, capturedErr)
}

func TestCloudNamespaceCertCaAddCommand_PollingError(t *testing.T) {
	parsedCerts, certPath := setupTestCertFile(t)

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
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
		Poller:          mockPoller,
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = certPath
	cmd.Async = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, pollErr, capturedErr)
}

func TestCloudNamespaceCertCaAddCommand_GetNamespaceError(t *testing.T) {
	parsedCerts, certPath := setupTestCertFile(t)

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
	var capturedErr error
	cctx := &temporalcloudcli.CommandContext{
		Context:         context.Background(),
		Printer:         &printer.Printer{Output: &buf, JSON: true},
		NamespaceClient: mockClient,
		Poller:          mockPoller,
	}
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaAddCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = certPath
	cmd.Async = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, getNamespaceErr, capturedErr)
}

// generateTestCertificate creates a self-signed certificate for testing and returns the PEM-encoded bytes.
func generateTestCertificate(t *testing.T) []byte {
	t.Helper()

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country:      []string{"US"},
			Province:     []string{"Washington"},
			Locality:     []string{"Bellevue"},
			Organization: []string{"Temporal"},
			CommonName:   "test.temporal.io",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	// Encode to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	return certPEM
}

// setupTestCertFile creates a temporary certificate file for testing.
// Returns the parsed certificates and the file path.
// The file will be automatically cleaned up when the test completes.
func setupTestCertFile(t *testing.T) ([]cert.CACert, string) {
	t.Helper()

	certData := generateTestCertificate(t)
	parsedCerts, err := cert.ParseCACerts(certData)
	require.NoError(t, err)

	tmpFile, err := os.CreateTemp("", "test-cert-*.pem")
	require.NoError(t, err)
	certPath := tmpFile.Name()
	t.Cleanup(func() {
		os.Remove(certPath)
	})

	_, err = tmpFile.Write(certData)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	return parsedCerts, certPath
}

func TestCloudNamespaceCertCaDeleteCommand_Success(t *testing.T) {
	parsedCerts, certPath := setupTestCertFile(t)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient := &cmdmock.MockNamespaceClient{}
	mockClient.On(
		"DeleteCACerts",
		mock.Anything,
		namespace.DeleteCACertsParams{
			Namespace:        "test-namespace.test-account",
			Certs:            parsedCerts,
			ResourceVersion:  "test-version",
			AsyncOperationID: "custom-operation-id",
		},
	).Return(expectedOp, nil)

	mockClient.On(
		"DeleteCACerts",
		mock.Anything,
		namespace.DeleteCACertsParams{
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
		setupCmd     func(*temporalcloudcli.CloudNamespaceCertCaDeleteCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificateFile = certPath
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
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificateFile = certPath
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
				RootCommand: &temporalcloudcli.CloudCommand{
					AutoConfirm: true,
				},
			}

			var capturedErr error
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCertCaDeleteCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)

			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceCertCaDeleteCommand_InvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		setupCmd    func(*temporalcloudcli.CloudNamespaceCertCaDeleteCommand)
		assertError func(*testing.T, error)
	}{
		{
			name: "file not found",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificateFile = "testdata/nonexistent-cert.pem"
				cmd.Async = true
			},
			assertError: func(t *testing.T, err error) {
				assert.True(t, os.IsNotExist(err))
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

			parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCertCaDeleteCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})

			require.Error(t, capturedErr)
			tt.assertError(t, capturedErr)
		})
	}
}

func TestCloudNamespaceCertCaDeleteCommand_DeleteCACertsError(t *testing.T) {
	parsedCerts, certPath := setupTestCertFile(t)

	mockClient := cmdmock.NewMockNamespaceClient(t)

	expectedErr := errors.New("failed to remove certs")
	mockClient.EXPECT().
		DeleteCACerts(
			mock.Anything,
			namespace.DeleteCACertsParams{
				Namespace:        "test-namespace.test-account",
				Certs:            parsedCerts,
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

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = certPath
	cmd.Async = true
	cmd.Idempotent = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceCertCaDeleteCommand_NothingToChange(t *testing.T) {
	parsedCerts, certPath := setupTestCertFile(t)

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
				DeleteCACerts(
					mock.Anything,
					namespace.DeleteCACertsParams{
						Namespace:        "test-namespace.test-account",
						Certs:            parsedCerts,
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

			parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCertCaDeleteCommand(cctx, parent)
			cmd.Namespace = "test-namespace.test-account"
			cmd.CaCertificateFile = certPath
			cmd.Async = true
			cmd.Idempotent = tt.idempotent

			cmd.Command.Run(&cmd.Command, []string{})

			tt.assertResult(t, capturedErr, buf)
		})
	}
}

func TestCloudNamespaceCertCaDeleteCommand_IdempotentWithOtherError(t *testing.T) {
	parsedCerts, certPath := setupTestCertFile(t)

	mockClient := cmdmock.NewMockNamespaceClient(t)

	otherErr := status.Error(codes.PermissionDenied, "permission denied")
	mockClient.EXPECT().
		DeleteCACerts(
			mock.Anything,
			namespace.DeleteCACertsParams{
				Namespace:        "test-namespace.test-account",
				Certs:            parsedCerts,
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

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = certPath
	cmd.Async = true
	cmd.Idempotent = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, otherErr, capturedErr)
}

func TestCloudNamespaceCertCaDeleteCommand_PollingError(t *testing.T) {
	parsedCerts, certPath := setupTestCertFile(t)

	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockPoller := cmdmock.NewMockPoller(t)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient.EXPECT().
		DeleteCACerts(
			mock.Anything,
			namespace.DeleteCACertsParams{
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

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = certPath
	cmd.Async = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, pollErr, capturedErr)
}

func TestCloudNamespaceCertCaDeleteCommand_GetNamespaceError(t *testing.T) {
	parsedCerts, certPath := setupTestCertFile(t)

	mockClient := cmdmock.NewMockNamespaceClient(t)
	mockPoller := cmdmock.NewMockPoller(t)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockClient.EXPECT().
		DeleteCACerts(
			mock.Anything,
			namespace.DeleteCACertsParams{
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

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = certPath
	cmd.Async = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, getNamespaceErr, capturedErr)
}

func TestCloudNamespaceCertCaDeleteCommand_UserDeclinesPrompt(t *testing.T) {
	_, certPath := setupTestCertFile(t)

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

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = certPath
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "Aborting delete")
}

func TestCloudNamespaceCertCaDeleteCommand_JSONOutputWithoutAutoConfirm(t *testing.T) {
	_, certPath := setupTestCertFile(t)

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

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = certPath
	cmd.Async = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Contains(t, capturedErr.Error(), "must bypass prompts when using JSON output")
}
