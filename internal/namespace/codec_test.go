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

func TestGetCodecServer_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	ns := &namespacev1.Namespace{
		Namespace: "test-namespace",
		Spec: &namespacev1.NamespaceSpec{
			CodecServer: &namespacev1.CodecServerSpec{
				Endpoint:        "https://codec.example.com",
				PassAccessToken: true,
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: ns}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.GetCodecServer(ctx, "test-namespace")

	require.NoError(t, err)
	assert.Equal(t, &namespacev1.CodecServerSpec{
		Endpoint:        "https://codec.example.com",
		PassAccessToken: true,
	}, result)
}

func TestGetCodecServer_NilCodecServer(t *testing.T) {
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

	result, err := client.GetCodecServer(ctx, "test-namespace")

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestGetCodecServer_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.GetCodecServer(ctx, "test-namespace")

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, result)
}

func TestSetCodec_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			Name:          "test-namespace",
			RetentionDays: 7,
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:        "test-namespace",
				ResourceVersion:  "v1",
				AsyncOperationId: "test-async-op",
				Spec: &namespacev1.NamespaceSpec{
					Name:          "test-namespace",
					RetentionDays: 7,
					CodecServer: &namespacev1.CodecServerSpec{
						Endpoint:        "https://codec.example.com",
						PassAccessToken: true,
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.SetCodec(ctx, namespace.SetCodecParams{
		Namespace:        "test-namespace",
		Endpoint:         "https://codec.example.com",
		PassAccessToken:  true,
		AsyncOperationID: "test-async-op",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestSetCodec_WithCustomErrorMessage(t *testing.T) {
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

	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:       "test-namespace",
				ResourceVersion: "v1",
				Spec: &namespacev1.NamespaceSpec{
					CodecServer: &namespacev1.CodecServerSpec{
						Endpoint: "https://codec.example.com",
						CustomErrorMessage: &namespacev1.CodecServerSpec_CustomErrorMessage{
							Default: &namespacev1.CodecServerSpec_CustomErrorMessage_ErrorMessage{
								Message: "Codec unavailable",
								Link:    "https://docs.example.com",
							},
						},
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.SetCodec(ctx, namespace.SetCodecParams{
		Namespace:                        "test-namespace",
		Endpoint:                         "https://codec.example.com",
		CustomErrorMessageDefaultMessage: "Codec unavailable",
		CustomErrorMessageDefaultLink:    "https://docs.example.com",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestSetCodec_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.SetCodec(ctx, namespace.SetCodecParams{
		Namespace: "test-namespace",
		Endpoint:  "https://codec.example.com",
	})

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, result)
}

func TestSetCodec_CustomResourceVersion(t *testing.T) {
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

	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:       "test-namespace",
				ResourceVersion: "custom-version",
				Spec: &namespacev1.NamespaceSpec{
					CodecServer: &namespacev1.CodecServerSpec{
						Endpoint: "https://codec.example.com",
					},
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.SetCodec(ctx, namespace.SetCodecParams{
		Namespace:       "test-namespace",
		Endpoint:        "https://codec.example.com",
		ResourceVersion: "custom-version",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestDeleteCodec_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			Name:          "test-namespace",
			RetentionDays: 7,
			CodecServer: &namespacev1.CodecServerSpec{
				Endpoint: "https://codec.example.com",
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:        "test-namespace",
				ResourceVersion:  "v1",
				AsyncOperationId: "test-async-op",
				Spec: &namespacev1.NamespaceSpec{
					Name:          "test-namespace",
					RetentionDays: 7,
					CodecServer:   nil,
				},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteCodec(ctx, namespace.DeleteCodecParams{
		Namespace:        "test-namespace",
		AsyncOperationID: "test-async-op",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}

func TestDeleteCodec_GetNamespaceError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	expectedErr := errors.New("namespace not found")

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(nil, expectedErr)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteCodec(ctx, namespace.DeleteCodecParams{
		Namespace: "test-namespace",
	})

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, result)
}

func TestDeleteCodec_CustomResourceVersion(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)

	existingNamespace := &namespacev1.Namespace{
		Namespace:       "test-namespace",
		ResourceVersion: "v1",
		Spec: &namespacev1.NamespaceSpec{
			CodecServer: &namespacev1.CodecServerSpec{
				Endpoint: "https://codec.example.com",
			},
		},
	}

	mockCloud.EXPECT().
		GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: "test-namespace"}).
		Return(&cloudservice.GetNamespaceResponse{Namespace: existingNamespace}, nil)

	expectedOp := &operation.AsyncOperation{Id: "test-operation-id"}

	mockCloud.EXPECT().
		UpdateNamespace(ctx, mock.MatchedBy(func(req *cloudservice.UpdateNamespaceRequest) bool {
			expected := &cloudservice.UpdateNamespaceRequest{
				Namespace:       "test-namespace",
				ResourceVersion: "custom-version",
				Spec:            &namespacev1.NamespaceSpec{},
			}
			return proto.Equal(expected, req)
		})).
		Return(&cloudservice.UpdateNamespaceResponse{AsyncOperation: expectedOp}, nil)

	client := &namespace.Client{Cloud: mockCloud}

	result, err := client.DeleteCodec(ctx, namespace.DeleteCodecParams{
		Namespace:       "test-namespace",
		ResourceVersion: "custom-version",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedOp, result)
}
