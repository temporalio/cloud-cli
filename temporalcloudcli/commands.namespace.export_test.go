package temporalcloudcli_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	sinkv1 "go.temporal.io/cloud-sdk/api/sink/v1"
	"google.golang.org/protobuf/proto"
)

type (
	exportCommandFactory func() (temporalcloudcli.CommandIfc, *cobra.Command, *bool)
	exportTestProvider   struct {
		name        string
		spec        *namespacev1.ExportSinkSpec
		updatedSpec *namespacev1.ExportSinkSpec
		create      exportCommandFactory
		update      func(string, bool) (temporalcloudcli.CommandIfc, *cobra.Command, *bool)
		validate    exportCommandFactory
	}
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
		SaId:         "my-sa",
		BucketName:   "my-bucket",
		GcpProjectId: "my-project",
		Region:       "us-central1",
	}
}

// testAzureBlobSpec returns a sample Azure Blob spec for use in tests.
func testAzureBlobSpec() *sinkv1.AzureBlobSpec {
	return &sinkv1.AzureBlobSpec{
		TenantId:       "my-tenant",
		SubscriptionId: "my-subscription",
		ResourceGroup:  "my-resource-group",
		StorageAccount: "my-storage-account",
		ContainerName:  "my-container",
		Region:         "eastus",
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

func exportTestProviders() []exportTestProvider {
	return []exportTestProvider{
		{
			name: "S3",
			spec: &namespacev1.ExportSinkSpec{Name: "my-sink", Enabled: true, S3: testS3Spec()},
			updatedSpec: &namespacev1.ExportSinkSpec{Name: "my-sink",
				Enabled: true,
				S3: &sinkv1.S3Spec{RoleName: "new-role",
					AwsAccountId: "987654321012",
					BucketName:   "new-bucket",
					KmsArn:       "new-kms",
					Region:       "us-east-1"}},
			create: func() (temporalcloudcli.CommandIfc, *cobra.Command, *bool) {
				c := &temporalcloudcli.CloudNamespaceExportS3CreateCommand{}
				c.Namespace = "my-namespace"
				c.SinkName = "my-sink"
				c.RoleArn = "arn:aws:iam::123456789012:role/my-role"
				c.BucketName = "my-bucket"
				c.Region = "us-east-1"
				c.AsyncOperationId = "request-id"
				return c, &c.Command, &c.DisableStandaloneActivityExport
			},
			update: func(resourceVersion string, filterOnly bool) (temporalcloudcli.CommandIfc, *cobra.Command, *bool) {
				c := &temporalcloudcli.CloudNamespaceExportS3UpdateCommand{}
				c.Namespace = "my-namespace"
				c.SinkName = "my-sink"
				c.ResourceVersion = resourceVersion
				if !filterOnly {
					c.RoleArn = "arn:aws:iam::987654321012:role/new-role"
					c.BucketName = "new-bucket"
					c.KmsArn = "new-kms"
				}
				c.AsyncOperationId = "request-id"
				return c, &c.Command, &c.DisableStandaloneActivityExport
			},
			validate: func() (temporalcloudcli.CommandIfc, *cobra.Command, *bool) {
				c := &temporalcloudcli.CloudNamespaceExportS3ValidateCommand{}
				c.Namespace = "my-namespace"
				c.SinkName = "my-sink"
				c.RoleArn = "arn:aws:iam::123456789012:role/my-role"
				c.BucketName = "my-bucket"
				c.Region = "us-east-1"
				return c, &c.Command, &c.DisableStandaloneActivityExport
			},
		},
		{
			name: "GCS",
			spec: &namespacev1.ExportSinkSpec{Name: "my-sink", Enabled: true, Gcs: testGCSSpec()},
			updatedSpec: &namespacev1.ExportSinkSpec{Name: "my-sink",
				Enabled: true,
				Gcs: &sinkv1.GCSSpec{SaId: "new-sa",
					GcpProjectId: "new-project",
					BucketName:   "new-bucket",
					Region:       "us-central1"}},
			create: func() (temporalcloudcli.CommandIfc, *cobra.Command, *bool) {
				c := &temporalcloudcli.CloudNamespaceExportGcsCreateCommand{}
				c.Namespace = "my-namespace"
				c.SinkName = "my-sink"
				c.ServiceAccountEmail = "my-sa@my-project.iam.gserviceaccount.com"
				c.BucketName = "my-bucket"
				c.Region = "us-central1"
				c.AsyncOperationId = "request-id"
				return c, &c.Command, &c.DisableStandaloneActivityExport
			},
			update: func(resourceVersion string, filterOnly bool) (temporalcloudcli.CommandIfc, *cobra.Command, *bool) {
				c := &temporalcloudcli.CloudNamespaceExportGcsUpdateCommand{}
				c.Namespace = "my-namespace"
				c.SinkName = "my-sink"
				c.ResourceVersion = resourceVersion
				if !filterOnly {
					c.ServiceAccountEmail = "new-sa@new-project.iam.gserviceaccount.com"
					c.BucketName = "new-bucket"
				}
				c.AsyncOperationId = "request-id"
				return c, &c.Command, &c.DisableStandaloneActivityExport
			},
			validate: func() (temporalcloudcli.CommandIfc, *cobra.Command, *bool) {
				c := &temporalcloudcli.CloudNamespaceExportGcsValidateCommand{}
				c.Namespace = "my-namespace"
				c.SinkName = "my-sink"
				c.ServiceAccountEmail = "my-sa@my-project.iam.gserviceaccount.com"
				c.BucketName = "my-bucket"
				c.Region = "us-central1"
				return c, &c.Command, &c.DisableStandaloneActivityExport
			},
		},
		{
			name: "AzureBlob",
			spec: &namespacev1.ExportSinkSpec{Name: "my-sink", Enabled: true, AzureBlob: testAzureBlobSpec()},
			updatedSpec: &namespacev1.ExportSinkSpec{Name: "my-sink",
				Enabled: true,
				AzureBlob: &sinkv1.AzureBlobSpec{TenantId: "my-tenant",
					SubscriptionId: "my-subscription",
					ResourceGroup:  "my-resource-group",
					StorageAccount: "new-storage-account",
					ContainerName:  "new-container",
					Region:         "eastus"}},
			create: func() (temporalcloudcli.CommandIfc, *cobra.Command, *bool) {
				c := &temporalcloudcli.CloudNamespaceExportAzureBlobCreateCommand{}
				c.Namespace = "my-namespace"
				c.SinkName = "my-sink"
				c.TenantId = "my-tenant"
				c.SubscriptionId = "my-subscription"
				c.ResourceGroup = "my-resource-group"
				c.StorageAccount = "my-storage-account"
				c.ContainerName = "my-container"
				c.Region = "eastus"
				c.AsyncOperationId = "request-id"
				return c, &c.Command, &c.DisableStandaloneActivityExport
			},
			update: func(resourceVersion string, filterOnly bool) (temporalcloudcli.CommandIfc, *cobra.Command, *bool) {
				c := &temporalcloudcli.CloudNamespaceExportAzureBlobUpdateCommand{}
				c.Namespace = "my-namespace"
				c.SinkName = "my-sink"
				c.ResourceVersion = resourceVersion
				if !filterOnly {
					c.StorageAccount = "new-storage-account"
					c.ContainerName = "new-container"
				}
				c.AsyncOperationId = "request-id"
				return c, &c.Command, &c.DisableStandaloneActivityExport
			},
			validate: func() (temporalcloudcli.CommandIfc, *cobra.Command, *bool) {
				c := &temporalcloudcli.CloudNamespaceExportAzureBlobValidateCommand{}
				c.Namespace = "my-namespace"
				c.SinkName = "my-sink"
				c.TenantId = "my-tenant"
				c.SubscriptionId = "my-subscription"
				c.ResourceGroup = "my-resource-group"
				c.StorageAccount = "my-storage-account"
				c.ContainerName = "my-container"
				c.Region = "eastus"
				return c, &c.Command, &c.DisableStandaloneActivityExport
			},
		},
	}
}

// AIDEV-NOTE: Register and parse the bool flag to exercise explicit false versus omission.
func parseExportFilterFlag(t *testing.T, command *cobra.Command, value *bool, flag string) {
	t.Helper()
	command.Flags().BoolVar(value, "disable-standalone-activity-export", false, "")
	if flag != "" {
		require.NoError(t, command.Flags().Parse([]string{"--disable-standalone-activity-export=" + flag}))
	}
}

func TestCreateExportSink_Filter(t *testing.T) {
	for _, provider := range exportTestProviders() {
		for _, tt := range []struct {
			name   string
			flag   string
			filter *namespacev1.ExportSinkSpec_SinkFilter
		}{
			{name: "Omitted"},
			{name: "Disabled", flag: "true", filter: &namespacev1.ExportSinkSpec_SinkFilter{
				DisableStandaloneActivityExport: true,
			}},
			{name: "Enabled", flag: "false", filter: &namespacev1.ExportSinkSpec_SinkFilter{}},
		} {
			t.Run(provider.name+"/"+tt.name, func(t *testing.T) {
				command, cobraCommand, flag := provider.create()
				parseExportFilterFlag(t, cobraCommand, flag, tt.flag)
				expected := proto.Clone(provider.spec).(*namespacev1.ExportSinkSpec)
				expected.Filter = tt.filter
				temporalcloudcli.TestCommand(t, command, temporalcloudcli.TestCommandOptions{
					CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
						c.EXPECT().CreateNamespaceExportSink(mock.Anything, &cloudservice.CreateNamespaceExportSinkRequest{
							Namespace: "my-namespace", Spec: expected, AsyncOperationId: "request-id",
						}, mock.Anything).Return(&cloudservice.CreateNamespaceExportSinkResponse{
							AsyncOperation: &operation.AsyncOperation{Id: "op-123"},
						}, nil)
					},
					PromptOptions: temporalcloudcli.TestPromptOptions{
						ExpectPrompApply: true, PromptResult: true,
						ExpectPromptApplyExisting: &namespacev1.ExportSinkSpec{},
						ExpectPromptApplyModified: expected,
					},
					AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-123"},
				})
			})
		}
	}
}

func TestUpdateExportSink_Filter(t *testing.T) {
	for _, provider := range exportTestProviders() {
		for _, tt := range []struct {
			name            string
			flag            string
			existing        *namespacev1.ExportSinkSpec_SinkFilter
			expected        *namespacev1.ExportSinkSpec_SinkFilter
			enabled         bool
			resourceVersion string
			filterOnly      bool
		}{
			{name: "OmittedWithoutFilter", enabled: true},
			{name: "FilterOnly", flag: "true", filterOnly: true, enabled: true,
				expected: &namespacev1.ExportSinkSpec_SinkFilter{DisableStandaloneActivityExport: true}},
			{name: "OmittedPreservesDisabled", existing: &namespacev1.ExportSinkSpec_SinkFilter{
				DisableStandaloneActivityExport: true,
			}, expected: &namespacev1.ExportSinkSpec_SinkFilter{DisableStandaloneActivityExport: true}},
			{name: "OmittedPreservesEnabled", existing: &namespacev1.ExportSinkSpec_SinkFilter{},
				expected: &namespacev1.ExportSinkSpec_SinkFilter{}, enabled: true},
			{name: "DisableWithoutFilter", flag: "true", expected: &namespacev1.ExportSinkSpec_SinkFilter{
				DisableStandaloneActivityExport: true,
			}, enabled: true},
			{name: "DisableExistingFilter", flag: "true", existing: &namespacev1.ExportSinkSpec_SinkFilter{},
				expected: &namespacev1.ExportSinkSpec_SinkFilter{DisableStandaloneActivityExport: true}, enabled: true},
			{name: "Reenable", flag: "false", existing: &namespacev1.ExportSinkSpec_SinkFilter{
				DisableStandaloneActivityExport: true,
			}, expected: &namespacev1.ExportSinkSpec_SinkFilter{}, resourceVersion: "explicit-rv"},
			{name: "EnableWithoutFilter", flag: "false", expected: &namespacev1.ExportSinkSpec_SinkFilter{}, enabled: true},
		} {
			t.Run(provider.name+"/"+tt.name, func(t *testing.T) {
				command, cobraCommand, flag := provider.update(tt.resourceVersion, tt.filterOnly)
				parseExportFilterFlag(t, cobraCommand, flag, tt.flag)
				existing := proto.Clone(provider.spec).(*namespacev1.ExportSinkSpec)
				existing.Filter = tt.existing
				existing.Enabled = tt.enabled
				original := proto.Clone(existing)
				expected := proto.Clone(provider.updatedSpec).(*namespacev1.ExportSinkSpec)
				if tt.filterOnly {
					expected = proto.Clone(provider.spec).(*namespacev1.ExportSinkSpec)
				}
				expected.Filter = tt.expected
				expected.Enabled = tt.enabled
				rv := "rv-1"
				if tt.resourceVersion != "" {
					rv = tt.resourceVersion
				}
				temporalcloudcli.TestCommand(t, command, temporalcloudcli.TestCommandOptions{
					CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
						c.EXPECT().GetNamespaceExportSink(mock.Anything, &cloudservice.GetNamespaceExportSinkRequest{
							Namespace: "my-namespace", Name: "my-sink",
						}, mock.Anything).Return(&cloudservice.GetNamespaceExportSinkResponse{
							Sink: &namespacev1.ExportSink{Spec: existing, ResourceVersion: "rv-1"},
						}, nil)
						c.EXPECT().UpdateNamespaceExportSink(mock.Anything, &cloudservice.UpdateNamespaceExportSinkRequest{
							Namespace: "my-namespace", Spec: expected, ResourceVersion: rv, AsyncOperationId: "request-id",
						}, mock.Anything).Return(&cloudservice.UpdateNamespaceExportSinkResponse{
							AsyncOperation: &operation.AsyncOperation{Id: "op-123"},
						}, nil)
					},
					PromptOptions: temporalcloudcli.TestPromptOptions{
						ExpectPrompApply: true, PromptResult: true,
						ExpectPromptApplyExisting: existing, ExpectPromptApplyModified: expected,
					},
					AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-123"},
				})
				assert.Equal(t, original, proto.Clone(existing))
			})
		}
	}
}

func TestValidateExportSink_Filter(t *testing.T) {
	for _, provider := range exportTestProviders() {
		for _, tt := range []struct {
			name   string
			flag   string
			filter *namespacev1.ExportSinkSpec_SinkFilter
		}{
			{name: "Omitted"},
			{name: "Disabled", flag: "true", filter: &namespacev1.ExportSinkSpec_SinkFilter{
				DisableStandaloneActivityExport: true,
			}},
			{name: "Enabled", flag: "false", filter: &namespacev1.ExportSinkSpec_SinkFilter{}},
		} {
			t.Run(provider.name+"/"+tt.name, func(t *testing.T) {
				command, cobraCommand, flag := provider.validate()
				parseExportFilterFlag(t, cobraCommand, flag, tt.flag)
				expected := proto.Clone(provider.spec).(*namespacev1.ExportSinkSpec)
				expected.Enabled = false
				expected.Filter = tt.filter
				temporalcloudcli.TestCommand(t, command, temporalcloudcli.TestCommandOptions{
					CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
						c.EXPECT().ValidateNamespaceExportSink(mock.Anything, &cloudservice.ValidateNamespaceExportSinkRequest{
							Namespace: "my-namespace", Spec: expected,
						}, mock.Anything).Return(&cloudservice.ValidateNamespaceExportSinkResponse{}, nil)
					},
					JSONOutput:         true,
					ExpectedOutputJson: struct{ Status string }{Status: `Export sink "my-sink" configuration is valid.`},
				})
			})
		}
	}
}

func TestCreateExportSink_Prompt(t *testing.T) {
	for _, provider := range exportTestProviders() {
		for _, tt := range []struct {
			name          string
			promptError   error
			expectedError string
		}{
			{name: "Declined"},
			{name: "Error", promptError: errors.New("prompt failed"), expectedError: "prompt failed"},
		} {
			t.Run(provider.name+"/"+tt.name, func(t *testing.T) {
				command, _, _ := provider.create()
				temporalcloudcli.TestCommand(t, command, temporalcloudcli.TestCommandOptions{
					PromptOptions: temporalcloudcli.TestPromptOptions{
						ExpectPrompApply: true, PromptError: tt.promptError,
					},
					ExpectedError: tt.expectedError,
				})
			})
		}
	}
}

func TestUpdateExportSink_Prompt(t *testing.T) {
	for _, provider := range exportTestProviders() {
		for _, tt := range []struct {
			name          string
			promptError   error
			expectedError string
		}{
			{name: "Declined"},
			{name: "Error", promptError: errors.New("prompt failed"), expectedError: "prompt failed"},
		} {
			t.Run(provider.name+"/"+tt.name, func(t *testing.T) {
				command, _, _ := provider.update("", false)
				temporalcloudcli.TestCommand(t, command, temporalcloudcli.TestCommandOptions{
					CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
						c.EXPECT().GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
							Return(&cloudservice.GetNamespaceExportSinkResponse{
								Sink: &namespacev1.ExportSink{Spec: provider.spec, ResourceVersion: "rv-1"},
							}, nil)
					},
					PromptOptions: temporalcloudcli.TestPromptOptions{
						ExpectPrompApply: true, PromptError: tt.promptError,
					},
					ExpectedError: tt.expectedError,
				})
			})
		}
	}
}

func TestCreateExportSink_APIError(t *testing.T) {
	for _, provider := range exportTestProviders() {
		t.Run(provider.name, func(t *testing.T) {
			command, _, _ := provider.create()
			temporalcloudcli.TestCommand(t, command, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					c.EXPECT().CreateNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
						Return(nil, errors.New("api failed"))
				},
				PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
				ExpectedError: "api failed",
			})
		})
	}
}

func TestUpdateExportSink_APIError(t *testing.T) {
	for _, provider := range exportTestProviders() {
		t.Run(provider.name, func(t *testing.T) {
			command, _, _ := provider.update("", false)
			temporalcloudcli.TestCommand(t, command, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					c.EXPECT().GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
						Return(&cloudservice.GetNamespaceExportSinkResponse{
							Sink: &namespacev1.ExportSink{Spec: provider.spec, ResourceVersion: "rv-1"},
						}, nil)
					c.EXPECT().UpdateNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
						Return(nil, errors.New("api failed"))
				},
				PromptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
				ExpectedError: "api failed",
			})
		})
	}
}

func TestValidateExportSink_APIError(t *testing.T) {
	for _, provider := range exportTestProviders() {
		t.Run(provider.name, func(t *testing.T) {
			command, _, _ := provider.validate()
			temporalcloudcli.TestCommand(t, command, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					c.EXPECT().ValidateNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
						Return(nil, errors.New("api failed"))
				},
				ExpectedError: "api failed",
			})
		})
	}
}

func TestUpdateExportSink_GetError(t *testing.T) {
	for _, provider := range exportTestProviders() {
		t.Run(provider.name, func(t *testing.T) {
			command, _, _ := provider.update("", false)
			temporalcloudcli.TestCommand(t, command, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					c.EXPECT().GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
						Return(nil, errors.New("get failed"))
				},
				ExpectedError: "get failed",
			})
		})
	}
}
