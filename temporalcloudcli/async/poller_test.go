package async

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	operationv1 "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	csmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

const testOpID = "op-123"

// testPollInterval is short so polling tests complete quickly.
const testPollInterval = time.Millisecond

func newTestPoller(
	t *testing.T,
	mockCloud *csmock.MockCloudServiceClient,
	idempotent bool,
	isAsync bool,
) (Poller, *bytes.Buffer) {
	var buf bytes.Buffer
	p := &printer.Printer{Output: &buf, JSON: true}
	return NewPoller(mockCloud, p, idempotent, isAsync, testPollInterval), &buf
}

// successResp is a minimal response with an async operation, used as a stand-in
// for any create/update/delete response in tests.
func successResp() *cloudservice.UpdateNamespaceResponse {
	return &cloudservice.UpdateNamespaceResponse{
		AsyncOperation: &operationv1.AsyncOperation{Id: testOpID},
	}
}

// mockFulfilledPoll sets up a single GetAsyncOperation call that returns FULFILLED.
func mockFulfilledPoll(mockCloud *csmock.MockCloudServiceClient) {
	mockCloud.EXPECT().
		GetAsyncOperation(
			mock.Anything,
			&cloudservice.GetAsyncOperationRequest{AsyncOperationId: testOpID},
			mock.Anything,
		).
		Return(&cloudservice.GetAsyncOperationResponse{
			AsyncOperation: &operationv1.AsyncOperation{
				Id:    testOpID,
				State: operationv1.AsyncOperation_STATE_FULFILLED,
			},
		}, nil).Once()
}

// ---------------------------------------------------------------------------
// HandleCreateAsyncOperationResponse
// ---------------------------------------------------------------------------

func TestHandleCreateAsyncOperationResponse(t *testing.T) {
	alreadyExistsErr := status.Error(codes.AlreadyExists, "resource already exists")
	otherErr := errors.New("internal error")

	tests := []struct {
		name       string
		idempotent bool
		isAsync    bool
		callErr    error
		response   RespWithAsyncOp
		setupMock  func(*csmock.MockCloudServiceClient)
		wantErr    string
	}{
		{
			name:    "api_error_propagated",
			callErr: otherErr,
			wantErr: "create operation failed: internal error",
		},
		{
			name:       "already_exists_idempotent_returns_nil",
			idempotent: true,
			callErr:    alreadyExistsErr,
		},
		{
			name:       "already_exists_not_idempotent_returns_error",
			idempotent: false,
			callErr:    alreadyExistsErr,
			wantErr:    "create operation failed",
		},
		{
			name:     "success_async_true_returns_immediately",
			isAsync:  true,
			response: successResp(),
		},
		{
			name:      "success_async_false_polls_to_fulfilled",
			isAsync:   false,
			response:  successResp(),
			setupMock: mockFulfilledPoll,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCloud := csmock.NewMockCloudServiceClient(t)
			if tc.setupMock != nil {
				tc.setupMock(mockCloud)
			}

			poller, buf := newTestPoller(t, mockCloud, tc.idempotent, tc.isAsync)
			err := poller.HandleCreateAsyncOperationResponse(context.Background(), tc.response, tc.callErr)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}

			// Async-true and fulfilled polls both print the response as JSON.
			if tc.wantErr == "" && tc.response != nil {
				assert.NotEmpty(t, buf.String())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// HandleUpdateOperation
// ---------------------------------------------------------------------------

func TestHandleUpdateOperation(t *testing.T) {
	nothingToChangeErr := status.Error(codes.InvalidArgument, "nothing to change in this request")
	otherErr := errors.New("update failed")

	tests := []struct {
		name       string
		idempotent bool
		isAsync    bool
		callErr    error
		response   RespWithAsyncOp
		setupMock  func(*csmock.MockCloudServiceClient)
		wantErr    string
	}{
		{
			name:    "api_error_propagated",
			callErr: otherErr,
			wantErr: "update operation failed: update failed",
		},
		{
			name:       "nothing_to_change_idempotent_returns_nil",
			idempotent: true,
			callErr:    nothingToChangeErr,
		},
		{
			name:       "nothing_to_change_not_idempotent_returns_error",
			idempotent: false,
			callErr:    nothingToChangeErr,
			wantErr:    "update operation failed",
		},
		{
			name:     "success_async_true_returns_immediately",
			isAsync:  true,
			response: successResp(),
		},
		{
			name:      "success_async_false_polls_to_fulfilled",
			isAsync:   false,
			response:  successResp(),
			setupMock: mockFulfilledPoll,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCloud := csmock.NewMockCloudServiceClient(t)
			if tc.setupMock != nil {
				tc.setupMock(mockCloud)
			}

			poller, buf := newTestPoller(t, mockCloud, tc.idempotent, tc.isAsync)
			err := poller.HandleUpdateOperation(context.Background(), tc.response, tc.callErr)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tc.wantErr == "" && tc.response != nil {
				assert.NotEmpty(t, buf.String())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// HandleDeleteOperation
// ---------------------------------------------------------------------------

func TestHandleDeleteOperation(t *testing.T) {
	notFoundErr := status.Error(codes.NotFound, "resource not found")
	otherErr := errors.New("delete failed")

	tests := []struct {
		name       string
		idempotent bool
		isAsync    bool
		callErr    error
		response   RespWithAsyncOp
		setupMock  func(*csmock.MockCloudServiceClient)
		wantErr    string
	}{
		{
			name:    "api_error_propagated",
			callErr: otherErr,
			wantErr: "delete operation failed: delete failed",
		},
		{
			name:       "not_found_idempotent_returns_nil",
			idempotent: true,
			callErr:    notFoundErr,
		},
		{
			name:       "not_found_not_idempotent_returns_error",
			idempotent: false,
			callErr:    notFoundErr,
			wantErr:    "delete operation failed",
		},
		{
			name:     "success_async_true_returns_immediately",
			isAsync:  true,
			response: successResp(),
		},
		{
			name:      "success_async_false_polls_to_fulfilled",
			isAsync:   false,
			response:  successResp(),
			setupMock: mockFulfilledPoll,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCloud := csmock.NewMockCloudServiceClient(t)
			if tc.setupMock != nil {
				tc.setupMock(mockCloud)
			}

			poller, buf := newTestPoller(t, mockCloud, tc.idempotent, tc.isAsync)
			err := poller.HandleDeleteOperation(context.Background(), tc.response, tc.callErr)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tc.wantErr == "" && tc.response != nil {
				assert.NotEmpty(t, buf.String())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Polling terminal states (via HandleCreateAsyncOperationResponse, async=false)
// ---------------------------------------------------------------------------

func TestPoll_TerminalStates(t *testing.T) {
	tests := []struct {
		name    string
		state   operationv1.AsyncOperation_State
		reason  string
		wantErr string
	}{
		{
			name:  "fulfilled",
			state: operationv1.AsyncOperation_STATE_FULFILLED,
		},
		{
			name:    "failed",
			state:   operationv1.AsyncOperation_STATE_FAILED,
			reason:  "something went wrong",
			wantErr: "Operation failed: something went wrong",
		},
		{
			name:    "cancelled",
			state:   operationv1.AsyncOperation_STATE_CANCELLED,
			wantErr: "Operation cancelled",
		},
		{
			name:    "rejected",
			state:   operationv1.AsyncOperation_STATE_REJECTED,
			wantErr: "Operation rejected",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCloud := csmock.NewMockCloudServiceClient(t)
			mockCloud.EXPECT().
				GetAsyncOperation(
					mock.Anything,
					&cloudservice.GetAsyncOperationRequest{AsyncOperationId: testOpID},
					mock.Anything,
				).
				Return(&cloudservice.GetAsyncOperationResponse{
					AsyncOperation: &operationv1.AsyncOperation{
						Id:            testOpID,
						State:         tc.state,
						FailureReason: tc.reason,
					},
				}, nil).Once()

			poller, _ := newTestPoller(t, mockCloud, false, false)
			err := poller.HandleCreateAsyncOperationResponse(context.Background(), successResp(), nil)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestPoll_PendingThenFulfilled verifies that the poller keeps polling through
// non-terminal states before eventually resolving.
func TestPoll_PendingThenFulfilled(t *testing.T) {
	mockCloud := csmock.NewMockCloudServiceClient(t)

	req := &cloudservice.GetAsyncOperationRequest{AsyncOperationId: testOpID}

	// First call: PENDING
	mockCloud.EXPECT().
		GetAsyncOperation(mock.Anything, req, mock.Anything).
		Return(&cloudservice.GetAsyncOperationResponse{
			AsyncOperation: &operationv1.AsyncOperation{
				Id:    testOpID,
				State: operationv1.AsyncOperation_STATE_PENDING,
			},
		}, nil).Once()

	// Second call: IN_PROGRESS
	mockCloud.EXPECT().
		GetAsyncOperation(mock.Anything, req, mock.Anything).
		Return(&cloudservice.GetAsyncOperationResponse{
			AsyncOperation: &operationv1.AsyncOperation{
				Id:    testOpID,
				State: operationv1.AsyncOperation_STATE_IN_PROGRESS,
			},
		}, nil).Once()

	// Third call: FULFILLED
	mockCloud.EXPECT().
		GetAsyncOperation(mock.Anything, req, mock.Anything).
		Return(&cloudservice.GetAsyncOperationResponse{
			AsyncOperation: &operationv1.AsyncOperation{
				Id:    testOpID,
				State: operationv1.AsyncOperation_STATE_FULFILLED,
			},
		}, nil).Once()

	poller, _ := newTestPoller(t, mockCloud, false, false)
	err := poller.HandleCreateAsyncOperationResponse(context.Background(), successResp(), nil)
	require.NoError(t, err)
}

// TestPoll_GetAsyncOperationError verifies that a GetAsyncOperation RPC error
// is surfaced to the caller.
func TestPoll_GetAsyncOperationError(t *testing.T) {
	mockCloud := csmock.NewMockCloudServiceClient(t)
	rpcErr := errors.New("transport error")

	mockCloud.EXPECT().
		GetAsyncOperation(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, rpcErr).Once()

	poller, _ := newTestPoller(t, mockCloud, false, false)
	err := poller.HandleCreateAsyncOperationResponse(context.Background(), successResp(), nil)
	require.ErrorContains(t, err, "failed to get async operation status")
}

// TestPoll_ContextCancelled verifies that cancelling the context stops polling.
func TestPoll_ContextCancelled(t *testing.T) {
	mockCloud := csmock.NewMockCloudServiceClient(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before entering the poll loop

	poller, _ := newTestPoller(t, mockCloud, false, false)
	err := poller.HandleCreateAsyncOperationResponse(ctx, successResp(), nil)
	require.ErrorContains(t, err, "operation polling cancelled")
}

func TestSetStructField(t *testing.T) {
	type testStruct struct {
		Field1 string
		Field2 int
	}

	tests := []struct {
		name      string
		input     any
		fieldName string
		value     any
		want      any
		wantErr   string
	}{
		{
			name: "non-pointer input",
			input: testStruct{
				Field1: "hello",
				Field2: 42,
			},
			fieldName: "Field1",
			value:     "new value",
			want:      testStruct{Field1: "hello", Field2: 42}, // should not change
			wantErr:   "input interface is not addressable: {hello 42}",
		},
		{
			name:      "non-struct pointer input",
			input:     &[]string{"not", "a", "struct"},
			fieldName: "Field1",
			value:     "value",
			want:      &[]string{"not", "a", "struct"}, // should not change
			wantErr:   "input is not an pointer to a struct but of type slice",
		},
		{
			name:      "set string field",
			input:     &testStruct{Field1: "old", Field2: 42},
			fieldName: "Field1",
			value:     "new",
			want:      &testStruct{Field1: "new", Field2: 42},
		},
		{
			name:      "set int field",
			input:     &testStruct{Field1: "hello", Field2: 42},
			fieldName: "Field2",
			value:     100,
			want:      &testStruct{Field1: "hello", Field2: 100},
		},
		{
			name:      "field not found",
			input:     &testStruct{},
			fieldName: "NonExistent",
			value:     "value",
			wantErr:   "field NonExistent not found in struct",
		},
		{
			name:      "type mismatch",
			input:     &testStruct{},
			fieldName: "Field2",
			value:     "not an int",
			wantErr:   "type of value does not match type of struct field: int vs string",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := setStructField(tc.input, tc.fieldName, tc.value)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, tc.input)
			}
		})
	}
}
