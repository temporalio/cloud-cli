package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	operationv1 "go.temporal.io/cloud-sdk/api/operation/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

func TestAwaitAsyncOperation(t *testing.T) {
	fulfilledOp := &operationv1.AsyncOperation{
		Id:    "op-123",
		State: operationv1.AsyncOperation_STATE_FULFILLED,
	}

	tests := []struct {
		name               string
		cmd                temporalcloudcli.CloudAsyncOperationAwaitCommand
		asyncPollerOptions temporalcloudcli.TestAsyncPollerOptions
		expectedErr        string
		expectedJsonOutput any
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudAsyncOperationAwaitCommand{AsyncOperationId: "op-123"},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{
				AsyncOperationID: "op-123",
			},
			expectedJsonOutput: fulfilledOp,
		},
		{
			name: "PollError",
			cmd:  temporalcloudcli.CloudAsyncOperationAwaitCommand{AsyncOperationId: "op-123"},
			asyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{
				ErrorToReturn: errors.New("network error"),
			},
			expectedErr: "failed to get async operation status: network error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				AsyncPollerOptions: tt.asyncPollerOptions,
				JSONOutput:         true,
				ExpectedError:      tt.expectedErr,
				ExpectedOutputJson: tt.expectedJsonOutput,
			})
		})
	}
}

func TestGetAsyncOperation(t *testing.T) {
	testOp := &operationv1.AsyncOperation{
		Id:    "op-123",
		State: operationv1.AsyncOperation_STATE_FULFILLED,
	}

	tests := []struct {
		name                  string
		cmd                   temporalcloudcli.CloudAsyncOperationGetCommand
		setClientExpectations func(*cloudmock.MockCloudServiceClient)
		expectedErr           string
		expectedJsonOutput    any
	}{
		{
			name: "Success",
			cmd:  temporalcloudcli.CloudAsyncOperationGetCommand{AsyncOperationId: "op-123"},
			setClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetAsyncOperation(mock.Anything, &cloudservice.GetAsyncOperationRequest{AsyncOperationId: "op-123"}, mock.Anything).
					Return(&cloudservice.GetAsyncOperationResponse{AsyncOperation: testOp}, nil)
			},
			expectedJsonOutput: testOp,
		},
		{
			name: "GetAsyncOperationError",
			cmd:  temporalcloudcli.CloudAsyncOperationGetCommand{AsyncOperationId: "op-123"},
			setClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
				c.EXPECT().
					GetAsyncOperation(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("not found"))
			},
			expectedErr: "not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temporalcloudcli.TestCommand(t, &tt.cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: tt.setClientExpectations,
				JSONOutput:              true,
				ExpectedError:           tt.expectedErr,
				ExpectedOutputJson:      tt.expectedJsonOutput,
			})
		})
	}
}
