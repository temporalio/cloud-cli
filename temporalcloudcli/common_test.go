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

func TestAsyncOperationHandler_Async(t *testing.T) {
	mockPoller := cmdmock.NewMockPoller(t)
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	cctx.Poller = mockPoller

	handler := temporalcloudcli.NewOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Async: true}, temporalcloudcli.ClientOptions{})
	op := &operation.AsyncOperation{Id: "op-async-123"}

	err := handler.HandleOperation(op, "my-namespace")
	require.NoError(t, err)

	var out temporalcloudcli.MutationResult
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, temporalcloudcli.MutationResult{AsyncOp: op, ID: "my-namespace"}, out)
}

func TestAsyncOperationHandler_Sync(t *testing.T) {
	mockPoller := cmdmock.NewMockPoller(t)
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	cctx.Poller = mockPoller

	handler := temporalcloudcli.NewOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Async: false}, temporalcloudcli.ClientOptions{})
	op := &operation.AsyncOperation{Id: "op-sync-456"}

	mockPoller.EXPECT().PollAsyncOperation(cctx, "op-sync-456", "my-namespace").Return(nil)

	err := handler.HandleOperation(op, "my-namespace")
	require.NoError(t, err)
}

func TestAsyncOperationHandler_PollingError(t *testing.T) {
	pollingErr := errors.New("polling failed")
	mockPoller := cmdmock.NewMockPoller(t)
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	cctx.Poller = mockPoller

	handler := temporalcloudcli.NewOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Async: false}, temporalcloudcli.ClientOptions{})
	op := &operation.AsyncOperation{Id: "op-poll-789"}

	mockPoller.EXPECT().PollAsyncOperation(cctx, "op-poll-789", "my-namespace").Return(pollingErr)

	err := handler.HandleOperation(op, "my-namespace")
	require.ErrorIs(t, err, pollingErr)
}

func TestAsyncOperationHandler_HandleErr_NothingToChange_Idempotent(t *testing.T) {
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	nothingToChangeErr := status.Error(codes.InvalidArgument, "nothing to change")

	handler := temporalcloudcli.NewOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Idempotent: true}, temporalcloudcli.ClientOptions{})

	err := handler.HandleUpdateErr(nothingToChangeErr)
	require.NoError(t, err)

	var out temporalcloudcli.Result
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, temporalcloudcli.Result{Status: "unchanged"}, out)
}

func TestAsyncOperationHandler_HandleErr_NothingToChange_NotIdempotent(t *testing.T) {
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	nothingToChangeErr := status.Error(codes.InvalidArgument, "nothing to change")

	handler := temporalcloudcli.NewOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Idempotent: false}, temporalcloudcli.ClientOptions{})

	err := handler.HandleUpdateErr(nothingToChangeErr)
	require.Error(t, err)
	assert.Empty(t, buf.String())
}

func TestAsyncOperationHandler_HandleErr_OtherError(t *testing.T) {
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	otherErr := errors.New("some other error")

	runner := temporalcloudcli.NewOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Idempotent: true}, temporalcloudcli.ClientOptions{})

	err := runner.HandleUpdateErr(otherErr)
	require.ErrorIs(t, err, otherErr)
	assert.Empty(t, buf.String())
}
