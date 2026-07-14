package temporalcloudcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cmdmock "github.com/temporalio/cloud-cli/temporalcloudcli/mock"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CommandHarness struct {
	*require.Assertions
	t       *testing.T
	Options temporalcloudcli.CommandOptions
	// Defaults to a context closed on close or test complete
	Context context.Context
	// Can be used to cancel context given to commands (simulating interrupt)
	CancelContext context.CancelFunc
	Stdin         bytes.Buffer
}

func NewCommandHarness(t *testing.T) *CommandHarness {
	h := &CommandHarness{Assertions: require.New(t), t: t}
	h.Context, h.CancelContext = context.WithCancel(context.Background())
	t.Cleanup(h.Close)
	return h
}

// Reentrant, called after test by default, cancels context
func (h *CommandHarness) Close() {
	// Cancel context
	if h.CancelContext != nil {
		h.CancelContext()
	}
}

// Pieces must appear in order on the line and not overlap
func (h *CommandHarness) ContainsOnSameLine(text string, pieces ...string) {
	h.NoError(AssertContainsOnSameLine(text, pieces...))
}

func AssertContainsOnSameLine(text string, pieces ...string) error {
	// Build regex pattern based on pieces
	pattern := ""
	for _, piece := range pieces {
		if pattern != "" {
			pattern += ".*"
		}
		pattern += regexp.QuoteMeta(piece)
	}
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	// Split into lines, then check each piece is present
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if regex.MatchString(line) {
			return nil
		}
	}
	return fmt.Errorf("pieces not found in order on any line together")
}

func TestAssertContainsOnSameLine(t *testing.T) {
	require.Error(t, AssertContainsOnSameLine("a b c", "b", "a"))
	require.Error(t, AssertContainsOnSameLine("a\nb c", "a", "b"))
	require.NoError(t, AssertContainsOnSameLine("aba", "b", "a"))
	require.NoError(t, AssertContainsOnSameLine("a b a", "b", "a"))
	require.NoError(t, AssertContainsOnSameLine("axb", "a", "b"))
	require.NoError(t, AssertContainsOnSameLine("a a", "a", "a"))
}

func (h *CommandHarness) Eventually(
	condition func() bool,
	waitFor time.Duration,
	tick time.Duration,
	msgAndArgs ...interface{},
) {
	h.t.Helper()
	// We cannot use require.Eventually because it was poorly developed to run the
	// condition function in a goroutine which means it can run after complete or
	// have other race conditions. Don't even need a complicated ticker because it
	// doesn't need to be interruptible.
	for start := time.Now(); time.Since(start) < waitFor; {
		if condition() {
			return
		}
		time.Sleep(tick)
	}
	h.Fail("condition did not evaluate to true within timeout", msgAndArgs...)
}

func (h *CommandHarness) T() *testing.T {
	return h.t
}

type CommandResult struct {
	Err    error
	Stdout bytes.Buffer
	Stderr bytes.Buffer
}

func (h *CommandHarness) Execute(args ...string) *CommandResult {
	// Copy options, update as needed
	res := &CommandResult{}
	options := h.Options
	// Set stdio
	options.Stdin = &h.Stdin
	options.Stdout = &res.Stdout
	options.Stderr = &res.Stderr
	// Set args
	options.Args = args
	// Capture error
	options.Fail = func(err error) {
		if res.Err != nil {
			panic("fail called twice, just failed with " + err.Error())
		}
		res.Err = err
	}

	// Run
	ctx, cancel := context.WithCancel(h.Context)
	h.t.Cleanup(cancel)
	defer cancel()
	h.t.Logf("Calling: %v", strings.Join(args, " "))
	temporalcloudcli.Execute(ctx, options)
	if res.Stdout.Len() > 0 {
		h.t.Logf("Stdout:\n-----\n%s\n-----", &res.Stdout)
	}
	if res.Stderr.Len() > 0 {
		h.t.Logf("Stderr:\n-----\n%s\n-----", &res.Stderr)
	}
	return res
}

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

func TestAsyncOperationHandler_HandleUpdateErr_NothingToChange_Idempotent(t *testing.T) {
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

func TestAsyncOperationHandler_HandleUpdateErr_NothingToChange_NotIdempotent(t *testing.T) {
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	nothingToChangeErr := status.Error(codes.InvalidArgument, "nothing to change")

	handler := temporalcloudcli.NewOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Idempotent: false}, temporalcloudcli.ClientOptions{})

	err := handler.HandleUpdateErr(nothingToChangeErr)
	require.Error(t, err)
	assert.Empty(t, buf.String())
}

func TestAsyncOperationHandler_HandleUpdateErr_OtherError(t *testing.T) {
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	otherErr := errors.New("some other error")

	handler := temporalcloudcli.NewOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Idempotent: true}, temporalcloudcli.ClientOptions{})

	err := handler.HandleUpdateErr(otherErr)
	require.ErrorIs(t, err, otherErr)
	assert.Empty(t, buf.String())
}

func TestAsyncOperationHandler_HandleCreateErr_AlreadyExists_Idempotent(t *testing.T) {
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	alreadyExistsErr := status.Error(codes.AlreadyExists, "already exists")

	handler := temporalcloudcli.NewOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Idempotent: true}, temporalcloudcli.ClientOptions{})

	err := handler.HandleCreateErr(alreadyExistsErr)
	require.NoError(t, err)

	var out temporalcloudcli.Result
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, temporalcloudcli.Result{Status: "unchanged"}, out)
}

func TestAsyncOperationHandler_HandleCreateErr_AlreadyExists_NotIdempotent(t *testing.T) {
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	alreadyExistsErr := status.Error(codes.AlreadyExists, "already exists")

	handler := temporalcloudcli.NewOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Idempotent: false}, temporalcloudcli.ClientOptions{})

	err := handler.HandleCreateErr(alreadyExistsErr)
	require.ErrorIs(t, err, alreadyExistsErr)
	assert.Empty(t, buf.String())
}

func TestAsyncOperationHandler_HandleDeleteErr_NotFound_Idempotent(t *testing.T) {
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	notFoundErr := status.Error(codes.NotFound, "not found")

	handler := temporalcloudcli.NewOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Idempotent: true}, temporalcloudcli.ClientOptions{})

	err := handler.HandleDeleteErr(notFoundErr)
	require.NoError(t, err)

	var out temporalcloudcli.Result
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, temporalcloudcli.Result{Status: "unchanged"}, out)
}

func TestAsyncOperationHandler_HandleDeleteErr_NotFound_NotIdempotent(t *testing.T) {
	var buf bytes.Buffer
	cctx := newTestCommandContext(t, &buf)
	notFoundErr := status.Error(codes.NotFound, "not found")

	handler := temporalcloudcli.NewOperationHandler(cctx, temporalcloudcli.AsyncOperationOptions{Idempotent: false}, temporalcloudcli.ClientOptions{})

	err := handler.HandleDeleteErr(notFoundErr)
	require.ErrorIs(t, err, notFoundErr)
	assert.Empty(t, buf.String())
}

func TestParseRoleARN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		input           string
		wantRoleName    string
		wantAccountID   string
		wantErrContains string
	}{
		{
			name:          "valid",
			input:         "arn:aws:iam::123456789012:role/my-role",
			wantRoleName:  "my-role",
			wantAccountID: "123456789012",
		},
		{
			name:          "valid with path",
			input:         "arn:aws:iam::123456789012:role/path/to/role",
			wantRoleName:  "path/to/role",
			wantAccountID: "123456789012",
		},
		{
			name:            "empty",
			input:           "",
			wantErrContains: "invalid role ARN",
		},
		{
			name:            "wrong service",
			input:           "arn:aws:s3:::my-bucket",
			wantErrContains: `expected an IAM role ARN, got service "s3"`,
		},
		{
			name:            "missing role/ prefix",
			input:           "arn:aws:iam::123456789012:user/jane",
			wantErrContains: `expected resource of the form role/<name>`,
		},
		{
			name:            "empty role name",
			input:           "arn:aws:iam::123456789012:role/",
			wantErrContains: `expected resource of the form role/<name>`,
		},
		{
			name:            "not an arn",
			input:           "not-an-arn",
			wantErrContains: "invalid role ARN",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			roleName, accountID, err := temporalcloudcli.ParseRoleARN(tt.input)
			if tt.wantErrContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantRoleName, roleName)
			assert.Equal(t, tt.wantAccountID, accountID)
		})
	}
}

func TestParseServiceAccountEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		input           string
		wantSaID        string
		wantProjectID   string
		wantErrContains string
	}{
		{
			name:          "valid",
			input:         "my-sa@my-project.iam.gserviceaccount.com",
			wantSaID:      "my-sa",
			wantProjectID: "my-project",
		},
		{
			name:            "empty",
			input:           "",
			wantErrContains: "invalid service account email",
		},
		{
			name:            "missing @",
			input:           "my-sa.iam.gserviceaccount.com",
			wantErrContains: "invalid service account email",
		},
		{
			name:            "wrong domain suffix",
			input:           "my-sa@my-project.example.com",
			wantErrContains: "invalid service account email",
		},
		{
			name:            "empty local part",
			input:           "@my-project.iam.gserviceaccount.com",
			wantErrContains: "invalid service account email",
		},
		{
			name:            "empty project part",
			input:           "my-sa@.iam.gserviceaccount.com",
			wantErrContains: "invalid service account email",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			saID, projectID, err := temporalcloudcli.ParseServiceAccountEmail(tt.input)
			if tt.wantErrContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantSaID, saID)
			assert.Equal(t, tt.wantProjectID, projectID)
		})
	}
}

func TestFriendlyError(t *testing.T) {
	originalErr := fmt.Errorf("details")
	err := temporalcloudcli.NewFriendlyError("a simple error", originalErr)
	formattedErr := temporalcloudcli.NewFriendlyErrorf("An error with %v inline", originalErr)

	t.Run("FriendlyErrorFormatting", func(t *testing.T) {
		assert.Equal(t, "a simple error", err.FriendlyError())
		assert.Equal(t, "An error with details inline", formattedErr.FriendlyError())
	})

	t.Run("ErrorFormatting", func(t *testing.T) {
		assert.Equal(t, "a simple error: details", err.Error())
		assert.Equal(t, "An error with details inline", formattedErr.Error())
	})

	t.Run("IsOriginalError", func(t *testing.T) {
		assert.ErrorIs(t, err, originalErr)
		assert.ErrorIs(t, formattedErr, originalErr)
	})
}

func TestErrorGrafting(t *testing.T) {
	originalErr := fmt.Errorf("some complex error message")
	friendlyErr := temporalcloudcli.NewFriendlyError("busted", originalErr)
	gRPCErr := status.Errorf(codes.Internal, "something bad happened: %v", friendlyErr)
	graftedErr := temporalcloudcli.GraftErrors(gRPCErr, friendlyErr)

	t.Run("DoesNotModifyErrorMessage", func(t *testing.T) {
		assert.Equal(t, gRPCErr.Error(), graftedErr.Error())
		assert.NotRegexp(t, "busted.*busted", graftedErr.Error(), "should not duplicate the message in the grafted error")
	})

	t.Run("LooksLikeGRPCError", func(t *testing.T) {
		assert.Equal(t, codes.Internal, status.Code(graftedErr))
		s, ok := status.FromError(graftedErr)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, s.Code())
	})

	t.Run("IsBothErrors", func(t *testing.T) {
		assert.ErrorIs(t, graftedErr, gRPCErr)
		assert.ErrorIs(t, graftedErr, friendlyErr)
	})

	t.Run("AsKnownErrorType", func(t *testing.T) {
		var target temporalcloudcli.FriendlyError
		assert.ErrorAs(t, graftedErr, &target)
		assert.Equal(t, friendlyErr.FriendlyError(), target.FriendlyError())
	})
}
