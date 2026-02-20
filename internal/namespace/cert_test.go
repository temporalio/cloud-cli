package namespace_test

import (
	"bytes"
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
	"google.golang.org/protobuf/proto"
)

func TestAddCACerts_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	// Generate test certificates
	existingCertPEM, _ := generateTestCert(t)
	_, newCert := generateTestCert(t)

	// Mock GetNamespace to return namespace with existing cert
	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: existingCertPEM,
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

	// Compute expected cert bundle
	newCertData, err := base64.StdEncoding.DecodeString(newCert.Base64EncodedData)
	require.NoError(t, err)
	expectedCertBundle := bytes.Join([][]byte{existingCertPEM, newCertData}, []byte("\n"))

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:        "test-namespace",
				ResourceVersion:  "v1",
				AsyncOperationId: "test-async-op",
				Spec: &namespacev1.NamespaceSpec{
					MtlsAuth: &namespacev1.MtlsAuthSpec{
						AcceptedClientCa: expectedCertBundle,
					},
				},
			}
			return proto.Equal(expected, req)
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

	// Mock GetNamespace
	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: existingCertPEM,
			},
		},
	}

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:       "test-namespace",
				ResourceVersion: "v1",
				Spec: &namespacev1.NamespaceSpec{
					MtlsAuth: &namespacev1.MtlsAuthSpec{
						AcceptedClientCa: existingCertPEM,
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	// Try to add the same certificate again
	result, err := client.AddCACerts(ctx, namespace.AddCACertsParams{
		Namespace: "test-namespace",
		Certs:     []cert.CACert{existingCert},
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
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

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: existingCertPEM,
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

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: existingCertPEM,
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	// Compute expected cert bundle
	newCertData, err := base64.StdEncoding.DecodeString(newCert.Base64EncodedData)
	require.NoError(t, err)
	expectedCertBundle := bytes.Join([][]byte{existingCertPEM, newCertData}, []byte("\n"))

	// Verify that custom resource version is used
	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:       "test-namespace",
				ResourceVersion: "custom-version",
				Spec: &namespacev1.NamespaceSpec{
					MtlsAuth: &namespacev1.MtlsAuthSpec{
						AcceptedClientCa: expectedCertBundle,
					},
				},
			}
			return proto.Equal(expected, req)
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

func TestDeleteCACerts_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	// Generate test certificates
	cert1PEM, _ := generateTestCert(t)
	cert2PEM, cert2 := generateTestCert(t)
	cert3PEM, _ := generateTestCert(t)

	// Combine certificates into a bundle (all 3 exist initially)
	certBundle := bytes.Join([][]byte{cert1PEM, cert2PEM, cert3PEM}, []byte{'\n'})

	// Mock GetNamespace to return namespace with all 3 certs
	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: certBundle,
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

	// Should have 2 certificates now (cert1 and cert3, cert2 was deleted)
	expectedCertBundle := bytes.Join([][]byte{cert1PEM, cert3PEM}, []byte{'\n'})

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:        "test-namespace",
				ResourceVersion:  "v1",
				AsyncOperationId: "test-async-op",
				Spec: &namespacev1.NamespaceSpec{
					MtlsAuth: &namespacev1.MtlsAuthSpec{
						AcceptedClientCa: expectedCertBundle,
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteCACerts(ctx, namespace.DeleteCACertsParams{
		Namespace:        "test-namespace",
		Certs:            []cert.CACert{cert2},
		ResourceVersion:  "",
		AsyncOperationID: "test-async-op",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestDeleteCACerts_NonExistentCertificate(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	// Generate test certificates
	existingCertPEM, _ := generateTestCert(t)
	_, nonExistentCert := generateTestCert(t)

	// Mock GetNamespace
	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: existingCertPEM,
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	// Verify the existing cert is still present
	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:       "test-namespace",
				ResourceVersion: "v1",
				Spec: &namespacev1.NamespaceSpec{
					MtlsAuth: &namespacev1.MtlsAuthSpec{
						AcceptedClientCa: existingCertPEM,
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	// Try to delete a certificate that doesn't exist - should succeed
	result, err := client.DeleteCACerts(ctx, namespace.DeleteCACertsParams{
		Namespace: "test-namespace",
		Certs:     []cert.CACert{nonExistentCert},
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestDeleteCACerts_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	_, certToDelete := generateTestCert(t)

	result, err := client.DeleteCACerts(ctx, namespace.DeleteCACertsParams{
		Namespace: "test-namespace",
		Certs:     []cert.CACert{certToDelete},
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestDeleteCACerts_UpdateNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	// Generate test certificates
	existingCertPEM, existingCert := generateTestCert(t)
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

	result, err := client.DeleteCACerts(ctx, namespace.DeleteCACertsParams{
		Namespace: "test-namespace",
		Certs:     []cert.CACert{existingCert},
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestDeleteCACerts_CustomResourceVersion(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingCertPEM, existingCert := generateTestCert(t)
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
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:       "test-namespace",
				ResourceVersion: "custom-version",
				Spec: &namespacev1.NamespaceSpec{
					MtlsAuth: nil,
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteCACerts(ctx, namespace.DeleteCACertsParams{
		Namespace:       "test-namespace",
		Certs:           []cert.CACert{existingCert},
		ResourceVersion: "custom-version",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestDeleteCACerts_MultipleFromBundle(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	// Generate test certificates
	cert1PEM, _ := generateTestCert(t)
	cert2PEM, cert2 := generateTestCert(t)
	cert3PEM, _ := generateTestCert(t)
	cert4PEM, cert4 := generateTestCert(t)

	// Combine certificates into a bundle (all 4 exist initially)
	certBundle := bytes.Join([][]byte{cert1PEM, cert2PEM, cert3PEM, cert4PEM}, []byte{'\n'})

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: certBundle,
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	// Delete cert2 and cert4, keep cert1 and cert3
	expectedCertBundle := bytes.Join([][]byte{cert1PEM, cert3PEM}, []byte{'\n'})
	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:       "test-namespace",
				ResourceVersion: "v1",
				Spec: &namespacev1.NamespaceSpec{
					MtlsAuth: &namespacev1.MtlsAuthSpec{
						AcceptedClientCa: expectedCertBundle,
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteCACerts(ctx, namespace.DeleteCACertsParams{
		Namespace: "test-namespace",
		Certs:     []cert.CACert{cert2, cert4},
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestDeleteCACerts_AllCertsDeleted(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	// Generate test certificate
	existingCertPEM, existingCert := generateTestCert(t)

	// Mock GetNamespace to return namespace with one cert
	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				AcceptedClientCa: existingCertPEM,
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
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:       "test-namespace",
				ResourceVersion: "v1",
				Spec: &namespacev1.NamespaceSpec{
					MtlsAuth: nil,
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteCACerts(ctx, namespace.DeleteCACertsParams{
		Namespace: "test-namespace",
		Certs:     []cert.CACert{existingCert},
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

// generateTestCert generates a test certificate and returns its PEM-encoded bytes and parsed representation.
// The returned PEM bytes are trimmed to match the behavior of ParseCACerts, which trims whitespace.
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

	certPEM = bytes.TrimSpace(certPEM)

	parsedCerts, err := cert.ParseCACerts(certPEM)
	require.NoError(t, err)
	require.Len(t, parsedCerts, 1)

	return certPEM, parsedCerts[0]
}
