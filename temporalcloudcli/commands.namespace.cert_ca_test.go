package temporalcloudcli_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operationv1 "go.temporal.io/cloud-sdk/api/operation/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/internal/cert"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

// --- Test helpers ---

func buildTestAsyncOp(id string) *operationv1.AsyncOperation {
	return &operationv1.AsyncOperation{Id: id}
}

// generateTestCertificate creates a self-signed certificate for testing and returns the PEM-encoded bytes.
func generateTestCertificate(t *testing.T) []byte {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

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

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
}

// setupTestCertFile creates a temporary certificate file for testing.
// Returns the parsed certificates, the file path, and the raw cert data.
// The file will be automatically cleaned up when the test completes.
func setupTestCertFile(t *testing.T) ([]cert.CACert, string, []byte) {
	t.Helper()

	certData := generateTestCertificate(t)
	parsedCerts, err := cert.ParseCACerts(certData)
	if err != nil {
		t.Fatal(err)
	}

	tmpFile, err := os.CreateTemp("", "test-cert-*.pem")
	if err != nil {
		t.Fatal(err)
	}
	certPath := tmpFile.Name()
	t.Cleanup(func() {
		os.Remove(certPath)
	})

	if _, err = tmpFile.Write(certData); err != nil {
		t.Fatal(err)
	}
	if err = tmpFile.Close(); err != nil {
		t.Fatal(err)
	}

	return parsedCerts, certPath, certData
}

// buildTestNamespaceWithCerts returns a minimal Namespace proto with the given cert PEM data.
func buildTestNamespaceWithCerts(certPEM []byte) *namespacev1.Namespace {
	return &namespacev1.Namespace{
		Namespace:       "my-ns.my-account",
		ResourceVersion: "rv-1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: certPEM,
			},
		},
	}
}

// buildTestNamespaceEmpty returns a minimal Namespace proto with no certs.
func buildTestNamespaceEmpty() *namespacev1.Namespace {
	return &namespacev1.Namespace{
		Namespace:       "my-ns.my-account",
		ResourceVersion: "rv-1",
		Spec:            &namespacev1.NamespaceSpec{},
	}
}

// makeNsOpts is a shorthand for constructing NamespaceOptions.
func makeNsOpts(ns string) temporalcloudcli.NamespaceOptions {
	return temporalcloudcli.NamespaceOptions{Namespace: ns}
}

// --- CloudNamespaceCertCaList ---

func TestCloudNamespaceCertCaListCommand_Success(t *testing.T) {
	parsedCerts, _, certData := setupTestCertFile(t)

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaListCommand{
		NamespaceOptions: makeNsOpts("my-ns.my-account"),
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-account"}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{
					Namespace: buildTestNamespaceWithCerts(certData),
				}, nil)
		},
		JSONOutput:         true,
		ExpectedOutputJson: parsedCerts,
	})
}

func TestCloudNamespaceCertCaListCommand_Empty(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaListCommand{
		NamespaceOptions: makeNsOpts("my-ns.my-account"),
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-account"}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{
					Namespace: buildTestNamespaceEmpty(),
				}, nil)
		},
		JSONOutput:         true,
		ExpectedOutputJson: []cert.CACert{},
	})
}

func TestCloudNamespaceCertCaListCommand_GetNamespaceError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaListCommand{
		NamespaceOptions: makeNsOpts("my-ns.my-account"),
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("not found"))
		},
		ExpectedError: "not found",
	})
}

// --- CloudNamespaceCertCaCreate ---

func TestCloudNamespaceCertCaCreateCommand_SuccessWithFile(t *testing.T) {
	_, certPath, _ := setupTestCertFile(t)

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaCreateCommand{
		NamespaceOptions:       makeNsOpts("my-ns.my-account"),
		CaCertificateOptions:   temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath},
		AsyncOperationOptions:  temporalcloudcli.AsyncOperationOptions{Async: true},
		ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-1"},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-account"}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{
					Namespace: buildTestNamespaceEmpty(),
				}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
					return req.Namespace == "my-ns.my-account" &&
						req.ResourceVersion == "rv-1" &&
						len(req.Spec.GetMtlsAuth().GetAcceptedClientCa()) > 0
				}), mock.Anything).
				Return(&cloudservice.UpdateNamespaceResponse{
					AsyncOperation: buildTestAsyncOp("op-123"),
				}, nil)
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{
			AsyncOperationID: "op-123",
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    true,
		},
		JSONOutput: true,
	})
}

func TestCloudNamespaceCertCaCreateCommand_SuccessWithBase64(t *testing.T) {
	_, _, certData := setupTestCertFile(t)
	base64Cert := base64.StdEncoding.EncodeToString(certData)

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaCreateCommand{
		NamespaceOptions:      makeNsOpts("my-ns.my-account"),
		CaCertificateOptions:  temporalcloudcli.CaCertificateOptions{CaCertificate: base64Cert},
		AsyncOperationOptions: temporalcloudcli.AsyncOperationOptions{Async: true},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-account"}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{
					Namespace: buildTestNamespaceEmpty(),
				}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
					return req.Namespace == "my-ns.my-account"
				}), mock.Anything).
				Return(&cloudservice.UpdateNamespaceResponse{
					AsyncOperation: buildTestAsyncOp("op-456"),
				}, nil)
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{
			AsyncOperationID: "op-456",
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    true,
		},
		JSONOutput: true,
	})
}

func TestCloudNamespaceCertCaCreateCommand_NoCertFlag(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaCreateCommand{
		NamespaceOptions: makeNsOpts("my-ns.my-account"),
	}, temporalcloudcli.TestCommandOptions{
		ExpectedError: "either --ca-certificate-file or --ca-certificate must be provided",
	})
}

func TestCloudNamespaceCertCaCreateCommand_BothCertFlags(t *testing.T) {
	_, certPath, certData := setupTestCertFile(t)
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaCreateCommand{
		NamespaceOptions: makeNsOpts("my-ns.my-account"),
		CaCertificateOptions: temporalcloudcli.CaCertificateOptions{
			CaCertificateFile: certPath,
			CaCertificate:     base64.StdEncoding.EncodeToString(certData),
		},
	}, temporalcloudcli.TestCommandOptions{
		ExpectedError: "cannot specify both",
	})
}

func TestCloudNamespaceCertCaCreateCommand_InvalidBase64(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaCreateCommand{
		NamespaceOptions:     makeNsOpts("my-ns.my-account"),
		CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificate: "invalid!!!base64"},
	}, temporalcloudcli.TestCommandOptions{
		ExpectedError: "invalid base64 encoded certificate data",
	})
}

func TestCloudNamespaceCertCaCreateCommand_FileNotFound(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaCreateCommand{
		NamespaceOptions:     makeNsOpts("my-ns.my-account"),
		CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: "testdata/nonexistent-cert.pem"},
	}, temporalcloudcli.TestCommandOptions{
		ExpectedError: "no such file or directory",
	})
}

func TestCloudNamespaceCertCaCreateCommand_GetNamespaceError(t *testing.T) {
	_, certPath, _ := setupTestCertFile(t)

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaCreateCommand{
		NamespaceOptions:     makeNsOpts("my-ns.my-account"),
		CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("namespace not found"))
		},
		ExpectedError: "namespace not found",
	})
}

func TestCloudNamespaceCertCaCreateCommand_UpdateNamespaceError(t *testing.T) {
	_, certPath, _ := setupTestCertFile(t)

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaCreateCommand{
		NamespaceOptions:     makeNsOpts("my-ns.my-account"),
		CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{
					Namespace: buildTestNamespaceEmpty(),
				}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("update failed"))
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    true,
		},
		ExpectedError: "update failed",
	})
}

func TestCloudNamespaceCertCaCreateCommand_UserDeclines(t *testing.T) {
	_, certPath, _ := setupTestCertFile(t)

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaCreateCommand{
		NamespaceOptions:     makeNsOpts("my-ns.my-account"),
		CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{
					Namespace: buildTestNamespaceEmpty(),
				}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    false,
		},
		ExpectedError: "Aborting create.",
	})
}

// --- CloudNamespaceCertCaDelete ---

func TestCloudNamespaceCertCaDeleteCommand_SuccessWithFile(t *testing.T) {
	_, certPath, certData := setupTestCertFile(t)

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaDeleteCommand{
		NamespaceOptions:       makeNsOpts("my-ns.my-account"),
		CaCertificateOptions:   temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath},
		AsyncOperationOptions:  temporalcloudcli.AsyncOperationOptions{Async: true},
		ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: "rv-1"},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-account"}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{
					Namespace: buildTestNamespaceWithCerts(certData),
				}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
					return req.Namespace == "my-ns.my-account" &&
						req.ResourceVersion == "rv-1"
				}), mock.Anything).
				Return(&cloudservice.UpdateNamespaceResponse{
					AsyncOperation: buildTestAsyncOp("op-del-123"),
				}, nil)
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{
			AsyncOperationID: "op-del-123",
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    true,
		},
		JSONOutput: true,
	})
}

func TestCloudNamespaceCertCaDeleteCommand_SuccessWithBase64(t *testing.T) {
	_, _, certData := setupTestCertFile(t)
	base64Cert := base64.StdEncoding.EncodeToString(certData)

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaDeleteCommand{
		NamespaceOptions:      makeNsOpts("my-ns.my-account"),
		CaCertificateOptions:  temporalcloudcli.CaCertificateOptions{CaCertificate: base64Cert},
		AsyncOperationOptions: temporalcloudcli.AsyncOperationOptions{Async: true},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, &cloudservice.GetNamespaceRequest{Namespace: "my-ns.my-account"}, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{
					Namespace: buildTestNamespaceWithCerts(certData),
				}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
					return req.Namespace == "my-ns.my-account"
				}), mock.Anything).
				Return(&cloudservice.UpdateNamespaceResponse{
					AsyncOperation: buildTestAsyncOp("op-del-456"),
				}, nil)
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{
			AsyncOperationID: "op-del-456",
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    true,
		},
		JSONOutput: true,
	})
}

func TestCloudNamespaceCertCaDeleteCommand_NoCertFlag(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaDeleteCommand{
		NamespaceOptions: makeNsOpts("my-ns.my-account"),
	}, temporalcloudcli.TestCommandOptions{
		ExpectedError: "either --ca-certificate-file or --ca-certificate must be provided",
	})
}

func TestCloudNamespaceCertCaDeleteCommand_BothCertFlags(t *testing.T) {
	_, certPath, certData := setupTestCertFile(t)
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaDeleteCommand{
		NamespaceOptions: makeNsOpts("my-ns.my-account"),
		CaCertificateOptions: temporalcloudcli.CaCertificateOptions{
			CaCertificateFile: certPath,
			CaCertificate:     base64.StdEncoding.EncodeToString(certData),
		},
	}, temporalcloudcli.TestCommandOptions{
		ExpectedError: "cannot specify both",
	})
}

func TestCloudNamespaceCertCaDeleteCommand_InvalidBase64(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaDeleteCommand{
		NamespaceOptions:     makeNsOpts("my-ns.my-account"),
		CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificate: "invalid!!!base64"},
	}, temporalcloudcli.TestCommandOptions{
		ExpectedError: "invalid base64 encoded certificate data",
	})
}

func TestCloudNamespaceCertCaDeleteCommand_FileNotFound(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaDeleteCommand{
		NamespaceOptions:     makeNsOpts("my-ns.my-account"),
		CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: "testdata/nonexistent-cert.pem"},
	}, temporalcloudcli.TestCommandOptions{
		ExpectedError: "no such file or directory",
	})
}

func TestCloudNamespaceCertCaDeleteCommand_GetNamespaceError(t *testing.T) {
	_, certPath, _ := setupTestCertFile(t)

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaDeleteCommand{
		NamespaceOptions:     makeNsOpts("my-ns.my-account"),
		CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("namespace not found"))
		},
		ExpectedError: "namespace not found",
	})
}

func TestCloudNamespaceCertCaDeleteCommand_UpdateNamespaceError(t *testing.T) {
	_, certPath, certData := setupTestCertFile(t)

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaDeleteCommand{
		NamespaceOptions:     makeNsOpts("my-ns.my-account"),
		CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{
					Namespace: buildTestNamespaceWithCerts(certData),
				}, nil)
			c.EXPECT().
				UpdateNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("update failed"))
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    true,
		},
		ExpectedError: "update failed",
	})
}

func TestCloudNamespaceCertCaDeleteCommand_UserDeclines(t *testing.T) {
	_, certPath, certData := setupTestCertFile(t)

	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceCertCaDeleteCommand{
		NamespaceOptions:     makeNsOpts("my-ns.my-account"),
		CaCertificateOptions: temporalcloudcli.CaCertificateOptions{CaCertificateFile: certPath},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespace(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceResponse{
					Namespace: buildTestNamespaceWithCerts(certData),
				}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    false,
		},
		ExpectedError: "Aborting delete.",
	})
}
