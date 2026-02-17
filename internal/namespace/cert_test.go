package namespace_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cloud-cli/internal/cert"
	"github.com/temporalio/cloud-cli/internal/namespace"
	nsmock "github.com/temporalio/cloud-cli/internal/namespace/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"go.temporal.io/cloud-sdk/api/operation/v1"
)

func TestAddCACerts_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	// Generate test certificates
	existingCertPEM, existingCert := generateTestCert(t)
	_, newCert := generateTestCert(t)

	// Encode existing cert as base64 (as it would be stored)
	existingCertBase64 := base64.StdEncoding.EncodeToString(existingCertPEM)

	// Mock GetNamespace to return namespace with existing cert
	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: []byte(existingCertBase64),
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	// Mock UpdateNamespace
	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			if req.Namespace != "test-namespace" ||
				req.ResourceVersion != "v1" ||
				req.AsyncOperationId != "test-async-op" {
				return false
			}
			// Verify the cert bundle contains both existing and new certificates
			if req.Spec == nil || req.Spec.MtlsAuth == nil {
				return false
			}
			bundleData := req.Spec.MtlsAuth.AcceptedClientCa
			parsedBundle, err := cert.ParseCACerts(bundleData)
			if err != nil {
				return false
			}
			// Should have 2 certificates now (existing + new)
			if len(parsedBundle) != 2 {
				return false
			}
			// Check that both fingerprints are present
			fingerprints := make(map[string]bool)
			for _, c := range parsedBundle {
				fingerprints[c.Fingerprint] = true
			}
			return fingerprints[existingCert.Fingerprint] && fingerprints[newCert.Fingerprint]
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.AddCACerts(ctx, namespace.AddCACertsParams{
		Namespace:        "test-namespace",
		Certs:            []cert.CACert{newCert},
		ResourceVersion:  "",
		AsyncOperationID: "test-async-op",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestAddCACerts_DuplicateCertificate(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	// Generate test certificate
	existingCertPEM, existingCert := generateTestCert(t)
	existingCertBase64 := base64.StdEncoding.EncodeToString(existingCertPEM)

	// Mock GetNamespace
	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: []byte(existingCertBase64),
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	// Try to add the same certificate again
	result, err := client.AddCACerts(ctx, namespace.AddCACertsParams{
		Namespace: "test-namespace",
		Certs:     []cert.CACert{existingCert},
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "already exists")
	assert.Contains(t, err.Error(), existingCert.Fingerprint)
}

func TestAddCACerts_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	_, newCert := generateTestCert(t)

	result, err := client.AddCACerts(ctx, namespace.AddCACertsParams{
		Namespace: "test-namespace",
		Certs:     []cert.CACert{newCert},
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestAddCACerts_UpdateNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	// Generate test certificates
	existingCertPEM, _ := generateTestCert(t)
	_, newCert := generateTestCert(t)

	existingCertBase64 := base64.StdEncoding.EncodeToString(existingCertPEM)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: []byte(existingCertBase64),
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedErr := errors.New("update failed")
	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.Anything).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.AddCACerts(ctx, namespace.AddCACertsParams{
		Namespace: "test-namespace",
		Certs:     []cert.CACert{newCert},
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestAddCACerts_CustomResourceVersion(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingCertPEM, _ := generateTestCert(t)
	_, newCert := generateTestCert(t)

	existingCertBase64 := base64.StdEncoding.EncodeToString(existingCertPEM)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: []byte(existingCertBase64),
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	// Verify that custom resource version is used
	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			return req.ResourceVersion == "custom-version"
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.AddCACerts(ctx, namespace.AddCACertsParams{
		Namespace:       "test-namespace",
		Certs:           []cert.CACert{newCert},
		ResourceVersion: "custom-version",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestListCACerts_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	// Generate test certificates
	cert1PEM, cert1 := generateTestCert(t)
	cert2PEM, cert2 := generateTestCert(t)

	// Combine certificates into a bundle
	certBundle := append(cert1PEM, '\n')
	certBundle = append(certBundle, cert2PEM...)

	ns := &namespacev1.Namespace{
		Namespace: "test-namespace",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: certBundle,
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.ListCACerts(ctx, "test-namespace")

	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, []cert.CACert{cert1, cert2}, result)
}

func TestListCACerts_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.ListCACerts(ctx, "test-namespace")

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, result)
}

func TestListCACerts_EmptyBundle(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	// Return namespace with empty certificate bundle
	ns := &namespacev1.Namespace{
		Namespace: "test-namespace",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: []byte{},
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.ListCACerts(ctx, "test-namespace")

	require.NoError(t, err)
	assert.Empty(t, result)
}

// generateTestCert generates a test certificate and returns its PEM-encoded bytes and parsed representation.
func generateTestCert(t *testing.T) ([]byte, cert.CACert) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

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
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	parsedCerts, err := cert.ParseCACerts(certPEM)
	require.NoError(t, err)
	require.Len(t, parsedCerts, 1)

	return certPEM, parsedCerts[0]
}
