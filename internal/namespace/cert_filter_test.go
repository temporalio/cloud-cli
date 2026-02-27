package namespace_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/internal/namespace"
	nsmock "github.com/temporalio/cloud-cli/internal/namespace/mock"
)

func TestListCertFilters_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace: "test-namespace",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				CertificateFilters: []*namespacev1.CertificateFilterSpec{
					{
						CommonName:   "test-cn",
						Organization: "test-org",
					},
				},
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.ListCertFilters(ctx, "test-namespace")

	require.NoError(t, err)
	require.Len(t, result, 1)
	expected := &namespacev1.CertificateFilterSpec{
		CommonName:   "test-cn",
		Organization: "test-org",
	}
	assert.True(t, proto.Equal(expected, result[0]))
}

func TestListCertFilters_EmptyList(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace: "test-namespace",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.ListCertFilters(ctx, "test-namespace")

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListCertFilters_NilMtlsAuth(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace: "test-namespace",
		Spec:      &namespacev1.NamespaceSpec{},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.ListCertFilters(ctx, "test-namespace")

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListCertFilters_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.ListCertFilters(ctx, "test-namespace")

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, result)
}

func TestAddCertFilters_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:        "test-namespace",
				ResourceVersion:  "v1",
				AsyncOperationId: "test-async-op",
				Spec: &namespacev1.NamespaceSpec{
					MtlsAuth: &namespacev1.MtlsAuthSpec{
						CertificateFilters: []*namespacev1.CertificateFilterSpec{
							{
								CommonName:   "new-cn",
								Organization: "new-org",
							},
						},
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.AddCertFilters(ctx, namespace.AddCertFiltersParams{
		Namespace: "test-namespace",
		Filters: []*namespacev1.CertificateFilterSpec{
			{
				CommonName:   "new-cn",
				Organization: "new-org",
			},
		},
		AsyncOperationID: "test-async-op",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestAddCertFilters_AppendsToExisting(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				CertificateFilters: []*namespacev1.CertificateFilterSpec{
					{CommonName: "existing-cn"},
				},
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:       "test-namespace",
				ResourceVersion: "v1",
				Spec: &namespacev1.NamespaceSpec{
					MtlsAuth: &namespacev1.MtlsAuthSpec{
						CertificateFilters: []*namespacev1.CertificateFilterSpec{
							{CommonName: "existing-cn"},
							{CommonName: "new-cn"},
						},
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.AddCertFilters(ctx, namespace.AddCertFiltersParams{
		Namespace: "test-namespace",
		Filters: []*namespacev1.CertificateFilterSpec{
			{CommonName: "new-cn"},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestAddCertFilters_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.AddCertFilters(ctx, namespace.AddCertFiltersParams{
		Namespace: "test-namespace",
		Filters: []*namespacev1.CertificateFilterSpec{
			{CommonName: "new-cn"},
		},
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestAddCertFilters_UpdateNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{},
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

	result, err := client.AddCertFilters(ctx, namespace.AddCertFiltersParams{
		Namespace: "test-namespace",
		Filters: []*namespacev1.CertificateFilterSpec{
			{CommonName: "new-cn"},
		},
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestAddCertFilters_CustomResourceVersion(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{},
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
					MtlsAuth: &namespacev1.MtlsAuthSpec{
						CertificateFilters: []*namespacev1.CertificateFilterSpec{
							{CommonName: "new-cn"},
						},
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.AddCertFilters(ctx, namespace.AddCertFiltersParams{
		Namespace: "test-namespace",
		Filters: []*namespacev1.CertificateFilterSpec{
			{CommonName: "new-cn"},
		},
		ResourceVersion: "custom-version",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestDeleteCertFilters_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				CertificateFilters: []*namespacev1.CertificateFilterSpec{
					{
						CommonName:   "cn-to-keep",
						Organization: "org-to-keep",
					},
					{
						CommonName:   "cn-to-remove",
						Organization: "org-to-remove",
					},
				},
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:        "test-namespace",
				ResourceVersion:  "v1",
				AsyncOperationId: "test-async-op",
				Spec: &namespacev1.NamespaceSpec{
					MtlsAuth: &namespacev1.MtlsAuthSpec{
						CertificateFilters: []*namespacev1.CertificateFilterSpec{
							{
								CommonName:   "cn-to-keep",
								Organization: "org-to-keep",
							},
						},
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteCertFilters(ctx, namespace.DeleteCertFiltersParams{
		Namespace: "test-namespace",
		Filters: []*namespacev1.CertificateFilterSpec{
			{
				CommonName:   "cn-to-remove",
				Organization: "org-to-remove",
			},
		},
		AsyncOperationID: "test-async-op",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestDeleteCertFilters_RemovesAllFilters(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				CertificateFilters: []*namespacev1.CertificateFilterSpec{
					{CommonName: "only-cn"},
				},
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{
		Id: "test-operation-id",
	}

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:       "test-namespace",
				ResourceVersion: "v1",
				Spec: &namespacev1.NamespaceSpec{
					MtlsAuth: &namespacev1.MtlsAuthSpec{
						CertificateFilters: []*namespacev1.CertificateFilterSpec{},
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteCertFilters(ctx, namespace.DeleteCertFiltersParams{
		Namespace: "test-namespace",
		Filters: []*namespacev1.CertificateFilterSpec{
			{CommonName: "only-cn"},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestDeleteCertFilters_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteCertFilters(ctx, namespace.DeleteCertFiltersParams{
		Namespace: "test-namespace",
		Filters: []*namespacev1.CertificateFilterSpec{
			{CommonName: "cn-to-remove"},
		},
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestDeleteCertFilters_UpdateNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				CertificateFilters: []*namespacev1.CertificateFilterSpec{
					{CommonName: "existing-cn"},
				},
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

	result, err := client.DeleteCertFilters(ctx, namespace.DeleteCertFiltersParams{
		Namespace: "test-namespace",
		Filters: []*namespacev1.CertificateFilterSpec{
			{CommonName: "existing-cn"},
		},
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestDeleteCertFilters_CustomResourceVersion(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				CertificateFilters: []*namespacev1.CertificateFilterSpec{
					{CommonName: "cn-to-remove"},
				},
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
					MtlsAuth: &namespacev1.MtlsAuthSpec{
						CertificateFilters: []*namespacev1.CertificateFilterSpec{},
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteCertFilters(ctx, namespace.DeleteCertFiltersParams{
		Namespace: "test-namespace",
		Filters: []*namespacev1.CertificateFilterSpec{
			{CommonName: "cn-to-remove"},
		},
		ResourceVersion: "custom-version",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}
