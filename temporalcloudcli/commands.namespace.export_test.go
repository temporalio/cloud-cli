package temporalcloudcli_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	sinkv1 "go.temporal.io/cloud-sdk/api/sink/v1"
)

// testS3Spec returns a sample S3 spec for use in tests.
func testS3Spec() *sinkv1.S3Spec {
	return &sinkv1.S3Spec{
		RoleName:     "my-role",
		BucketName:   "my-bucket",
		Region:       "us-east-1",
		AwsAccountId: "123456789012",
	}
}

// testGCSSpec returns a sample GCS spec for use in tests.
func testGCSSpec() *sinkv1.GCSSpec {
	return &sinkv1.GCSSpec{
		SaId:         "my-sa@project.iam.gserviceaccount.com",
		BucketName:   "my-bucket",
		GcpProjectId: "my-project",
		Region:       "us-central1",
	}
}

// testExportSink returns a sample ExportSink with an S3 spec for use in tests.
func testExportSink(enabled bool) *namespacev1.ExportSink {
	return &namespacev1.ExportSink{
		ResourceVersion: "rv-1",
		Spec: &namespacev1.ExportSinkSpec{
			Name:    "my-sink",
			Enabled: enabled,
			S3:      testS3Spec(),
		},
	}
}

// --- GetExportSink ---

func TestGetExportSink_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	sink := testExportSink(true)
	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.GetExportSink(context.Background(), temporalcloudcli.GetExportSinkParams{
		Namespace: "my-namespace",
		SinkName:  "my-sink",
		Cloud:     mockCloud,
		Printer:   newTestPrinter(&buf),
	})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "my-sink")
}

func TestGetExportSink_Error(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(nil, apiErr)

	var buf bytes.Buffer
	err := temporalcloudcli.GetExportSink(context.Background(), temporalcloudcli.GetExportSinkParams{
		Namespace: "my-namespace",
		SinkName:  "my-sink",
		Cloud:     mockCloud,
		Printer:   newTestPrinter(&buf),
	})
	require.ErrorIs(t, err, apiErr)
}

// --- ListExportSinks ---

func TestListExportSinks_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	sinks := []*namespacev1.ExportSink{testExportSink(true)}
	mockCloud.EXPECT().
		GetNamespaceExportSinks(context.Background(), &cloudservice.GetNamespaceExportSinksRequest{
			Namespace: "my-namespace",
			PageToken: "",
		}).
		Return(&cloudservice.GetNamespaceExportSinksResponse{Sinks: sinks}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.ListExportSinks(context.Background(), temporalcloudcli.ListExportSinksParams{
		Namespace: "my-namespace",
		Cloud:     mockCloud,
		Printer:   &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)
}

func TestListExportSinks_Error(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetNamespaceExportSinks(context.Background(), &cloudservice.GetNamespaceExportSinksRequest{
			Namespace: "my-namespace",
			PageToken: "",
		}).
		Return(nil, apiErr)

	var buf bytes.Buffer
	err := temporalcloudcli.ListExportSinks(context.Background(), temporalcloudcli.ListExportSinksParams{
		Namespace: "my-namespace",
		Cloud:     mockCloud,
		Printer:   newTestPrinter(&buf),
	})
	require.ErrorIs(t, err, apiErr)
}

// --- DeleteExportSink ---

func TestDeleteExportSink_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	sink := testExportSink(true)
	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)

	mockPrompter.EXPECT().
		PromptApply(sink.Spec, &namespacev1.ExportSinkSpec{}, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-123"}
	mockCloud.EXPECT().
		DeleteNamespaceExportSink(context.Background(), &cloudservice.DeleteNamespaceExportSinkRequest{
			Namespace:       "my-namespace",
			Name:            "my-sink",
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.DeleteNamespaceExportSinkResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().
		HandleOperation(op, "my-sink").
		Return(nil)

	err := temporalcloudcli.DeleteExportSink(context.Background(), temporalcloudcli.DeleteExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestDeleteExportSink_GetSinkError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("sink not found")

	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(nil, apiErr)

	err := temporalcloudcli.DeleteExportSink(context.Background(), temporalcloudcli.DeleteExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

func TestDeleteExportSink_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	sink := testExportSink(true)
	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)

	mockPrompter.EXPECT().
		PromptApply(sink.Spec, &namespacev1.ExportSinkSpec{}, false).
		Return(promptErr)

	err := temporalcloudcli.DeleteExportSink(context.Background(), temporalcloudcli.DeleteExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, promptErr)
}

func TestDeleteExportSink_ExplicitResourceVersion(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	sink := testExportSink(true)
	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)

	mockPrompter.EXPECT().
		PromptApply(sink.Spec, &namespacev1.ExportSinkSpec{}, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-456"}
	mockCloud.EXPECT().
		DeleteNamespaceExportSink(context.Background(), &cloudservice.DeleteNamespaceExportSinkRequest{
			Namespace:       "my-namespace",
			Name:            "my-sink",
			ResourceVersion: "rv-explicit",
		}).
		Return(&cloudservice.DeleteNamespaceExportSinkResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().
		HandleOperation(op, "my-sink").
		Return(nil)

	err := temporalcloudcli.DeleteExportSink(context.Background(), temporalcloudcli.DeleteExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		ResourceVersion:  "rv-explicit",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// --- EnableExportSink ---

func TestEnableExportSink_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	sink := testExportSink(false)
	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)

	oldSpec := sink.Spec
	newSpec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: true,
		S3:      testS3Spec(),
	}
	mockPrompter.EXPECT().
		PromptApply(oldSpec, newSpec, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-123"}
	mockCloud.EXPECT().
		UpdateNamespaceExportSink(context.Background(), &cloudservice.UpdateNamespaceExportSinkRequest{
			Namespace:       "my-namespace",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateNamespaceExportSinkResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().
		HandleOperation(op, "my-sink").
		Return(nil)

	err := temporalcloudcli.EnableExportSink(context.Background(), temporalcloudcli.EnableExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestEnableExportSink_GetSinkError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(nil, apiErr)

	err := temporalcloudcli.EnableExportSink(context.Background(), temporalcloudcli.EnableExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

func TestEnableExportSink_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	sink := testExportSink(false)
	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)

	mockPrompter.EXPECT().
		PromptApply(sink.Spec, &namespacev1.ExportSinkSpec{Name: "my-sink", Enabled: true, S3: testS3Spec()}, false).
		Return(promptErr)

	err := temporalcloudcli.EnableExportSink(context.Background(), temporalcloudcli.EnableExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, promptErr)
}

func TestEnableExportSink_UpdateError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	updateErr := errors.New("update error")

	sink := testExportSink(false)
	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)

	newSpec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: true,
		S3:      testS3Spec(),
	}
	mockPrompter.EXPECT().
		PromptApply(sink.Spec, newSpec, false).
		Return(nil)

	mockCloud.EXPECT().
		UpdateNamespaceExportSink(context.Background(), &cloudservice.UpdateNamespaceExportSinkRequest{
			Namespace:       "my-namespace",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(nil, updateErr)

	mockHandler.EXPECT().
		HandleUpdateErr(updateErr).
		Return(updateErr)

	err := temporalcloudcli.EnableExportSink(context.Background(), temporalcloudcli.EnableExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, updateErr)
}

// --- DisableExportSink ---

func TestDisableExportSink_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	sink := testExportSink(true)
	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)

	oldSpec := sink.Spec
	newSpec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: false,
		S3:      testS3Spec(),
	}
	mockPrompter.EXPECT().
		PromptApply(oldSpec, newSpec, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-123"}
	mockCloud.EXPECT().
		UpdateNamespaceExportSink(context.Background(), &cloudservice.UpdateNamespaceExportSinkRequest{
			Namespace:       "my-namespace",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateNamespaceExportSinkResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().
		HandleOperation(op, "my-sink").
		Return(nil)

	err := temporalcloudcli.DisableExportSink(context.Background(), temporalcloudcli.DisableExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestDisableExportSink_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	sink := testExportSink(true)
	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)

	mockPrompter.EXPECT().
		PromptApply(sink.Spec, &namespacev1.ExportSinkSpec{Name: "my-sink", Enabled: false, S3: testS3Spec()}, false).
		Return(promptErr)

	err := temporalcloudcli.DisableExportSink(context.Background(), temporalcloudcli.DisableExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, promptErr)
}

// --- CreateS3ExportSink ---

func TestCreateS3ExportSink_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	expectedSpec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: true,
		S3:      testS3Spec(),
	}
	mockPrompter.EXPECT().
		PromptApply(&namespacev1.ExportSinkSpec{}, expectedSpec, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-123"}
	mockCloud.EXPECT().
		CreateNamespaceExportSink(context.Background(), &cloudservice.CreateNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Spec:      expectedSpec,
		}).
		Return(&cloudservice.CreateNamespaceExportSinkResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().
		HandleOperation(op, "my-sink").
		Return(nil)

	err := temporalcloudcli.CreateS3ExportSink(context.Background(), temporalcloudcli.CreateS3ExportSinkParams{
		Namespace:    "my-namespace",
		SinkName:     "my-sink",
		RoleName:     "my-role",
		BucketName:   "my-bucket",
		Region:       "us-east-1",
		AwsAccountID: "123456789012",
		Cloud:        mockCloud,
		Prompter:     mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestCreateS3ExportSink_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	expectedSpec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: true,
		S3:      testS3Spec(),
	}
	mockPrompter.EXPECT().
		PromptApply(&namespacev1.ExportSinkSpec{}, expectedSpec, false).
		Return(promptErr)

	err := temporalcloudcli.CreateS3ExportSink(context.Background(), temporalcloudcli.CreateS3ExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		RoleName:         "my-role",
		BucketName:       "my-bucket",
		Region:           "us-east-1",
		AwsAccountID:     "123456789012",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, promptErr)
}

func TestCreateS3ExportSink_CreateError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	createErr := errors.New("create error")

	expectedSpec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: true,
		S3:      testS3Spec(),
	}
	mockPrompter.EXPECT().
		PromptApply(&namespacev1.ExportSinkSpec{}, expectedSpec, false).
		Return(nil)

	mockCloud.EXPECT().
		CreateNamespaceExportSink(context.Background(), &cloudservice.CreateNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Spec:      expectedSpec,
		}).
		Return(nil, createErr)

	mockHandler.EXPECT().
		HandleCreateErr(createErr).
		Return(createErr)

	err := temporalcloudcli.CreateS3ExportSink(context.Background(), temporalcloudcli.CreateS3ExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		RoleName:         "my-role",
		BucketName:       "my-bucket",
		Region:           "us-east-1",
		AwsAccountID:     "123456789012",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, createErr)
}

// --- UpdateS3ExportSink ---

func TestUpdateS3ExportSink_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	sink := testExportSink(true)
	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)

	newSpec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: true,
		S3: &sinkv1.S3Spec{
			RoleName:     "new-role",
			BucketName:   "my-bucket",
			Region:       "us-east-1",
			AwsAccountId: "123456789012",
		},
	}
	op := &operation.AsyncOperation{Id: "op-123"}
	mockCloud.EXPECT().
		UpdateNamespaceExportSink(context.Background(), &cloudservice.UpdateNamespaceExportSinkRequest{
			Namespace:       "my-namespace",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateNamespaceExportSinkResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().
		HandleOperation(op, "my-sink").
		Return(nil)

	err := temporalcloudcli.UpdateS3ExportSink(context.Background(), temporalcloudcli.UpdateS3ExportSinkParams{
		Namespace:    "my-namespace",
		SinkName:     "my-sink",
		RoleName:     "new-role",
		BucketName:   "my-bucket",
		Region:       "us-east-1",
		AwsAccountID: "123456789012",
		Cloud:        mockCloud,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestUpdateS3ExportSink_GetSinkError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(nil, apiErr)

	err := temporalcloudcli.UpdateS3ExportSink(context.Background(), temporalcloudcli.UpdateS3ExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		RoleName:         "my-role",
		BucketName:       "my-bucket",
		Region:           "us-east-1",
		AwsAccountID:     "123456789012",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

func TestUpdateS3ExportSink_PreservesEnabledState(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	// Sink is currently disabled — update should preserve that.
	sink := testExportSink(false)
	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)

	newSpec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: false, // preserved from existing sink
		S3:      testS3Spec(),
	}
	op := &operation.AsyncOperation{Id: "op-123"}
	mockCloud.EXPECT().
		UpdateNamespaceExportSink(context.Background(), &cloudservice.UpdateNamespaceExportSinkRequest{
			Namespace:       "my-namespace",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateNamespaceExportSinkResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().
		HandleOperation(op, "my-sink").
		Return(nil)

	err := temporalcloudcli.UpdateS3ExportSink(context.Background(), temporalcloudcli.UpdateS3ExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		RoleName:         "my-role",
		BucketName:       "my-bucket",
		Region:           "us-east-1",
		AwsAccountID:     "123456789012",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

// --- ValidateS3ExportSink ---

func TestValidateS3ExportSink_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	mockCloud.EXPECT().
		ValidateNamespaceExportSink(context.Background(), &cloudservice.ValidateNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Spec: &namespacev1.ExportSinkSpec{
				Name: "my-sink",
				S3:   testS3Spec(),
			},
		}).
		Return(&cloudservice.ValidateNamespaceExportSinkResponse{}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.ValidateS3ExportSink(context.Background(), temporalcloudcli.ValidateS3ExportSinkParams{
		Namespace:    "my-namespace",
		SinkName:     "my-sink",
		RoleName:     "my-role",
		BucketName:   "my-bucket",
		Region:       "us-east-1",
		AwsAccountID: "123456789012",
		Cloud:        mockCloud,
		Printer:      newTestPrinter(&buf),
	})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "my-sink")
}

func TestValidateS3ExportSink_Error(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	apiErr := errors.New("invalid config")

	mockCloud.EXPECT().
		ValidateNamespaceExportSink(context.Background(), &cloudservice.ValidateNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Spec: &namespacev1.ExportSinkSpec{
				Name: "my-sink",
				S3:   testS3Spec(),
			},
		}).
		Return(nil, apiErr)

	var buf bytes.Buffer
	err := temporalcloudcli.ValidateS3ExportSink(context.Background(), temporalcloudcli.ValidateS3ExportSinkParams{
		Namespace:    "my-namespace",
		SinkName:     "my-sink",
		RoleName:     "my-role",
		BucketName:   "my-bucket",
		Region:       "us-east-1",
		AwsAccountID: "123456789012",
		Cloud:        mockCloud,
		Printer:      newTestPrinter(&buf),
	})
	require.ErrorIs(t, err, apiErr)
}

// --- CreateGCSExportSink ---

func TestCreateGCSExportSink_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	expectedSpec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: true,
		Gcs:     testGCSSpec(),
	}
	mockPrompter.EXPECT().
		PromptApply(&namespacev1.ExportSinkSpec{}, expectedSpec, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-123"}
	mockCloud.EXPECT().
		CreateNamespaceExportSink(context.Background(), &cloudservice.CreateNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Spec:      expectedSpec,
		}).
		Return(&cloudservice.CreateNamespaceExportSinkResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().
		HandleOperation(op, "my-sink").
		Return(nil)

	err := temporalcloudcli.CreateGCSExportSink(context.Background(), temporalcloudcli.CreateGCSExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		SaID:             "my-sa@project.iam.gserviceaccount.com",
		BucketName:       "my-bucket",
		GcpProjectID:     "my-project",
		Region:           "us-central1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestCreateGCSExportSink_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	expectedSpec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: true,
		Gcs:     testGCSSpec(),
	}
	mockPrompter.EXPECT().
		PromptApply(&namespacev1.ExportSinkSpec{}, expectedSpec, false).
		Return(promptErr)

	err := temporalcloudcli.CreateGCSExportSink(context.Background(), temporalcloudcli.CreateGCSExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		SaID:             "my-sa@project.iam.gserviceaccount.com",
		BucketName:       "my-bucket",
		GcpProjectID:     "my-project",
		Region:           "us-central1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, promptErr)
}

// --- UpdateGCSExportSink ---

func TestUpdateGCSExportSink_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)

	existingSink := &namespacev1.ExportSink{
		ResourceVersion: "rv-1",
		Spec: &namespacev1.ExportSinkSpec{
			Name:    "my-sink",
			Enabled: true,
			Gcs:     testGCSSpec(),
		},
	}
	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: existingSink}, nil)

	newSpec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: true,
		Gcs: &sinkv1.GCSSpec{
			SaId:         "new-sa@project.iam.gserviceaccount.com",
			BucketName:   "my-bucket",
			GcpProjectId: "my-project",
			Region:       "us-central1",
		},
	}
	op := &operation.AsyncOperation{Id: "op-123"}
	mockCloud.EXPECT().
		UpdateNamespaceExportSink(context.Background(), &cloudservice.UpdateNamespaceExportSinkRequest{
			Namespace:       "my-namespace",
			Spec:            newSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateNamespaceExportSinkResponse{AsyncOperation: op}, nil)

	mockHandler.EXPECT().
		HandleOperation(op, "my-sink").
		Return(nil)

	err := temporalcloudcli.UpdateGCSExportSink(context.Background(), temporalcloudcli.UpdateGCSExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		SaID:             "new-sa@project.iam.gserviceaccount.com",
		BucketName:       "my-bucket",
		GcpProjectID:     "my-project",
		Region:           "us-central1",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.NoError(t, err)
}

func TestUpdateGCSExportSink_GetSinkError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockHandler := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetNamespaceExportSink(context.Background(), &cloudservice.GetNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Name:      "my-sink",
		}).
		Return(nil, apiErr)

	err := temporalcloudcli.UpdateGCSExportSink(context.Background(), temporalcloudcli.UpdateGCSExportSinkParams{
		Namespace:        "my-namespace",
		SinkName:         "my-sink",
		SaID:             "my-sa@project.iam.gserviceaccount.com",
		BucketName:       "my-bucket",
		GcpProjectID:     "my-project",
		Region:           "us-central1",
		Cloud:            mockCloud,
		OperationHandler: mockHandler,
	})
	require.ErrorIs(t, err, apiErr)
}

// --- ValidateGCSExportSink ---

func TestValidateGCSExportSink_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	mockCloud.EXPECT().
		ValidateNamespaceExportSink(context.Background(), &cloudservice.ValidateNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Spec: &namespacev1.ExportSinkSpec{
				Name: "my-sink",
				Gcs:  testGCSSpec(),
			},
		}).
		Return(&cloudservice.ValidateNamespaceExportSinkResponse{}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.ValidateGCSExportSink(context.Background(), temporalcloudcli.ValidateGCSExportSinkParams{
		Namespace:    "my-namespace",
		SinkName:     "my-sink",
		SaID:         "my-sa@project.iam.gserviceaccount.com",
		BucketName:   "my-bucket",
		GcpProjectID: "my-project",
		Region:       "us-central1",
		Cloud:        mockCloud,
		Printer:      newTestPrinter(&buf),
	})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "my-sink")
}

func TestValidateGCSExportSink_Error(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	apiErr := errors.New("invalid config")

	mockCloud.EXPECT().
		ValidateNamespaceExportSink(context.Background(), &cloudservice.ValidateNamespaceExportSinkRequest{
			Namespace: "my-namespace",
			Spec: &namespacev1.ExportSinkSpec{
				Name: "my-sink",
				Gcs:  testGCSSpec(),
			},
		}).
		Return(nil, apiErr)

	var buf bytes.Buffer
	err := temporalcloudcli.ValidateGCSExportSink(context.Background(), temporalcloudcli.ValidateGCSExportSinkParams{
		Namespace:    "my-namespace",
		SinkName:     "my-sink",
		SaID:         "my-sa@project.iam.gserviceaccount.com",
		BucketName:   "my-bucket",
		GcpProjectID: "my-project",
		Region:       "us-central1",
		Cloud:        mockCloud,
		Printer:      newTestPrinter(&buf),
	})
	require.ErrorIs(t, err, apiErr)
}
