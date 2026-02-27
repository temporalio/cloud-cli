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

func TestListTags_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace: "test-namespace",
		Tags: map[string]string{
			"environment": "production",
			"team":        "platform",
			"cost-center": "engineering",
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.ListTags(ctx, "test-namespace")

	require.NoError(t, err)
	require.Len(t, result, 3)
	// Verify tags are sorted by key
	assert.Equal(t, []namespace.Tag{
		{Key: "cost-center", Value: "engineering"},
		{Key: "environment", Value: "production"},
		{Key: "team", Value: "platform"},
	}, result)
}

func TestListTags_EmptyList(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace: "test-namespace",
		Tags:      nil,
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.ListTags(ctx, "test-namespace")

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListTags_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.ListTags(ctx, "test-namespace")

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, result)
}

func TestSetTag_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}

	mockCloud.EXPECT().
		UpdateNamespaceTags(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceTagsRequest) bool {
			expected := &cloudservice.UpdateNamespaceTagsRequest{
				Namespace:        "test-namespace",
				TagsToUpsert:     map[string]string{"environment": "production"},
				AsyncOperationId: "test-async-op",
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceTagsResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.SetTag(ctx, namespace.SetTagParams{
		Namespace:        "test-namespace",
		Key:              "environment",
		Value:            "production",
		AsyncOperationID: "test-async-op",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestSetTag_UpdateNamespaceTagsError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("update failed")

	mockCloud.EXPECT().
		UpdateNamespaceTags(ctx, mock.Anything).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.SetTag(ctx, namespace.SetTagParams{
		Namespace: "test-namespace",
		Key:       "environment",
		Value:     "production",
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestDeleteTags_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}

	mockCloud.EXPECT().
		UpdateNamespaceTags(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceTagsRequest) bool {
			expected := &cloudservice.UpdateNamespaceTagsRequest{
				Namespace:        "test-namespace",
				TagsToRemove:     []string{"environment"},
				AsyncOperationId: "test-async-op",
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceTagsResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteTags(ctx, namespace.DeleteTagsParams{
		Namespace:        "test-namespace",
		Keys:             []string{"environment"},
		AsyncOperationID: "test-async-op",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestDeleteTags_UpdateNamespaceTagsError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("update failed")

	mockCloud.EXPECT().
		UpdateNamespaceTags(ctx, mock.Anything).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteTags(ctx, namespace.DeleteTagsParams{
		Namespace: "test-namespace",
		Keys:      []string{"environment"},
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, expectedErr, err)
}

func TestDeleteTags_WithAsyncOperationID(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedOp := &operation.AsyncOperation{Id: "custom-async-op-id"}

	mockCloud.EXPECT().
		UpdateNamespaceTags(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceTagsRequest) bool {
			expected := &cloudservice.UpdateNamespaceTagsRequest{
				Namespace:        "test-namespace",
				TagsToRemove:     []string{"test-key"},
				AsyncOperationId: "custom-async-op-id",
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceTagsResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteTags(ctx, namespace.DeleteTagsParams{
		Namespace:        "test-namespace",
		Keys:             []string{"test-key"},
		AsyncOperationID: "custom-async-op-id",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}
