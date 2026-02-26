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
	operationv1 "go.temporal.io/cloud-sdk/api/operation/v1"
	sinkv1 "go.temporal.io/cloud-sdk/api/sink/v1"
)

func TestClient_GetExportSink_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	expectedSink := &namespacev1.ExportSink{
		Name:            "my-sink",
		ResourceVersion: "v1",
		Spec: &namespacev1.ExportSinkSpec{
			Name:    "my-sink",
			Enabled: true,
		},
	}

	mockCloud.EXPECT().
		GetNamespaceExportSink(ctx, &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: expectedSink}, nil)

	sink, err := client.GetExportSink(ctx, "test-ns", "my-sink")
	require.NoError(t, err)
	assert.Equal(t, expectedSink, sink)
}

func TestClient_GetExportSink_Error(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	mockCloud.EXPECT().
		GetNamespaceExportSink(ctx, &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Name:      "my-sink",
		}).
		Return(nil, errors.New("not found"))

	_, err := client.GetExportSink(ctx, "test-ns", "my-sink")
	require.Error(t, err)
}

func TestClient_ListExportSinks_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	sinks := []*namespacev1.ExportSink{
		{Name: "sink-1"},
		{Name: "sink-2"},
	}

	mockCloud.EXPECT().
		GetNamespaceExportSinks(ctx, &cloudservice.GetNamespaceExportSinksRequest{
			Namespace: "test-ns",
			PageToken: "",
		}).
		Return(&cloudservice.GetNamespaceExportSinksResponse{
			Sinks:         sinks,
			NextPageToken: "",
		}, nil)

	result, err := client.ListExportSinks(ctx, "test-ns")
	require.NoError(t, err)
	assert.Equal(t, sinks, result)
}

func TestClient_ListExportSinks_Pagination(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	page1 := []*namespacev1.ExportSink{{Name: "sink-1"}}
	page2 := []*namespacev1.ExportSink{{Name: "sink-2"}}

	mockCloud.EXPECT().
		GetNamespaceExportSinks(ctx, &cloudservice.GetNamespaceExportSinksRequest{
			Namespace: "test-ns",
			PageToken: "",
		}).
		Return(&cloudservice.GetNamespaceExportSinksResponse{
			Sinks:         page1,
			NextPageToken: "token-1",
		}, nil)

	mockCloud.EXPECT().
		GetNamespaceExportSinks(ctx, &cloudservice.GetNamespaceExportSinksRequest{
			Namespace: "test-ns",
			PageToken: "token-1",
		}).
		Return(&cloudservice.GetNamespaceExportSinksResponse{
			Sinks:         page2,
			NextPageToken: "",
		}, nil)

	result, err := client.ListExportSinks(ctx, "test-ns")
	require.NoError(t, err)
	assert.Equal(t, append(page1, page2...), result)
}

func TestClient_CreateS3ExportSink_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	expectedOp := &operationv1.AsyncOperation{Id: "op-1"}

	mockCloud.EXPECT().
		CreateNamespaceExportSink(ctx, &cloudservice.CreateNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Spec: &namespacev1.ExportSinkSpec{
				Name:    "my-sink",
				Enabled: true,
				S3: &sinkv1.S3Spec{
					RoleName:     "my-role",
					BucketName:   "my-bucket",
					Region:       "us-east-1",
					AwsAccountId: "123456789012",
					KmsArn:       "",
				},
			},
			AsyncOperationId: "",
		}).
		Return(&cloudservice.CreateNamespaceExportSinkResponse{AsyncOperation: expectedOp}, nil)

	op, err := client.CreateS3ExportSink(ctx, namespace.CreateS3ExportSinkParams{
		Namespace:    "test-ns",
		SinkName:     "my-sink",
		RoleName:     "my-role",
		BucketName:   "my-bucket",
		Region:       "us-east-1",
		AwsAccountID: "123456789012",
	})
	require.NoError(t, err)
	assert.Equal(t, expectedOp, op)
}

func TestClient_UpdateS3ExportSink_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	existingSink := &namespacev1.ExportSink{
		Name:            "my-sink",
		ResourceVersion: "v1",
		Spec: &namespacev1.ExportSinkSpec{
			Name:    "my-sink",
			Enabled: true,
			S3: &sinkv1.S3Spec{
				RoleName:     "old-role",
				BucketName:   "my-bucket",
				Region:       "us-east-1",
				AwsAccountId: "123456789012",
			},
		},
	}

	expectedOp := &operationv1.AsyncOperation{Id: "op-1"}

	mockCloud.EXPECT().
		GetNamespaceExportSink(ctx, &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: existingSink}, nil)

	mockCloud.EXPECT().
		UpdateNamespaceExportSink(ctx, &cloudservice.UpdateNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Spec: &namespacev1.ExportSinkSpec{
				Name:    "my-sink",
				Enabled: true,
				S3: &sinkv1.S3Spec{
					RoleName:     "new-role",
					BucketName:   "my-bucket",
					Region:       "us-east-1",
					AwsAccountId: "123456789012",
					KmsArn:       "",
				},
			},
			ResourceVersion:  "v1",
			AsyncOperationId: "",
		}).
		Return(&cloudservice.UpdateNamespaceExportSinkResponse{AsyncOperation: expectedOp}, nil)

	op, err := client.UpdateS3ExportSink(ctx, namespace.UpdateS3ExportSinkParams{
		Namespace:    "test-ns",
		SinkName:     "my-sink",
		RoleName:     "new-role",
		BucketName:   "my-bucket",
		Region:       "us-east-1",
		AwsAccountID: "123456789012",
	})
	require.NoError(t, err)
	assert.Equal(t, expectedOp, op)
}

func TestClient_UpdateS3ExportSink_GetSinkError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	mockCloud.EXPECT().
		GetNamespaceExportSink(ctx, &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Name:      "my-sink",
		}).
		Return(nil, errors.New("not found"))

	_, err := client.UpdateS3ExportSink(ctx, namespace.UpdateS3ExportSinkParams{
		Namespace:    "test-ns",
		SinkName:     "my-sink",
		RoleName:     "new-role",
		BucketName:   "my-bucket",
		Region:       "us-east-1",
		AwsAccountID: "123456789012",
	})
	require.Error(t, err)
}

func TestClient_UpdateS3ExportSink_CustomResourceVersion(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	existingSink := &namespacev1.ExportSink{
		Name:            "my-sink",
		ResourceVersion: "v1",
		Spec: &namespacev1.ExportSinkSpec{
			Name:    "my-sink",
			Enabled: false,
			S3:      &sinkv1.S3Spec{},
		},
	}

	expectedOp := &operationv1.AsyncOperation{Id: "op-1"}

	mockCloud.EXPECT().
		GetNamespaceExportSink(ctx, &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: existingSink}, nil)

	mockCloud.EXPECT().
		UpdateNamespaceExportSink(ctx, &cloudservice.UpdateNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Spec: &namespacev1.ExportSinkSpec{
				Name:    "my-sink",
				Enabled: false,
				S3: &sinkv1.S3Spec{
					RoleName:     "new-role",
					BucketName:   "my-bucket",
					Region:       "us-east-1",
					AwsAccountId: "123456789012",
				},
			},
			// custom resource version should override the fetched one
			ResourceVersion:  "v2",
			AsyncOperationId: "",
		}).
		Return(&cloudservice.UpdateNamespaceExportSinkResponse{AsyncOperation: expectedOp}, nil)

	op, err := client.UpdateS3ExportSink(ctx, namespace.UpdateS3ExportSinkParams{
		Namespace:       "test-ns",
		SinkName:        "my-sink",
		RoleName:        "new-role",
		BucketName:      "my-bucket",
		Region:          "us-east-1",
		AwsAccountID:    "123456789012",
		ResourceVersion: "v2",
	})
	require.NoError(t, err)
	assert.Equal(t, expectedOp, op)
}

func TestClient_ValidateS3ExportSink_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	mockCloud.EXPECT().
		ValidateNamespaceExportSink(ctx, &cloudservice.ValidateNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Spec: &namespacev1.ExportSinkSpec{
				Name:    "my-sink",
				Enabled: true,
				S3: &sinkv1.S3Spec{
					RoleName:     "my-role",
					BucketName:   "my-bucket",
					Region:       "us-east-1",
					AwsAccountId: "123456789012",
				},
			},
		}).
		Return(&cloudservice.ValidateNamespaceExportSinkResponse{}, nil)

	err := client.ValidateS3ExportSink(ctx, namespace.ValidateS3ExportSinkParams{
		Namespace:    "test-ns",
		SinkName:     "my-sink",
		RoleName:     "my-role",
		BucketName:   "my-bucket",
		Region:       "us-east-1",
		AwsAccountID: "123456789012",
	})
	require.NoError(t, err)
}

func TestClient_CreateGCSExportSink_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	expectedOp := &operationv1.AsyncOperation{Id: "op-1"}

	mockCloud.EXPECT().
		CreateNamespaceExportSink(ctx, &cloudservice.CreateNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Spec: &namespacev1.ExportSinkSpec{
				Name:    "my-sink",
				Enabled: true,
				Gcs: &sinkv1.GCSSpec{
					SaId:         "sa@project.iam.gserviceaccount.com",
					BucketName:   "my-bucket",
					GcpProjectId: "my-project",
					Region:       "us-central1",
				},
			},
			AsyncOperationId: "",
		}).
		Return(&cloudservice.CreateNamespaceExportSinkResponse{AsyncOperation: expectedOp}, nil)

	op, err := client.CreateGCSExportSink(ctx, namespace.CreateGCSExportSinkParams{
		Namespace:    "test-ns",
		SinkName:     "my-sink",
		SaID:         "sa@project.iam.gserviceaccount.com",
		BucketName:   "my-bucket",
		GcpProjectID: "my-project",
		Region:       "us-central1",
	})
	require.NoError(t, err)
	assert.Equal(t, expectedOp, op)
}

func TestClient_UpdateGCSExportSink_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	existingSink := &namespacev1.ExportSink{
		Name:            "my-sink",
		ResourceVersion: "v1",
		Spec: &namespacev1.ExportSinkSpec{
			Name:    "my-sink",
			Enabled: true,
			Gcs: &sinkv1.GCSSpec{
				SaId:         "old-sa@project.iam.gserviceaccount.com",
				BucketName:   "my-bucket",
				GcpProjectId: "my-project",
				Region:       "us-central1",
			},
		},
	}

	expectedOp := &operationv1.AsyncOperation{Id: "op-1"}

	mockCloud.EXPECT().
		GetNamespaceExportSink(ctx, &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: existingSink}, nil)

	mockCloud.EXPECT().
		UpdateNamespaceExportSink(ctx, &cloudservice.UpdateNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Spec: &namespacev1.ExportSinkSpec{
				Name:    "my-sink",
				Enabled: true,
				Gcs: &sinkv1.GCSSpec{
					SaId:         "new-sa@project.iam.gserviceaccount.com",
					BucketName:   "my-bucket",
					GcpProjectId: "my-project",
					Region:       "us-central1",
				},
			},
			ResourceVersion:  "v1",
			AsyncOperationId: "",
		}).
		Return(&cloudservice.UpdateNamespaceExportSinkResponse{AsyncOperation: expectedOp}, nil)

	op, err := client.UpdateGCSExportSink(ctx, namespace.UpdateGCSExportSinkParams{
		Namespace:    "test-ns",
		SinkName:     "my-sink",
		SaID:         "new-sa@project.iam.gserviceaccount.com",
		BucketName:   "my-bucket",
		GcpProjectID: "my-project",
		Region:       "us-central1",
	})
	require.NoError(t, err)
	assert.Equal(t, expectedOp, op)
}

func TestClient_UpdateGCSExportSink_GetSinkError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	mockCloud.EXPECT().
		GetNamespaceExportSink(ctx, &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Name:      "my-sink",
		}).
		Return(nil, errors.New("not found"))

	_, err := client.UpdateGCSExportSink(ctx, namespace.UpdateGCSExportSinkParams{
		Namespace:    "test-ns",
		SinkName:     "my-sink",
		SaID:         "sa@project.iam.gserviceaccount.com",
		BucketName:   "my-bucket",
		GcpProjectID: "my-project",
		Region:       "us-central1",
	})
	require.Error(t, err)
}

func TestClient_ValidateGCSExportSink_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	mockCloud.EXPECT().
		ValidateNamespaceExportSink(ctx, &cloudservice.ValidateNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Spec: &namespacev1.ExportSinkSpec{
				Name:    "my-sink",
				Enabled: true,
				Gcs: &sinkv1.GCSSpec{
					SaId:         "sa@project.iam.gserviceaccount.com",
					BucketName:   "my-bucket",
					GcpProjectId: "my-project",
					Region:       "us-central1",
				},
			},
		}).
		Return(&cloudservice.ValidateNamespaceExportSinkResponse{}, nil)

	err := client.ValidateGCSExportSink(ctx, namespace.ValidateGCSExportSinkParams{
		Namespace:    "test-ns",
		SinkName:     "my-sink",
		SaID:         "sa@project.iam.gserviceaccount.com",
		BucketName:   "my-bucket",
		GcpProjectID: "my-project",
		Region:       "us-central1",
	})
	require.NoError(t, err)
}

func TestClient_EnableExportSink_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	existingSink := &namespacev1.ExportSink{
		Name:            "my-sink",
		ResourceVersion: "v1",
		Spec: &namespacev1.ExportSinkSpec{
			Name:    "my-sink",
			Enabled: false,
			S3:      &sinkv1.S3Spec{RoleName: "my-role"},
		},
	}

	expectedOp := &operationv1.AsyncOperation{Id: "op-1"}

	mockCloud.EXPECT().
		GetNamespaceExportSink(ctx, &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: existingSink}, nil)

	mockCloud.EXPECT().
		UpdateNamespaceExportSink(ctx, &cloudservice.UpdateNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Spec: &namespacev1.ExportSinkSpec{
				Name:    "my-sink",
				Enabled: true,
				S3:      &sinkv1.S3Spec{RoleName: "my-role"},
			},
			ResourceVersion:  "v1",
			AsyncOperationId: "",
		}).
		Return(&cloudservice.UpdateNamespaceExportSinkResponse{AsyncOperation: expectedOp}, nil)

	op, err := client.EnableExportSink(ctx, namespace.EnableExportSinkParams{
		Namespace: "test-ns",
		SinkName:  "my-sink",
	})
	require.NoError(t, err)
	assert.Equal(t, expectedOp, op)
}

func TestClient_EnableExportSink_GetSinkError(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	mockCloud.EXPECT().
		GetNamespaceExportSink(ctx, &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Name:      "my-sink",
		}).
		Return(nil, errors.New("not found"))

	_, err := client.EnableExportSink(ctx, namespace.EnableExportSinkParams{
		Namespace: "test-ns",
		SinkName:  "my-sink",
	})
	require.Error(t, err)
}

func TestClient_DisableExportSink_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	existingSink := &namespacev1.ExportSink{
		Name:            "my-sink",
		ResourceVersion: "v1",
		Spec: &namespacev1.ExportSinkSpec{
			Name:    "my-sink",
			Enabled: true,
			S3:      &sinkv1.S3Spec{RoleName: "my-role"},
		},
	}

	expectedOp := &operationv1.AsyncOperation{Id: "op-1"}

	mockCloud.EXPECT().
		GetNamespaceExportSink(ctx, &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: existingSink}, nil)

	mockCloud.EXPECT().
		UpdateNamespaceExportSink(ctx, &cloudservice.UpdateNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Spec: &namespacev1.ExportSinkSpec{
				Name:    "my-sink",
				Enabled: false,
				S3:      &sinkv1.S3Spec{RoleName: "my-role"},
			},
			ResourceVersion:  "v1",
			AsyncOperationId: "",
		}).
		Return(&cloudservice.UpdateNamespaceExportSinkResponse{AsyncOperation: expectedOp}, nil)

	op, err := client.DisableExportSink(ctx, namespace.DisableExportSinkParams{
		Namespace: "test-ns",
		SinkName:  "my-sink",
	})
	require.NoError(t, err)
	assert.Equal(t, expectedOp, op)
}

func TestClient_DeleteExportSink_Success(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	expectedOp := &operationv1.AsyncOperation{Id: "op-1"}

	mockCloud.EXPECT().
		DeleteNamespaceExportSink(ctx, &cloudservice.DeleteNamespaceExportSinkRequest{
			Namespace:        "test-ns",
			Name:             "my-sink",
			ResourceVersion:  "v1",
			AsyncOperationId: "my-op-id",
		}).
		Return(&cloudservice.DeleteNamespaceExportSinkResponse{AsyncOperation: expectedOp}, nil)

	op, err := client.DeleteExportSink(ctx, namespace.DeleteExportSinkParams{
		Namespace:        "test-ns",
		SinkName:         "my-sink",
		ResourceVersion:  "v1",
		AsyncOperationID: "my-op-id",
	})
	require.NoError(t, err)
	assert.Equal(t, expectedOp, op)
}

func TestClient_DeleteExportSink_Error(t *testing.T) {
	ctx := context.Background()
	mockCloud := nsmock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: mockCloud}

	mockCloud.EXPECT().
		DeleteNamespaceExportSink(ctx, &cloudservice.DeleteNamespaceExportSinkRequest{
			Namespace: "test-ns",
			Name:      "my-sink",
		}).
		Return(nil, errors.New("delete failed"))

	_, err := client.DeleteExportSink(ctx, namespace.DeleteExportSinkParams{
		Namespace: "test-ns",
		SinkName:  "my-sink",
	})
	require.Error(t, err)
}
