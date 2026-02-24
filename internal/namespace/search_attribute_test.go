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

func TestListSearchAttributes_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace: "test-namespace",
		Spec: &namespacev1.NamespaceSpec{
			SearchAttributes: map[string]namespacev1.NamespaceSpec_SearchAttributeType{
				"CustomKeyword": namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
				"CustomText":    namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_TEXT,
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.ListSearchAttributes(ctx, "test-namespace")

	require.NoError(t, err)
	require.Len(t, result, 2)

	// Map iteration order is non-deterministic, so compare via a map
	resultMap := make(map[string]namespacev1.NamespaceSpec_SearchAttributeType)
	for _, attr := range result {
		resultMap[attr.Name] = attr.Type
	}
	assert.Equal(t, namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD, resultMap["CustomKeyword"])
	assert.Equal(t, namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_TEXT, resultMap["CustomText"])
}

func TestListSearchAttributes_EmptyList(t *testing.T) {
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

	result, err := client.ListSearchAttributes(ctx, "test-namespace")

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListSearchAttributes_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.ListSearchAttributes(ctx, "test-namespace")

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, result)
}

func TestCreateSearchAttribute_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			SearchAttributes: map[string]namespacev1.NamespaceSpec_SearchAttributeType{
				"ExistingField": namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{Id: "test-op-id"}

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:        "test-namespace",
				ResourceVersion:  "v1",
				AsyncOperationId: "test-async-op",
				Spec: &namespacev1.NamespaceSpec{
					SearchAttributes: map[string]namespacev1.NamespaceSpec_SearchAttributeType{
						"ExistingField": namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
						"NewField":      namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_TEXT,
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.CreateSearchAttribute(ctx, namespace.CreateSearchAttributeParams{
		Namespace:        "test-namespace",
		Name:             "NewField",
		Type:             namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_TEXT,
		AsyncOperationID: "test-async-op",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestCreateSearchAttribute_NilSearchAttributes(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec:            &namespacev1.NamespaceSpec{},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{Id: "test-op-id"}

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:       "test-namespace",
				ResourceVersion: "v1",
				Spec: &namespacev1.NamespaceSpec{
					SearchAttributes: map[string]namespacev1.NamespaceSpec_SearchAttributeType{
						"NewField": namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.CreateSearchAttribute(ctx, namespace.CreateSearchAttributeParams{
		Namespace: "test-namespace",
		Name:      "NewField",
		Type:      namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestCreateSearchAttribute_DuplicateName(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			SearchAttributes: map[string]namespacev1.NamespaceSpec_SearchAttributeType{
				"ExistingField": namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.CreateSearchAttribute(ctx, namespace.CreateSearchAttributeParams{
		Namespace: "test-namespace",
		Name:      "ExistingField",
		Type:      namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_TEXT,
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "already exists")
	assert.Contains(t, err.Error(), "ExistingField")
}

func TestCreateSearchAttribute_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.CreateSearchAttribute(ctx, namespace.CreateSearchAttributeParams{
		Namespace: "test-namespace",
		Name:      "NewField",
		Type:      namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestCreateSearchAttribute_UpdateNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec:            &namespacev1.NamespaceSpec{},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedErr := errors.New("update failed")
	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.Anything).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.CreateSearchAttribute(ctx, namespace.CreateSearchAttributeParams{
		Namespace: "test-namespace",
		Name:      "NewField",
		Type:      namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestCreateSearchAttribute_CustomResourceVersion(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec:            &namespacev1.NamespaceSpec{},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{Id: "test-op-id"}

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:       "test-namespace",
				ResourceVersion: "custom-version",
				Spec: &namespacev1.NamespaceSpec{
					SearchAttributes: map[string]namespacev1.NamespaceSpec_SearchAttributeType{
						"NewField": namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.CreateSearchAttribute(ctx, namespace.CreateSearchAttributeParams{
		Namespace:       "test-namespace",
		Name:            "NewField",
		Type:            namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
		ResourceVersion: "custom-version",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestRenameSearchAttribute_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			SearchAttributes: map[string]namespacev1.NamespaceSpec_SearchAttributeType{
				"OldName": namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{Id: "test-op-id"}

	mockCloud.EXPECT().
		RenameCustomSearchAttribute(ctx, mock.MatchedBy(func(req *cloudservice.RenameCustomSearchAttributeRequest) bool {
			expected := &cloudservice.RenameCustomSearchAttributeRequest{
				Namespace:                         "test-namespace",
				ExistingCustomSearchAttributeName: "OldName",
				NewCustomSearchAttributeName:      "NewName",
				ResourceVersion:                   "v1",
				AsyncOperationId:                  "test-async-op",
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.RenameCustomSearchAttributeResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.RenameSearchAttribute(ctx, namespace.RenameSearchAttributeParams{
		Namespace:                         "test-namespace",
		ExistingCustomSearchAttributeName: "OldName",
		NewCustomSearchAttributeName:      "NewName",
		AsyncOperationID:                  "test-async-op",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestRenameSearchAttribute_NonExistentAttribute(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			SearchAttributes: map[string]namespacev1.NamespaceSpec_SearchAttributeType{
				"ExistingField": namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.RenameSearchAttribute(ctx, namespace.RenameSearchAttributeParams{
		Namespace:                         "test-namespace",
		ExistingCustomSearchAttributeName: "NonExistent",
		NewCustomSearchAttributeName:      "NewName",
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "does not exist")
	assert.Contains(t, err.Error(), "NonExistent")
}

func TestRenameSearchAttribute_TargetAlreadyExists(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			SearchAttributes: map[string]namespacev1.NamespaceSpec_SearchAttributeType{
				"OldName":     namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
				"ExistingNew": namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_TEXT,
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.RenameSearchAttribute(ctx, namespace.RenameSearchAttributeParams{
		Namespace:                         "test-namespace",
		ExistingCustomSearchAttributeName: "OldName",
		NewCustomSearchAttributeName:      "ExistingNew",
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "already exists")
	assert.Contains(t, err.Error(), "ExistingNew")
}

func TestRenameSearchAttribute_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.RenameSearchAttribute(ctx, namespace.RenameSearchAttributeParams{
		Namespace:                         "test-namespace",
		ExistingCustomSearchAttributeName: "OldName",
		NewCustomSearchAttributeName:      "NewName",
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestRenameSearchAttribute_RenameError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			SearchAttributes: map[string]namespacev1.NamespaceSpec_SearchAttributeType{
				"OldName": namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedErr := errors.New("rename failed")
	mockCloud.EXPECT().
		RenameCustomSearchAttribute(ctx, mock.Anything).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.RenameSearchAttribute(ctx, namespace.RenameSearchAttributeParams{
		Namespace:                         "test-namespace",
		ExistingCustomSearchAttributeName: "OldName",
		NewCustomSearchAttributeName:      "NewName",
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestRenameSearchAttribute_CustomResourceVersion(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			SearchAttributes: map[string]namespacev1.NamespaceSpec_SearchAttributeType{
				"OldName": namespacev1.NamespaceSpec_SEARCH_ATTRIBUTE_TYPE_KEYWORD,
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{Id: "test-op-id"}

	mockCloud.EXPECT().
		RenameCustomSearchAttribute(ctx, mock.MatchedBy(func(req *cloudservice.RenameCustomSearchAttributeRequest) bool {
			expected := &cloudservice.RenameCustomSearchAttributeRequest{
				Namespace:                         "test-namespace",
				ExistingCustomSearchAttributeName: "OldName",
				NewCustomSearchAttributeName:      "NewName",
				ResourceVersion:                   "custom-version",
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.RenameCustomSearchAttributeResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.RenameSearchAttribute(ctx, namespace.RenameSearchAttributeParams{
		Namespace:                         "test-namespace",
		ExistingCustomSearchAttributeName: "OldName",
		NewCustomSearchAttributeName:      "NewName",
		ResourceVersion:                   "custom-version",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}
