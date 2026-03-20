package temporalcloudcli_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
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
	"go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCloudNamespaceCertCaCreateCommand_Success(t *testing.T) {
	parsedCerts, certPath, certData := setupTestCertFile(t)
	base64Cert := base64.StdEncoding.EncodeToString(certData)

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
	mockPoller.On("PollAsyncOperation", mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceCertCaCreateCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async with file",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaCreateCommand) {
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
			name: "sync with file",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificateFile = certPath
				cmd.Async = false
			},
			assertResult: func(t *testing.T, buf bytes.Buffer) {
				assert.Empty(t, buf.String())
			},
		},
		{
			name: "async with base64",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificate = base64Cert
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
			name: "sync with base64",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificate = base64Cert
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
			cctx.Options.Fail = func(err error) {
				capturedErr = err
			}

			parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
			cmd := temporalcloudcli.NewCloudNamespaceCertCaCreateCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})
			require.NoError(t, capturedErr)

			tt.assertResult(t, buf)
		})
	}
}

func TestCloudNamespaceCertCaCreateCommand_InvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		setupCmd    func(*temporalcloudcli.CloudNamespaceCertCaCreateCommand)
		assertError func(*testing.T, error)
	}{
		{
			name: "file not found",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificateFile = "testdata/nonexistent-cert.pem"
				cmd.Async = true
			},
			assertError: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, os.ErrNotExist))
			},
		},
		{
			name: "invalid certificate",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificateFile = "testdata/invalid-cert.pem"
				cmd.Async = true
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "invalid certificate")
			},
		},
		{
			name: "both flags provided",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificateFile = "testdata/cert.pem"
				cmd.CaCertificate = "base64data"
				cmd.Async = true
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "cannot specify both")
			},
		},
		{
			name: "neither flag provided",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Async = true
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "either --ca-certificate-file or --ca-certificate must be provided")
			},
		},
		{
			name: "invalid base64",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaCreateCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificate = "invalid!!!base64"
				cmd.Async = true
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "invalid base64 encoded certificate data")
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
			cmd := temporalcloudcli.NewCloudNamespaceCertCaCreateCommand(cctx, parent)
			tt.setupCmd(cmd)

			cmd.Command.Run(&cmd.Command, []string{})

			require.Error(t, capturedErr)
			tt.assertError(t, capturedErr)
		})
	}
}

func TestCloudNamespaceCertCaCreateCommand_AddCACertsError(t *testing.T) {
	parsedCerts, certPath, _ := setupTestCertFile(t)

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
	cmd := temporalcloudcli.NewCloudNamespaceCertCaCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = certPath
	cmd.Async = true
	cmd.Idempotent = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, expectedErr, capturedErr)
}

func TestCloudNamespaceCertCaCreateCommand_NothingToChange(t *testing.T) {
	parsedCerts, certPath, _ := setupTestCertFile(t)

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
				assert.Equal(t, "unchanged", result.Status)
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
			cmd := temporalcloudcli.NewCloudNamespaceCertCaCreateCommand(cctx, parent)
			cmd.Namespace = "test-namespace.test-account"
			cmd.CaCertificateFile = certPath
			cmd.Async = true
			cmd.Idempotent = tt.idempotent

			cmd.Command.Run(&cmd.Command, []string{})

			tt.assertResult(t, capturedErr, buf)
		})
	}
}

func TestCloudNamespaceCertCaCreateCommand_IdempotentWithOtherError(t *testing.T) {
	parsedCerts, certPath, _ := setupTestCertFile(t)

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
	cmd := temporalcloudcli.NewCloudNamespaceCertCaCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = certPath
	cmd.Async = true
	cmd.Idempotent = true

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, otherErr, capturedErr)
}

func TestCloudNamespaceCertCaCreateCommand_PollingError(t *testing.T) {
	parsedCerts, certPath, _ := setupTestCertFile(t)

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
	cctx.Options.Fail = func(err error) {
		capturedErr = err
	}

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaCreateCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = certPath
	cmd.Async = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, pollErr, capturedErr)
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
// Returns the parsed certificates, the file path, and the raw cert data.
// The file will be automatically cleaned up when the test completes.
func setupTestCertFile(t *testing.T) ([]cert.CACert, string, []byte) {
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

	return parsedCerts, certPath, certData
}

func TestCloudNamespaceCertCaDeleteCommand_Success(t *testing.T) {
	parsedCerts, certPath, certData := setupTestCertFile(t)
	base64Cert := base64.StdEncoding.EncodeToString(certData)

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
	mockPoller.On("PollAsyncOperation", mock.Anything, "test-operation-id", "test-namespace.test-account").
		Return(nil)

	tests := []struct {
		name         string
		setupCmd     func(*temporalcloudcli.CloudNamespaceCertCaDeleteCommand)
		assertResult func(*testing.T, bytes.Buffer)
	}{
		{
			name: "async with file",
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
			name: "sync with file",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificateFile = certPath
				cmd.Async = false
			},
			assertResult: func(t *testing.T, buf bytes.Buffer) {
				assert.Empty(t, buf.String())
			},
		},
		{
			name: "async with base64",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificate = base64Cert
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
			name: "sync with base64",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificate = base64Cert
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
				assert.True(t, errors.Is(err, os.ErrNotExist))
			},
		},
		{
			name: "both flags provided",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificateFile = "testdata/cert.pem"
				cmd.CaCertificate = "base64data"
				cmd.Async = true
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "cannot specify both")
			},
		},
		{
			name: "neither flag provided",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.Async = true
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "either --ca-certificate-file or --ca-certificate must be provided")
			},
		},
		{
			name: "invalid base64",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaDeleteCommand) {
				cmd.Namespace = "test-namespace.test-account"
				cmd.CaCertificate = "invalid!!!base64"
				cmd.Async = true
			},
			assertError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "invalid base64 encoded certificate data")
			},
		},
		{
			name: "invalid certificate",
			setupCmd: func(cmd *temporalcloudcli.CloudNamespaceCertCaDeleteCommand) {
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
	parsedCerts, certPath, _ := setupTestCertFile(t)

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
	parsedCerts, certPath, _ := setupTestCertFile(t)

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
				assert.Equal(t, "unchanged", result.Status)
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
	parsedCerts, certPath, _ := setupTestCertFile(t)

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
	parsedCerts, certPath, _ := setupTestCertFile(t)

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

	parent := &temporalcloudcli.CloudNamespaceCertCaCommand{}
	cmd := temporalcloudcli.NewCloudNamespaceCertCaDeleteCommand(cctx, parent)
	cmd.Namespace = "test-namespace.test-account"
	cmd.CaCertificateFile = certPath
	cmd.Async = false

	cmd.Command.Run(&cmd.Command, []string{})
	require.Error(t, capturedErr)
	assert.Equal(t, pollErr, capturedErr)
}

func TestCloudNamespaceCertCaDeleteCommand_UserDeclinesPrompt(t *testing.T) {
	_, certPath, _ := setupTestCertFile(t)

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
	_, certPath, _ := setupTestCertFile(t)

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
