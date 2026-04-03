package temporalcloudcli_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	sinkv1 "go.temporal.io/cloud-sdk/api/sink/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

func testS3Spec() *sinkv1.S3Spec {
	return &sinkv1.S3Spec{
		RoleName:     "my-role",
		BucketName:   "my-bucket",
		Region:       "us-east-1",
		AwsAccountId: "123456789012",
	}
}

func testGCSSpec() *sinkv1.GCSSpec {
	return &sinkv1.GCSSpec{
		SaId:         "my-sa@project.iam.gserviceaccount.com",
		BucketName:   "my-bucket",
		GcpProjectId: "my-project",
		Region:       "us-central1",
	}
}

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

func TestGetExportSink(t *testing.T) {
	sink := testExportSink(true)
	tests := []struct {
		name                    string
		cmd                     temporalcloudcli.CloudNamespaceExportGetCommand
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
	}{
		{
			name: "Success",
			cmd: temporalcloudcli.CloudNamespaceExportGetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
				ExportSinkOptions: temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, &cloudservice.GetNamespaceExportSinkRequest{
						Namespace: "my-namespace",
						Name:      "my-sink",
					}, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)
			},
		},
		{
			name: "GetSinkError",
			cmd: temporalcloudcli.CloudNamespaceExportGetCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
				ExportSinkOptions: temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"},
			},
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- ListExportSinks ---

func TestListExportSinks(t *testing.T) {
	tests := []struct {
		name                    string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSinks(mock.Anything, &cloudservice.GetNamespaceExportSinksRequest{
						Namespace: "my-namespace",
						PageToken: "",
					}, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinksResponse{
						Sinks: []*namespacev1.ExportSink{testExportSink(true)},
					}, nil)
			},
		},
		{
			name: "Error",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSinks(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("api error"))
			},
			expectedErr: "api error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := temporalcloudcli.CloudNamespaceExportListCommand{
				NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- DeleteExportSink ---

func TestDeleteExportSink_Success(t *testing.T) {
	sink := testExportSink(true)
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceExportDeleteCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
		ExportSinkOptions: temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespaceExportSink(mock.Anything, &cloudservice.GetNamespaceExportSinkRequest{
					Namespace: "my-namespace",
					Name:      "my-sink",
				}, mock.Anything).
				Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)
			c.EXPECT().
				DeleteNamespaceExportSink(mock.Anything, &cloudservice.DeleteNamespaceExportSinkRequest{
					Namespace:       "my-namespace",
					Name:            "my-sink",
					ResourceVersion: "rv-1",
				}, mock.Anything).
				Return(&cloudservice.DeleteNamespaceExportSinkResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-123"},
				}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    true,
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-123"},
	})
}

func TestDeleteExportSink_PromptDeclined(t *testing.T) {
	sink := testExportSink(true)
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceExportDeleteCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
		ExportSinkOptions: temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    false,
		},
		ExpectedError: "Aborting delete.",
	})
}

func TestDeleteExportSink_GetSinkError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceExportDeleteCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
		ExportSinkOptions: temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("not found"))
		},
		ExpectedError: "not found",
	})
}

// --- EnableExportSink ---

func TestEnableExportSink_Success(t *testing.T) {
	sink := testExportSink(false)
	enabledSpec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: true,
		S3:      testS3Spec(),
	}
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceExportEnableCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
		ExportSinkOptions: temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespaceExportSink(mock.Anything, &cloudservice.GetNamespaceExportSinkRequest{
					Namespace: "my-namespace",
					Name:      "my-sink",
				}, mock.Anything).
				Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)
			c.EXPECT().
				UpdateNamespaceExportSink(mock.Anything, &cloudservice.UpdateNamespaceExportSinkRequest{
					Namespace:       "my-namespace",
					Spec:            enabledSpec,
					ResourceVersion: "rv-1",
				}, mock.Anything).
				Return(&cloudservice.UpdateNamespaceExportSinkResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-123"},
				}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply: true,
			PromptResult:     true,
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-123"},
	})
}

func TestEnableExportSink_PromptDeclined(t *testing.T) {
	sink := testExportSink(false)
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceExportEnableCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
		ExportSinkOptions: temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply: true,
			PromptResult:     false,
		},
		ExpectedError: "Aborting enable.",
	})
}

// --- DisableExportSink ---

func TestDisableExportSink_Success(t *testing.T) {
	sink := testExportSink(true)
	disabledSpec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: false,
		S3:      testS3Spec(),
	}
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceExportDisableCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
		ExportSinkOptions: temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespaceExportSink(mock.Anything, &cloudservice.GetNamespaceExportSinkRequest{
					Namespace: "my-namespace",
					Name:      "my-sink",
				}, mock.Anything).
				Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: sink}, nil)
			c.EXPECT().
				UpdateNamespaceExportSink(mock.Anything, &cloudservice.UpdateNamespaceExportSinkRequest{
					Namespace:       "my-namespace",
					Spec:            disabledSpec,
					ResourceVersion: "rv-1",
				}, mock.Anything).
				Return(&cloudservice.UpdateNamespaceExportSinkResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-123"},
				}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply: true,
			PromptResult:     true,
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-123"},
	})
}

// --- S3Create ---

func TestS3CreateExportSink_Success(t *testing.T) {
	spec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: true,
		S3:      testS3Spec(),
	}
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceExportS3CreateCommand{
		NamespaceOptions:  temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
		ExportSinkOptions: temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"},
		ExportS3Options: temporalcloudcli.ExportS3Options{
			RoleName:     "my-role",
			BucketName:   "my-bucket",
			Region:       "us-east-1",
			AwsAccountId: "123456789012",
		},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				CreateNamespaceExportSink(mock.Anything, &cloudservice.CreateNamespaceExportSinkRequest{
					Namespace: "my-namespace",
					Spec:      spec,
				}, mock.Anything).
				Return(&cloudservice.CreateNamespaceExportSinkResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-123"},
				}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    true,
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-123"},
	})
}

func TestS3CreateExportSink_PromptDeclined(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceExportS3CreateCommand{
		NamespaceOptions:  temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
		ExportSinkOptions: temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"},
		ExportS3Options: temporalcloudcli.ExportS3Options{
			RoleName:     "my-role",
			BucketName:   "my-bucket",
			Region:       "us-east-1",
			AwsAccountId: "123456789012",
		},
	}, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    false,
		},
		ExpectedError: "Aborting create.",
	})
}

// --- S3Update ---

func TestS3UpdateExportSink_Success(t *testing.T) {
	existingSink := testExportSink(true)
	newSpec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: true,
		S3: &sinkv1.S3Spec{
			RoleName:     "new-role",
			BucketName:   "new-bucket",
			Region:       "us-west-2",
			AwsAccountId: "123456789012",
		},
	}
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceExportS3UpdateCommand{
		NamespaceOptions:  temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
		ExportSinkOptions: temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"},
		ExportS3Options: temporalcloudcli.ExportS3Options{
			RoleName:     "new-role",
			BucketName:   "new-bucket",
			Region:       "us-west-2",
			AwsAccountId: "123456789012",
		},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespaceExportSink(mock.Anything, &cloudservice.GetNamespaceExportSinkRequest{
					Namespace: "my-namespace",
					Name:      "my-sink",
				}, mock.Anything).
				Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: existingSink}, nil)
			c.EXPECT().
				UpdateNamespaceExportSink(mock.Anything, &cloudservice.UpdateNamespaceExportSinkRequest{
					Namespace:       "my-namespace",
					Spec:            newSpec,
					ResourceVersion: "rv-1",
				}, mock.Anything).
				Return(&cloudservice.UpdateNamespaceExportSinkResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-123"},
				}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPrompApply: true,
			PromptResult:     true,
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-123"},
	})
}

func TestS3UpdateExportSink_GetSinkError(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceExportS3UpdateCommand{
		NamespaceOptions:  temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
		ExportSinkOptions: temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"},
		ExportS3Options: temporalcloudcli.ExportS3Options{
			RoleName:     "my-role",
			BucketName:   "my-bucket",
			Region:       "us-east-1",
			AwsAccountId: "123456789012",
		},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
				Return(nil, errors.New("not found"))
		},
		ExpectedError: "not found",
	})
}

// --- S3Validate ---

func TestS3ValidateExportSink(t *testing.T) {
	tests := []struct {
		name                    string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					ValidateNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.ValidateNamespaceExportSinkResponse{}, nil)
			},
		},
		{
			name: "Error",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					ValidateNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("invalid config"))
			},
			expectedErr: "invalid config",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := temporalcloudcli.CloudNamespaceExportS3ValidateCommand{
				NamespaceOptions:  temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
				ExportSinkOptions: temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"},
				ExportS3Options: temporalcloudcli.ExportS3Options{
					RoleName:     "my-role",
					BucketName:   "my-bucket",
					Region:       "us-east-1",
					AwsAccountId: "123456789012",
				},
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson: func() any {
					if tt.expectedErr != "" {
						return nil
					}
					return map[string]any{"Status": fmt.Sprintf("Export sink %q configuration is valid.", "my-sink")}
				}(),
			})
		})
	}
}

// --- GCSCreate ---

func TestGCSCreateExportSink_Success(t *testing.T) {
	spec := &namespacev1.ExportSinkSpec{
		Name:    "my-sink",
		Enabled: true,
		Gcs:     testGCSSpec(),
	}
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceExportGcsCreateCommand{
		NamespaceOptions:  temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
		ExportSinkOptions: temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"},
		ExportGcsOptions: temporalcloudcli.ExportGcsOptions{
			SaId:         "my-sa@project.iam.gserviceaccount.com",
			BucketName:   "my-bucket",
			GcpProjectId: "my-project",
			Region:       "us-central1",
		},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				CreateNamespaceExportSink(mock.Anything, &cloudservice.CreateNamespaceExportSinkRequest{
					Namespace: "my-namespace",
					Spec:      spec,
				}, mock.Anything).
				Return(&cloudservice.CreateNamespaceExportSinkResponse{
					AsyncOperation: &operation.AsyncOperation{Id: "op-123"},
				}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    true,
		},
		AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-123"},
	})
}

// --- GCSValidate ---

func TestGCSValidateExportSink_Success(t *testing.T) {
	temporalcloudcli.TestCommand(t, &temporalcloudcli.CloudNamespaceExportGcsValidateCommand{
		NamespaceOptions:  temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"},
		ExportSinkOptions: temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"},
		ExportGcsOptions: temporalcloudcli.ExportGcsOptions{
			SaId:         "my-sa@project.iam.gserviceaccount.com",
			BucketName:   "my-bucket",
			GcpProjectId: "my-project",
			Region:       "us-central1",
		},
	}, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				ValidateNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.ValidateNamespaceExportSinkResponse{}, nil)
		},
		JSONOutput: true,
		ExpectedOutputJson: map[string]any{
			"Status": fmt.Sprintf("Export sink %q configuration is valid.", "my-sink"),
		},
	})
}
