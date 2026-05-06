package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	sinkv1 "go.temporal.io/cloud-sdk/api/sink/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
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

func nsOpts() temporalcloudcli.NamespaceOptions {
	return temporalcloudcli.NamespaceOptions{Namespace: "my-namespace"}
}

func sinkOpts() temporalcloudcli.ExportSinkOptions {
	return temporalcloudcli.ExportSinkOptions{SinkName: "my-sink"}
}

// --- GetExportSink ---

func TestGetExportSink(t *testing.T) {
	tests := []struct {
		name                    string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, &cloudservice.GetNamespaceExportSinkRequest{
						Namespace: "my-namespace",
						Name:      "my-sink",
					}, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: testExportSink(true)}, nil)
			},
		},
		{
			name: "GetSinkError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("api error"))
			},
			expectedErr: "api error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := temporalcloudcli.CloudNamespaceExportGetCommand{
				NamespaceOptions:  nsOpts(),
				ExportSinkOptions: sinkOpts(),
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
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
					}, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinksResponse{
						Sinks: []*namespacev1.ExportSink{testExportSink(true)},
					}, nil)
			},
		},
		{
			name: "Pagination",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSinks(mock.Anything, &cloudservice.GetNamespaceExportSinksRequest{
						Namespace: "my-namespace",
					}, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinksResponse{
						Sinks:         []*namespacev1.ExportSink{testExportSink(true)},
						NextPageToken: "next",
					}, nil).Once()
				c.EXPECT().
					GetNamespaceExportSinks(mock.Anything, &cloudservice.GetNamespaceExportSinksRequest{
						Namespace: "my-namespace",
						PageToken: "next",
					}, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinksResponse{
						Sinks: []*namespacev1.ExportSink{testExportSink(false)},
					}, nil).Once()
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
				NamespaceOptions: nsOpts(),
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

func TestDeleteExportSink(t *testing.T) {
	tests := []struct {
		name                    string
		resourceVersion         string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, &cloudservice.GetNamespaceExportSinkRequest{
						Namespace: "my-namespace", Name: "my-sink",
					}, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: testExportSink(true)}, nil)
				c.EXPECT().
					DeleteNamespaceExportSink(mock.Anything, &cloudservice.DeleteNamespaceExportSinkRequest{
						Namespace: "my-namespace", Name: "my-sink", ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.DeleteNamespaceExportSinkResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name: "GetSinkError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("sink not found"))
			},
			expectedErr: "sink not found",
		},
		{
			name: "PromptDeclined",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: testExportSink(true)}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting delete.",
		},
		{
			name:            "ResourceVersionOverride",
			resourceVersion: "rv-override",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: testExportSink(true)}, nil)
				c.EXPECT().
					DeleteNamespaceExportSink(mock.Anything, &cloudservice.DeleteNamespaceExportSinkRequest{
						Namespace: "my-namespace", Name: "my-sink", ResourceVersion: "rv-override",
					}, mock.Anything).
					Return(&cloudservice.DeleteNamespaceExportSinkResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name: "DeleteError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: testExportSink(true)}, nil)
				c.EXPECT().
					DeleteNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("delete failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			expectedErr:   "delete operation failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := temporalcloudcli.CloudNamespaceExportDeleteCommand{
				NamespaceOptions:       nsOpts(),
				ExportSinkOptions:      sinkOpts(),
				ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: tt.resourceVersion},
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- EnableExportSink ---

func TestEnableExportSink(t *testing.T) {
	enabledSpec := &namespacev1.ExportSinkSpec{Name: "my-sink", Enabled: true, S3: testS3Spec()}

	tests := []struct {
		name                    string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: testExportSink(false)}, nil)
				c.EXPECT().
					UpdateNamespaceExportSink(mock.Anything, &cloudservice.UpdateNamespaceExportSinkRequest{
						Namespace: "my-namespace", Spec: enabledSpec, ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceExportSinkResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-en"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-en"},
		},
		{
			name: "GetSinkError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("api error"))
			},
			expectedErr: "api error",
		},
		{
			name: "PromptDeclined",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: testExportSink(false)}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting enable.",
		},
		{
			name: "UpdateError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: testExportSink(false)}, nil)
				c.EXPECT().
					UpdateNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("update failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			expectedErr:   "update operation failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := temporalcloudcli.CloudNamespaceExportEnableCommand{
				NamespaceOptions:  nsOpts(),
				ExportSinkOptions: sinkOpts(),
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- DisableExportSink ---

func TestDisableExportSink(t *testing.T) {
	disabledSpec := &namespacev1.ExportSinkSpec{Name: "my-sink", Enabled: false, S3: testS3Spec()}

	tests := []struct {
		name                    string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: testExportSink(true)}, nil)
				c.EXPECT().
					UpdateNamespaceExportSink(mock.Anything, &cloudservice.UpdateNamespaceExportSinkRequest{
						Namespace: "my-namespace", Spec: disabledSpec, ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceExportSinkResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-dis"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-dis"},
		},
		{
			name: "PromptDeclined",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: testExportSink(true)}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting disable.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := temporalcloudcli.CloudNamespaceExportDisableCommand{
				NamespaceOptions:  nsOpts(),
				ExportSinkOptions: sinkOpts(),
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- CreateS3ExportSink ---

func TestCreateS3ExportSink(t *testing.T) {
	expectedSpec := &namespacev1.ExportSinkSpec{Name: "my-sink", Enabled: true, S3: testS3Spec()}

	tests := []struct {
		name                    string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateNamespaceExportSink(mock.Anything, &cloudservice.CreateNamespaceExportSinkRequest{
						Namespace: "my-namespace", Spec: expectedSpec,
					}, mock.Anything).
					Return(&cloudservice.CreateNamespaceExportSinkResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-create"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-create"},
		},
		{
			name:          "PromptDeclined",
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting create.",
		},
		{
			name: "CreateError",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("create failed"))
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			expectedErr:   "create operation failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := temporalcloudcli.CloudNamespaceExportS3CreateCommand{
				NamespaceOptions:  nsOpts(),
				ExportSinkOptions: sinkOpts(),
				ExportS3Options: temporalcloudcli.ExportS3Options{
					RoleName:     "my-role",
					BucketName:   "my-bucket",
					Region:       "us-east-1",
					AwsAccountId: "123456789012",
				},
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- UpdateS3ExportSink ---

func TestUpdateS3ExportSink(t *testing.T) {
	newS3 := &sinkv1.S3Spec{
		RoleName:     "new-role",
		BucketName:   "my-bucket",
		Region:       "us-east-1",
		AwsAccountId: "123456789012",
	}

	tests := []struct {
		name                    string
		roleName                string
		existingEnabled         bool
		expectedEnabled         bool
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name:            "Success",
			roleName:        "new-role",
			existingEnabled: true,
			expectedEnabled: true,
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: testExportSink(true)}, nil)
				c.EXPECT().
					UpdateNamespaceExportSink(mock.Anything, &cloudservice.UpdateNamespaceExportSinkRequest{
						Namespace: "my-namespace",
						Spec:      &namespacev1.ExportSinkSpec{Name: "my-sink", Enabled: true, S3: newS3},
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceExportSinkResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-upd"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-upd"},
		},
		{
			name:            "PreservesEnabledState",
			roleName:        "my-role",
			existingEnabled: false,
			expectedEnabled: false,
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: testExportSink(false)}, nil)
				c.EXPECT().
					UpdateNamespaceExportSink(mock.Anything, &cloudservice.UpdateNamespaceExportSinkRequest{
						Namespace: "my-namespace",
						Spec: &namespacev1.ExportSinkSpec{
							Name: "my-sink", Enabled: false, S3: testS3Spec(),
						},
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceExportSinkResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-upd"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-upd"},
		},
		{
			name:     "GetSinkError",
			roleName: "my-role",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("api error"))
			},
			expectedErr: "api error",
		},
		{
			name:     "PromptDeclined",
			roleName: "new-role",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: testExportSink(true)}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting update.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := temporalcloudcli.CloudNamespaceExportS3UpdateCommand{
				NamespaceOptions:  nsOpts(),
				ExportSinkOptions: sinkOpts(),
				ExportS3Options: temporalcloudcli.ExportS3Options{
					RoleName:     tt.roleName,
					BucketName:   "my-bucket",
					Region:       "us-east-1",
					AwsAccountId: "123456789012",
				},
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- ValidateS3ExportSink ---

func TestValidateS3ExportSink(t *testing.T) {
	expectedSpec := &namespacev1.ExportSinkSpec{Name: "my-sink", S3: testS3Spec()}

	tests := []struct {
		name                    string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					ValidateNamespaceExportSink(mock.Anything, &cloudservice.ValidateNamespaceExportSinkRequest{
						Namespace: "my-namespace", Spec: expectedSpec,
					}, mock.Anything).
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
				NamespaceOptions:  nsOpts(),
				ExportSinkOptions: sinkOpts(),
				ExportS3Options: temporalcloudcli.ExportS3Options{
					RoleName:     "my-role",
					BucketName:   "my-bucket",
					Region:       "us-east-1",
					AwsAccountId: "123456789012",
				},
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- CreateGCSExportSink ---

func TestCreateGCSExportSink(t *testing.T) {
	expectedSpec := &namespacev1.ExportSinkSpec{Name: "my-sink", Enabled: true, Gcs: testGCSSpec()}

	tests := []struct {
		name                    string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					CreateNamespaceExportSink(mock.Anything, &cloudservice.CreateNamespaceExportSinkRequest{
						Namespace: "my-namespace", Spec: expectedSpec,
					}, mock.Anything).
					Return(&cloudservice.CreateNamespaceExportSinkResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-create"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-create"},
		},
		{
			name:          "PromptDeclined",
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting create.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := temporalcloudcli.CloudNamespaceExportGcsCreateCommand{
				NamespaceOptions:  nsOpts(),
				ExportSinkOptions: sinkOpts(),
				ExportGcsOptions: temporalcloudcli.ExportGcsOptions{
					SaId:         "my-sa@project.iam.gserviceaccount.com",
					BucketName:   "my-bucket",
					GcpProjectId: "my-project",
					Region:       "us-central1",
				},
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- UpdateGCSExportSink ---

func TestUpdateGCSExportSink(t *testing.T) {
	existingSink := &namespacev1.ExportSink{
		ResourceVersion: "rv-1",
		Spec: &namespacev1.ExportSinkSpec{
			Name: "my-sink", Enabled: true, Gcs: testGCSSpec(),
		},
	}
	newGCS := &sinkv1.GCSSpec{
		SaId:         "new-sa@project.iam.gserviceaccount.com",
		BucketName:   "my-bucket",
		GcpProjectId: "my-project",
		Region:       "us-central1",
	}

	tests := []struct {
		name                    string
		saID                    string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		promptOptions           temporalcloudcli.TestPromptOptions
		asyncPollerOptions      temporalcloudcli.TestAsyncPollerOptions
		expectedErr             string
	}{
		{
			name: "Success",
			saID: "new-sa@project.iam.gserviceaccount.com",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: existingSink}, nil)
				c.EXPECT().
					UpdateNamespaceExportSink(mock.Anything, &cloudservice.UpdateNamespaceExportSinkRequest{
						Namespace: "my-namespace",
						Spec: &namespacev1.ExportSinkSpec{
							Name: "my-sink", Enabled: true, Gcs: newGCS,
						},
						ResourceVersion: "rv-1",
					}, mock.Anything).
					Return(&cloudservice.UpdateNamespaceExportSinkResponse{
						AsyncOperation: &operation.AsyncOperation{Id: "op-upd"},
					}, nil)
			},
			promptOptions:      temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: true},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-upd"},
		},
		{
			name: "GetSinkError",
			saID: "my-sa@project.iam.gserviceaccount.com",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("api error"))
			},
			expectedErr: "api error",
		},
		{
			name: "PromptDeclined",
			saID: "new-sa@project.iam.gserviceaccount.com",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetNamespaceExportSink(mock.Anything, mock.Anything, mock.Anything).
					Return(&cloudservice.GetNamespaceExportSinkResponse{Sink: existingSink}, nil)
			},
			promptOptions: temporalcloudcli.TestPromptOptions{ExpectPrompApply: true, PromptResult: false},
			expectedErr:   "Aborting update.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := temporalcloudcli.CloudNamespaceExportGcsUpdateCommand{
				NamespaceOptions:  nsOpts(),
				ExportSinkOptions: sinkOpts(),
				ExportGcsOptions: temporalcloudcli.ExportGcsOptions{
					SaId:         tt.saID,
					BucketName:   "my-bucket",
					GcpProjectId: "my-project",
					Region:       "us-central1",
				},
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				PromptOptions:           tt.promptOptions,
				AsyncPollerOptions:      tt.asyncPollerOptions,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}

// --- ValidateGCSExportSink ---

func TestValidateGCSExportSink(t *testing.T) {
	expectedSpec := &namespacev1.ExportSinkSpec{Name: "my-sink", Gcs: testGCSSpec()}

	tests := []struct {
		name                    string
		cloudClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr             string
	}{
		{
			name: "Success",
			cloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					ValidateNamespaceExportSink(mock.Anything, &cloudservice.ValidateNamespaceExportSinkRequest{
						Namespace: "my-namespace", Spec: expectedSpec,
					}, mock.Anything).
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
			cmd := temporalcloudcli.CloudNamespaceExportGcsValidateCommand{
				NamespaceOptions:  nsOpts(),
				ExportSinkOptions: sinkOpts(),
				ExportGcsOptions: temporalcloudcli.ExportGcsOptions{
					SaId:         "my-sa@project.iam.gserviceaccount.com",
					BucketName:   "my-bucket",
					GcpProjectId: "my-project",
					Region:       "us-central1",
				},
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.cloudClientExpectations,
				ExpectedError:           tt.expectedErr,
			})
		})
	}
}
