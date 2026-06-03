package temporalcloudcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	accountv1 "go.temporal.io/cloud-sdk/api/account/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	sinkv1 "go.temporal.io/cloud-sdk/api/sink/v1"
)

// TestCreateAuditLogSinkKinesis_Success verifies CreateAccountAuditLogSink is called with the correct spec.
func TestCreateAuditLogSinkKinesis_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)

	spec := &accountv1.AuditLogSinkSpec{
		Name: "my-sink",
		SinkType: &accountv1.AuditLogSinkSpec_KinesisSink{
			KinesisSink: &sinkv1.KinesisSpec{
				RoleName:       "MyRole",
				DestinationUri: "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
				Region:         "us-east-1",
			},
		},
		Enabled: true,
	}

	mockPrompter.EXPECT().
		PromptApply(&accountv1.AuditLogSinkSpec{}, spec, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-123"}
	mockCloud.EXPECT().
		CreateAccountAuditLogSink(context.Background(), &cloudservice.CreateAccountAuditLogSinkRequest{
			Spec: spec,
		}).
		Return(&cloudservice.CreateAccountAuditLogSinkResponse{AsyncOperation: op}, nil)

	mockRunner.EXPECT().
		HandleOperation(op, "my-sink").
		Return(nil)

	err := temporalcloudcli.CreateAuditLogSinkKinesis(context.Background(), temporalcloudcli.CreateAuditLogSinkKinesisParams{
		Name:             "my-sink",
		RoleName:         "MyRole",
		DestinationURI:   "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
		Region:           "us-east-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.NoError(t, err)
}

// TestCreateAuditLogSinkKinesis_PromptDeclined verifies CreateAccountAuditLogSink is never called when prompt is declined.
func TestCreateAuditLogSinkKinesis_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	spec := &accountv1.AuditLogSinkSpec{
		Name: "my-sink",
		SinkType: &accountv1.AuditLogSinkSpec_KinesisSink{
			KinesisSink: &sinkv1.KinesisSpec{
				RoleName:       "MyRole",
				DestinationUri: "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
				Region:         "us-east-1",
			},
		},
		Enabled: true,
	}

	mockPrompter.EXPECT().
		PromptApply(&accountv1.AuditLogSinkSpec{}, spec, false).
		Return(promptErr)

	err := temporalcloudcli.CreateAuditLogSinkKinesis(context.Background(), temporalcloudcli.CreateAuditLogSinkKinesisParams{
		Name:             "my-sink",
		RoleName:         "MyRole",
		DestinationURI:   "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
		Region:           "us-east-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.ErrorIs(t, err, promptErr)
}

// TestCreateAuditLogSinkKinesis_CreateError verifies that Runner.HandleCreateErr receives the CreateAccountAuditLogSink error.
func TestCreateAuditLogSinkKinesis_CreateError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	createErr := errors.New("create error")

	spec := &accountv1.AuditLogSinkSpec{
		Name: "my-sink",
		SinkType: &accountv1.AuditLogSinkSpec_KinesisSink{
			KinesisSink: &sinkv1.KinesisSpec{
				RoleName:       "MyRole",
				DestinationUri: "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
				Region:         "us-east-1",
			},
		},
		Enabled: true,
	}

	mockPrompter.EXPECT().
		PromptApply(&accountv1.AuditLogSinkSpec{}, spec, false).
		Return(nil)

	mockCloud.EXPECT().
		CreateAccountAuditLogSink(context.Background(), &cloudservice.CreateAccountAuditLogSinkRequest{
			Spec: spec,
		}).
		Return(nil, createErr)

	mockRunner.EXPECT().
		HandleCreateErr(createErr).
		Return(createErr)

	err := temporalcloudcli.CreateAuditLogSinkKinesis(context.Background(), temporalcloudcli.CreateAuditLogSinkKinesisParams{
		Name:             "my-sink",
		RoleName:         "MyRole",
		DestinationURI:   "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
		Region:           "us-east-1",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.ErrorIs(t, err, createErr)
}

// TestUpdateAuditLogSinkKinesis_Success verifies UpdateAccountAuditLogSink is called with the merged spec.
func TestUpdateAuditLogSinkKinesis_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)

	existingSpec := &accountv1.AuditLogSinkSpec{
		Name: "my-sink",
		SinkType: &accountv1.AuditLogSinkSpec_KinesisSink{
			KinesisSink: &sinkv1.KinesisSpec{
				RoleName:       "OldRole",
				DestinationUri: "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
				Region:         "us-east-1",
			},
		},
		Enabled: true,
	}
	existingSink := &accountv1.AuditLogSink{
		ResourceVersion: "rv-1",
		Spec:            existingSpec,
	}

	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{Name: "my-sink"}).
		Return(&cloudservice.GetAccountAuditLogSinkResponse{Sink: existingSink}, nil)

	updatedSpec := &accountv1.AuditLogSinkSpec{
		Name: "my-sink",
		SinkType: &accountv1.AuditLogSinkSpec_KinesisSink{
			KinesisSink: &sinkv1.KinesisSpec{
				RoleName:       "NewRole",
				DestinationUri: "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
				Region:         "us-east-1",
			},
		},
		Enabled: true,
	}

	mockPrompter.EXPECT().
		PromptApply(existingSpec, updatedSpec, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-456"}
	mockCloud.EXPECT().
		UpdateAccountAuditLogSink(context.Background(), &cloudservice.UpdateAccountAuditLogSinkRequest{
			Spec:            updatedSpec,
			ResourceVersion: "rv-1",
		}).
		Return(&cloudservice.UpdateAccountAuditLogSinkResponse{AsyncOperation: op}, nil)

	mockRunner.EXPECT().
		HandleOperation(op, "my-sink").
		Return(nil)

	err := temporalcloudcli.UpdateAuditLogSinkKinesis(context.Background(), temporalcloudcli.UpdateAuditLogSinkKinesisParams{
		Name:             "my-sink",
		RoleName:         "NewRole",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.NoError(t, err)
}

// TestUpdateAuditLogSinkKinesis_GetSinkError verifies that a GetAccountAuditLogSink error propagates.
func TestUpdateAuditLogSinkKinesis_GetSinkError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{Name: "my-sink"}).
		Return(nil, apiErr)

	err := temporalcloudcli.UpdateAuditLogSinkKinesis(context.Background(), temporalcloudcli.UpdateAuditLogSinkKinesisParams{
		Name:             "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.ErrorIs(t, err, apiErr)
}

// TestUpdateAuditLogSinkKinesis_PromptDeclined verifies UpdateAccountAuditLogSink is never called when prompt is declined.
func TestUpdateAuditLogSinkKinesis_PromptDeclined(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	promptErr := errors.New("Aborting apply.")

	existingSpec := &accountv1.AuditLogSinkSpec{
		Name: "my-sink",
		SinkType: &accountv1.AuditLogSinkSpec_KinesisSink{
			KinesisSink: &sinkv1.KinesisSpec{
				RoleName:       "OldRole",
				DestinationUri: "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
				Region:         "us-east-1",
			},
		},
	}
	existingSink := &accountv1.AuditLogSink{
		ResourceVersion: "rv-1",
		Spec:            existingSpec,
	}

	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{Name: "my-sink"}).
		Return(&cloudservice.GetAccountAuditLogSinkResponse{Sink: existingSink}, nil)

	updatedSpec := &accountv1.AuditLogSinkSpec{
		Name: "my-sink",
		SinkType: &accountv1.AuditLogSinkSpec_KinesisSink{
			KinesisSink: &sinkv1.KinesisSpec{
				RoleName:       "NewRole",
				DestinationUri: "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
				Region:         "us-east-1",
			},
		},
	}

	mockPrompter.EXPECT().
		PromptApply(existingSpec, updatedSpec, false).
		Return(promptErr)

	err := temporalcloudcli.UpdateAuditLogSinkKinesis(context.Background(), temporalcloudcli.UpdateAuditLogSinkKinesisParams{
		Name:             "my-sink",
		RoleName:         "NewRole",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.ErrorIs(t, err, promptErr)
}

// TestUpdateAuditLogSinkKinesis_UpdateError verifies that Runner.HandleUpdateErr receives the UpdateAccountAuditLogSink error.
func TestUpdateAuditLogSinkKinesis_UpdateError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)
	updateErr := errors.New("update error")

	existingSpec := &accountv1.AuditLogSinkSpec{
		Name: "my-sink",
		SinkType: &accountv1.AuditLogSinkSpec_KinesisSink{
			KinesisSink: &sinkv1.KinesisSpec{
				RoleName:       "OldRole",
				DestinationUri: "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
				Region:         "us-east-1",
			},
		},
		Enabled: true,
	}
	existingSink := &accountv1.AuditLogSink{
		ResourceVersion: "rv-1",
		Spec:            existingSpec,
	}

	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{Name: "my-sink"}).
		Return(&cloudservice.GetAccountAuditLogSinkResponse{Sink: existingSink}, nil)

	updatedSpec := &accountv1.AuditLogSinkSpec{
		Name: "my-sink",
		SinkType: &accountv1.AuditLogSinkSpec_KinesisSink{
			KinesisSink: &sinkv1.KinesisSpec{
				RoleName:       "OldRole",
				DestinationUri: "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
				Region:         "us-east-1",
			},
		},
		Enabled: true,
	}

	mockPrompter.EXPECT().
		PromptApply(existingSpec, updatedSpec, false).
		Return(nil)

	mockCloud.EXPECT().
		UpdateAccountAuditLogSink(context.Background(), &cloudservice.UpdateAccountAuditLogSinkRequest{
			Spec:            updatedSpec,
			ResourceVersion: "rv-1",
		}).
		Return(nil, updateErr)

	mockRunner.EXPECT().
		HandleUpdateErr(updateErr).
		Return(updateErr)

	err := temporalcloudcli.UpdateAuditLogSinkKinesis(context.Background(), temporalcloudcli.UpdateAuditLogSinkKinesisParams{
		Name:             "my-sink",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.ErrorIs(t, err, updateErr)
}

// TestUpdateAuditLogSinkKinesis_ResourceVersionOverride verifies that an explicit ResourceVersion overrides the fetched one.
func TestUpdateAuditLogSinkKinesis_ResourceVersionOverride(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	mockPrompter := cmdmock.NewMockPrompter(t)
	mockRunner := cmdmock.NewMockAsyncOperationHandler(t)

	existingSpec := &accountv1.AuditLogSinkSpec{
		Name: "my-sink",
		SinkType: &accountv1.AuditLogSinkSpec_KinesisSink{
			KinesisSink: &sinkv1.KinesisSpec{
				RoleName:       "MyRole",
				DestinationUri: "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
				Region:         "us-east-1",
			},
		},
		Enabled: true,
	}
	existingSink := &accountv1.AuditLogSink{
		ResourceVersion: "rv-1",
		Spec:            existingSpec,
	}

	mockCloud.EXPECT().
		GetAccountAuditLogSink(context.Background(), &cloudservice.GetAccountAuditLogSinkRequest{Name: "my-sink"}).
		Return(&cloudservice.GetAccountAuditLogSinkResponse{Sink: existingSink}, nil)

	updatedSpec := &accountv1.AuditLogSinkSpec{
		Name: "my-sink",
		SinkType: &accountv1.AuditLogSinkSpec_KinesisSink{
			KinesisSink: &sinkv1.KinesisSpec{
				RoleName:       "MyRole",
				DestinationUri: "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
				Region:         "us-east-1",
			},
		},
		Enabled: true,
	}

	mockPrompter.EXPECT().
		PromptApply(existingSpec, updatedSpec, false).
		Return(nil)

	op := &operation.AsyncOperation{Id: "op-789"}
	mockCloud.EXPECT().
		UpdateAccountAuditLogSink(context.Background(), &cloudservice.UpdateAccountAuditLogSinkRequest{
			Spec:            updatedSpec,
			ResourceVersion: "rv-override",
		}).
		Return(&cloudservice.UpdateAccountAuditLogSinkResponse{AsyncOperation: op}, nil)

	mockRunner.EXPECT().
		HandleOperation(op, "my-sink").
		Return(nil)

	err := temporalcloudcli.UpdateAuditLogSinkKinesis(context.Background(), temporalcloudcli.UpdateAuditLogSinkKinesisParams{
		Name:             "my-sink",
		ResourceVersion:  "rv-override",
		Cloud:            mockCloud,
		Prompter:         mockPrompter,
		OperationHandler: mockRunner,
	})
	require.NoError(t, err)
}

// TestValidateAuditLogSinkKinesis_Success verifies ValidateAccountAuditLogSink is called with the correct spec
// and prints a valid status.
func TestValidateAuditLogSinkKinesis_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	spec := &accountv1.AuditLogSinkSpec{
		SinkType: &accountv1.AuditLogSinkSpec_KinesisSink{
			KinesisSink: &sinkv1.KinesisSpec{
				RoleName:       "MyRole",
				DestinationUri: "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
				Region:         "us-east-1",
			},
		},
	}

	mockCloud.EXPECT().
		ValidateAccountAuditLogSink(context.Background(), &cloudservice.ValidateAccountAuditLogSinkRequest{
			Spec: spec,
		}).
		Return(&cloudservice.ValidateAccountAuditLogSinkResponse{}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.ValidateAuditLogSinkKinesis(context.Background(), temporalcloudcli.ValidateAuditLogSinkKinesisParams{
		RoleName:       "MyRole",
		DestinationURI: "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
		Region:         "us-east-1",
		Cloud:          mockCloud,
		Printer:        &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)

	var out struct {
		Status string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "valid", out.Status)
}

// TestValidateAuditLogSinkKinesis_Error verifies that a ValidateAccountAuditLogSink error propagates.
func TestValidateAuditLogSinkKinesis_Error(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	validateErr := errors.New("validation error")

	spec := &accountv1.AuditLogSinkSpec{
		SinkType: &accountv1.AuditLogSinkSpec_KinesisSink{
			KinesisSink: &sinkv1.KinesisSpec{
				RoleName:       "MyRole",
				DestinationUri: "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
				Region:         "us-east-1",
			},
		},
	}

	mockCloud.EXPECT().
		ValidateAccountAuditLogSink(context.Background(), &cloudservice.ValidateAccountAuditLogSinkRequest{
			Spec: spec,
		}).
		Return(nil, validateErr)

	var buf bytes.Buffer
	err := temporalcloudcli.ValidateAuditLogSinkKinesis(context.Background(), temporalcloudcli.ValidateAuditLogSinkKinesisParams{
		RoleName:       "MyRole",
		DestinationURI: "arn:aws:kinesis:us-east-1:123456789012:stream/MyStream",
		Region:         "us-east-1",
		Cloud:          mockCloud,
		Printer:        &printer.Printer{Output: &buf, JSON: true},
	})
	require.ErrorIs(t, err, validateErr)
	assert.Empty(t, buf.String())
}
