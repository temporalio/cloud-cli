package temporalcloudcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newTestCommandContext(t *testing.T, buf *bytes.Buffer) *temporalcloudcli.CommandContext {
	t.Helper()
	return &temporalcloudcli.CommandContext{
		Context: context.Background(),
		Printer: &printer.Printer{Output: buf, JSON: true},
	}
}

// TestAsyncOperationHandler_Async verifies that when Async=true, MutationResult is printed immediately without polling.
func TestAsyncOperationHandler_Async(t *testing.T) {
	mockPoller := cmdmock.NewMockPoller(t)
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	cctx.Poller = mockPoller

	runner := temporalcloudcli.NewAsyncOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Async: true}, "my-namespace", temporalcloudcli.ClientOptions{})
	op := &operation.AsyncOperation{Id: "op-async-123"}

	err := runner.Handle(op)
	require.NoError(t, err)

	var out temporalcloudcli.MutationResult
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, temporalcloudcli.MutationResult{AsyncOp: op, ID: "my-namespace"}, out)
}

// TestAsyncOperationHandler_Sync verifies that when Async=false, the poller is called with the correct operation ID.
func TestAsyncOperationHandler_Sync(t *testing.T) {
	mockPoller := cmdmock.NewMockPoller(t)
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	cctx.Poller = mockPoller

	runner := temporalcloudcli.NewAsyncOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Async: false}, "my-namespace", temporalcloudcli.ClientOptions{})
	op := &operation.AsyncOperation{Id: "op-sync-456"}

	mockPoller.EXPECT().PollAsyncOperation(cctx, "op-sync-456", "my-namespace").Return(nil)

	err := runner.Handle(op)
	require.NoError(t, err)
}

// TestAsyncOperationHandler_PollingError verifies that a poller error propagates.
func TestAsyncOperationHandler_PollingError(t *testing.T) {
	pollingErr := errors.New("polling failed")
	mockPoller := cmdmock.NewMockPoller(t)
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	cctx.Poller = mockPoller

	runner := temporalcloudcli.NewAsyncOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Async: false}, "my-namespace", temporalcloudcli.ClientOptions{})
	op := &operation.AsyncOperation{Id: "op-poll-789"}

	mockPoller.EXPECT().PollAsyncOperation(cctx, "op-poll-789", "my-namespace").Return(pollingErr)

	err := runner.Handle(op)
	require.ErrorIs(t, err, pollingErr)
}

// TestAsyncOperationHandler_HandleErr_NothingToChange_Idempotent verifies that a nothing-to-change error with Idempotent=true
// prints the unchanged result without returning an error.
func TestAsyncOperationHandler_HandleErr_NothingToChange_Idempotent(t *testing.T) {
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	nothingToChangeErr := status.Error(codes.InvalidArgument, "nothing to change")

	runner := temporalcloudcli.NewAsyncOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Idempotent: true}, "my-namespace", temporalcloudcli.ClientOptions{})

	err := runner.HandleErr(nothingToChangeErr)
	require.NoError(t, err)

	var out temporalcloudcli.Result
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, temporalcloudcli.Result{Status: "unchanged", ID: "my-namespace"}, out)
}

// TestAsyncOperationHandler_HandleErr_NothingToChange_NotIdempotent verifies that a nothing-to-change error with Idempotent=false
// propagates as an error.
func TestAsyncOperationHandler_HandleErr_NothingToChange_NotIdempotent(t *testing.T) {
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	nothingToChangeErr := status.Error(codes.InvalidArgument, "nothing to change")

	runner := temporalcloudcli.NewAsyncOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Idempotent: false}, "my-namespace", temporalcloudcli.ClientOptions{})

	err := runner.HandleErr(nothingToChangeErr)
	require.Error(t, err)
	assert.Empty(t, buf.String())
}

// TestAsyncOperationHandler_HandleErr_OtherError verifies that non-nothing-to-change errors propagate as-is.
func TestAsyncOperationHandler_HandleErr_OtherError(t *testing.T) {
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	otherErr := errors.New("some other error")

	runner := temporalcloudcli.NewAsyncOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Idempotent: true}, "my-namespace", temporalcloudcli.ClientOptions{})

	err := runner.HandleErr(otherErr)
	require.ErrorIs(t, err, otherErr)
	assert.Empty(t, buf.String())
}
