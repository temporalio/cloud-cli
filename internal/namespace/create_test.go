package namespace_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cloud-cli/internal/namespace"
	nsmock "github.com/temporalio/cloud-cli/internal/namespace/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
)

func TestClient_CreateNamespace_Success(t *testing.T) {
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	mockCloud.EXPECT().CreateNamespace(context.Background(), &cloudservice.CreateNamespaceRequest{
		AsyncOperationId: "op-123",
		Spec:             &namespacev1.NamespaceSpec{Name: "my-ns"},
	}).Return(
		&cloudservice.CreateNamespaceResponse{
			Namespace:      "my-ns.account123",
			AsyncOperation: &operation.AsyncOperation{Id: "op-123"},
		}, nil,
	)

	result, err := client.CreateNamespace(context.Background(), namespace.CreateNamespaceParams{
		Spec:             &namespacev1.NamespaceSpec{Name: "my-ns"},
		AsyncOperationID: "op-123",
	})
	require.NoError(t, err)
	assert.Equal(t, "my-ns.account123", result.NamespaceID)
	assert.Equal(t, "op-123", result.AsyncOp.Id)
}

func TestClient_CreateNamespace_Error(t *testing.T) {
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	mockCloud.EXPECT().CreateNamespace(context.Background(), &cloudservice.CreateNamespaceRequest{}).
		Return(nil, errors.New("create failed"))

	_, err := client.CreateNamespace(context.Background(), namespace.CreateNamespaceParams{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create failed")
}
