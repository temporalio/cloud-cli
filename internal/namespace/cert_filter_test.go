package namespace_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cloud-cli/internal/namespace"
	nsmock "github.com/temporalio/cloud-cli/internal/namespace/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/protobuf/proto"
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
						CommonName:             "test.example.com",
						Organization:           "Example Corp",
						OrganizationalUnit:     "Engineering",
						SubjectAlternativeName: "*.example.com",
					},
					{
						CommonName:   "test2.example.com",
						Organization: "Another Corp",
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
	assert.Equal(t, []namespace.CertFilter{
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
	}, result)
}

func TestListCertFilters_NoMtlsAuth(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace: "test-namespace",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: nil,
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
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				CertificateFilters: []*namespacev1.CertificateFilterSpec{
					{
						CommonName:   "existing.example.com",
						Organization: "Existing Corp",
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
								CommonName:   "existing.example.com",
								Organization: "Existing Corp",
							},
							{
								CommonName:   "new.example.com",
								Organization: "New Corp",
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
		Filters: []namespace.CertFilter{
			{
				CommonName:   "new.example.com",
				Organization: "New Corp",
			},
		},
		AsyncOperationID: "test-async-op",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestAddCertFilters_NoExistingMtlsAuth(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: nil,
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
							{
								CommonName: "new.example.com",
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
		Filters: []namespace.CertFilter{
			{
				CommonName: "new.example.com",
			},
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
		Filters: []namespace.CertFilter{
			{CommonName: "test.example.com"},
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
		Filters: []namespace.CertFilter{
			{CommonName: "test.example.com"},
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
							{
								CommonName: "test.example.com",
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
		Filters: []namespace.CertFilter{
			{CommonName: "test.example.com"},
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
						CommonName:   "filter1.example.com",
						Organization: "Corp1",
					},
					{
						CommonName:   "filter2.example.com",
						Organization: "Corp2",
					},
					{
						CommonName:   "filter3.example.com",
						Organization: "Corp3",
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
								CommonName:   "filter1.example.com",
								Organization: "Corp1",
							},
							{
								CommonName:   "filter3.example.com",
								Organization: "Corp3",
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
		Filters: []namespace.CertFilter{
			{
				CommonName:   "filter2.example.com",
				Organization: "Corp2",
			},
		},
		AsyncOperationID: "test-async-op",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestDeleteCertFilters_NonExistentFilter(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				CertificateFilters: []*namespacev1.CertificateFilterSpec{
					{
						CommonName:   "existing.example.com",
						Organization: "Existing Corp",
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

	// Verify the existing filter is still present
	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			filters := req.GetSpec().GetMtlsAuth().GetCertificateFilters()
			return len(filters) == 1 && filters[0].CommonName == "existing.example.com"
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	// Try to delete a filter that doesn't exist - should succeed
	result, err := client.DeleteCertFilters(ctx, namespace.DeleteCertFiltersParams{
		Namespace: "test-namespace",
		Filters: []namespace.CertFilter{
			{
				CommonName:   "nonexistent.example.com",
				Organization: "Nonexistent Corp",
			},
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
		Filters: []namespace.CertFilter{
			{CommonName: "test.example.com"},
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
					{
						CommonName:   "test.example.com",
						Organization: "Test Corp",
					},
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
		Filters: []namespace.CertFilter{
			{
				CommonName:   "test.example.com",
				Organization: "Test Corp",
			},
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
					{
						CommonName:   "test.example.com",
						Organization: "Test Corp",
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
		Filters: []namespace.CertFilter{
			{
				CommonName:   "test.example.com",
				Organization: "Test Corp",
			},
		},
		ResourceVersion: "custom-version",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestDeleteCertFilters_MultipleFilters(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				CertificateFilters: []*namespacev1.CertificateFilterSpec{
					{
						CommonName:   "filter1.example.com",
						Organization: "Corp1",
					},
					{
						CommonName:   "filter2.example.com",
						Organization: "Corp2",
					},
					{
						CommonName:   "filter3.example.com",
						Organization: "Corp3",
					},
					{
						CommonName:   "filter4.example.com",
						Organization: "Corp4",
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

	// Delete filter2 and filter4, keep filter1 and filter3
	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			if req.Spec == nil || req.Spec.MtlsAuth == nil {
				return false
			}
			filters := req.Spec.MtlsAuth.CertificateFilters
			if len(filters) != 2 {
				return false
			}
			return filters[0].CommonName == "filter1.example.com" &&
				filters[1].CommonName == "filter3.example.com"
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteCertFilters(ctx, namespace.DeleteCertFiltersParams{
		Namespace: "test-namespace",
		Filters: []namespace.CertFilter{
			{
				CommonName:   "filter2.example.com",
				Organization: "Corp2",
			},
			{
				CommonName:   "filter4.example.com",
				Organization: "Corp4",
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestDeleteCertFilters_AllFiltersDeleted(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			MtlsAuth: &namespacev1.MtlsAuthSpec{
				CertificateFilters: []*namespacev1.CertificateFilterSpec{
					{
						CommonName:   "filter1.example.com",
						Organization: "Corp1",
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
			if req.Namespace != "test-namespace" ||
				req.ResourceVersion != "v1" {
				return false
			}
			// Verify that certificate filters list is empty
			filters := req.GetSpec().GetMtlsAuth().GetCertificateFilters()
			return len(filters) == 0
		})).
		Return(&cloudservice.UpdateNamespaceResponse{
			AsyncOperation: expectedOp,
		}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteCertFilters(ctx, namespace.DeleteCertFiltersParams{
		Namespace: "test-namespace",
		Filters: []namespace.CertFilter{
			{
				CommonName:   "filter1.example.com",
				Organization: "Corp1",
			},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}
