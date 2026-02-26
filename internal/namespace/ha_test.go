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
	operationv1 "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/protobuf/proto"
)


// --- ListRegions ---

func TestListRegions_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace: "test-namespace",
		RegionStatus: map[string]*namespacev1.NamespaceRegionStatus{
			"aws-us-east-1": {State: namespacev1.NamespaceRegionStatus_STATE_ACTIVE},
			"aws-us-west-2": {State: namespacev1.NamespaceRegionStatus_STATE_PASSIVE},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	regions, err := client.ListRegions(ctx, "test-namespace")

	require.NoError(t, err)
	// Sorted by region ID
	assert.Equal(t, []namespace.RegionStatus{
		{Region: "aws-us-east-1", Status: namespacev1.NamespaceRegionStatus_STATE_ACTIVE},
		{Region: "aws-us-west-2", Status: namespacev1.NamespaceRegionStatus_STATE_PASSIVE},
	}, regions)
}

func TestListRegions_Empty(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace:    "test-namespace",
		RegionStatus: map[string]*namespacev1.NamespaceRegionStatus{},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	regions, err := client.ListRegions(ctx, "test-namespace")

	require.NoError(t, err)
	assert.Empty(t, regions)
}

func TestListRegions_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")
	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	regions, err := client.ListRegions(ctx, "test-namespace")

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, regions)
}

// --- UpdateHA ---

func TestUpdateHA_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			HighAvailability: &namespacev1.HighAvailabilitySpec{
				DisableManagedFailover: false,
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	expectedOp := &operationv1.AsyncOperation{Id: "test-op"}
	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:        "test-namespace",
				ResourceVersion:  "v1",
				AsyncOperationId: "test-op-id",
				Spec: &namespacev1.NamespaceSpec{
					HighAvailability: &namespacev1.HighAvailabilitySpec{
						DisableManagedFailover: true,
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	op, err := client.UpdateHA(ctx, namespace.UpdateHAParams{
		Namespace:           "test-namespace",
		DisableAutoFailover: true,
		AsyncOperationID:    "test-op-id",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, op)
}

func TestUpdateHA_CustomResourceVersion(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec:            &namespacev1.NamespaceSpec{},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	expectedOp := &operationv1.AsyncOperation{Id: "test-op"}
	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			return req.ResourceVersion == "custom-v2"
		})).
		Return(&cloudservice.UpdateNamespaceResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	op, err := client.UpdateHA(ctx, namespace.UpdateHAParams{
		Namespace:       "test-namespace",
		ResourceVersion: "custom-v2",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, op)
}

func TestUpdateHA_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")
	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	op, err := client.UpdateHA(ctx, namespace.UpdateHAParams{Namespace: "test-namespace"})

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, op)
}

// --- AddRegion ---

func TestAddRegion_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
	}
	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	expectedOp := &operationv1.AsyncOperation{Id: "add-op"}
	mockCloud.EXPECT().
		AddNamespaceRegion(ctx, mock.MatchedBy(func(req *cloudservice.AddNamespaceRegionRequest) bool {
			expected := &cloudservice.AddNamespaceRegionRequest{
				Namespace:        "test-namespace",
				Region:           "aws-us-west-2",
				ResourceVersion:  "v1",
				AsyncOperationId: "add-op-id",
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.AddNamespaceRegionResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	op, err := client.AddRegion(ctx, namespace.AddRegionParams{
		Namespace:        "test-namespace",
		Region:           "aws-us-west-2",
		AsyncOperationID: "add-op-id",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, op)
}

func TestAddRegion_CustomResourceVersion(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
	}
	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	expectedOp := &operationv1.AsyncOperation{Id: "add-op"}
	mockCloud.EXPECT().
		AddNamespaceRegion(ctx, mock.MatchedBy(func(req *cloudservice.AddNamespaceRegionRequest) bool {
			return req.ResourceVersion == "custom-v2"
		})).
		Return(&cloudservice.AddNamespaceRegionResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	op, err := client.AddRegion(ctx, namespace.AddRegionParams{
		Namespace:       "test-namespace",
		Region:          "aws-us-west-2",
		ResourceVersion: "custom-v2",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, op)
}

func TestAddRegion_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")
	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	op, err := client.AddRegion(ctx, namespace.AddRegionParams{
		Namespace: "test-namespace",
		Region:    "aws-us-west-2",
	})

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, op)
}

func TestAddRegion_Error(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
	}
	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	expectedErr := errors.New("add region failed")
	mockCloud.EXPECT().
		AddNamespaceRegion(ctx, mock.Anything).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	op, err := client.AddRegion(ctx, namespace.AddRegionParams{
		Namespace: "test-namespace",
		Region:    "aws-us-west-2",
	})

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, op)
}

// --- RemoveRegion ---

func TestRemoveRegion_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
	}
	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	expectedOp := &operationv1.AsyncOperation{Id: "remove-op"}
	mockCloud.EXPECT().
		DeleteNamespaceRegion(ctx, mock.MatchedBy(func(req *cloudservice.DeleteNamespaceRegionRequest) bool {
			expected := &cloudservice.DeleteNamespaceRegionRequest{
				Namespace:        "test-namespace",
				Region:           "aws-us-west-2",
				ResourceVersion:  "v1",
				AsyncOperationId: "remove-op-id",
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.DeleteNamespaceRegionResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	op, err := client.RemoveRegion(ctx, namespace.RemoveRegionParams{
		Namespace:        "test-namespace",
		Region:           "aws-us-west-2",
		AsyncOperationID: "remove-op-id",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, op)
}

func TestRemoveRegion_CustomResourceVersion(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
	}
	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	expectedOp := &operationv1.AsyncOperation{Id: "remove-op"}
	mockCloud.EXPECT().
		DeleteNamespaceRegion(ctx, mock.MatchedBy(func(req *cloudservice.DeleteNamespaceRegionRequest) bool {
			return req.ResourceVersion == "custom-v2"
		})).
		Return(&cloudservice.DeleteNamespaceRegionResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	op, err := client.RemoveRegion(ctx, namespace.RemoveRegionParams{
		Namespace:       "test-namespace",
		Region:          "aws-us-west-2",
		ResourceVersion: "custom-v2",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, op)
}

func TestRemoveRegion_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")
	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	op, err := client.RemoveRegion(ctx, namespace.RemoveRegionParams{
		Namespace: "test-namespace",
		Region:    "aws-us-west-2",
	})

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, op)
}

func TestRemoveRegion_Error(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
	}
	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	expectedErr := errors.New("delete region failed")
	mockCloud.EXPECT().
		DeleteNamespaceRegion(ctx, mock.Anything).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	op, err := client.RemoveRegion(ctx, namespace.RemoveRegionParams{
		Namespace: "test-namespace",
		Region:    "aws-us-west-2",
	})

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, op)
}

// --- Failover ---

func TestFailover_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedOp := &operationv1.AsyncOperation{Id: "failover-op"}
	mockCloud.EXPECT().
		FailoverNamespaceRegion(ctx, mock.MatchedBy(func(req *cloudservice.FailoverNamespaceRegionRequest) bool {
			expected := &cloudservice.FailoverNamespaceRegionRequest{
				Namespace:        "test-namespace",
				Region:           "aws-us-west-2",
				AsyncOperationId: "failover-op-id",
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.FailoverNamespaceRegionResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	op, err := client.Failover(ctx, namespace.FailoverParams{
		Namespace:        "test-namespace",
		Region:           "aws-us-west-2",
		AsyncOperationID: "failover-op-id",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, op)
}

func TestFailover_Error(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("failover failed")
	mockCloud.EXPECT().
		FailoverNamespaceRegion(ctx, mock.Anything).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	op, err := client.Failover(ctx, namespace.FailoverParams{
		Namespace: "test-namespace",
		Region:    "aws-us-west-2",
	})

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, op)
}
