package temporalcloudcli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

// A FriendlyError is an error intended for presentation directly to users
type FriendlyError struct {
	reason      string
	cause       error
	embedsCause bool
}

func (err FriendlyError) Error() string {
	if err.cause != nil && !err.embedsCause {
		return fmt.Sprintf("%s: %v", err.reason, err.cause)
	} else {
		return err.reason
	}
}

func (err FriendlyError) Unwrap() error {
	return err.cause
}

func (err FriendlyError) FriendlyError() string {
	return err.reason
}

// NewFriendlyError returns an error that preserves the given reason for later presentation
// as a user-facing error message, available via FriendlyError(). Full context is available
// via Error() and Unwrap()
func NewFriendlyError(reason string, cause error) FriendlyError {
	return FriendlyError{reason, cause, false}
}

func NewFriendlyErrorf(format string, args ...any) FriendlyError {
	reason := fmt.Sprintf(format, args...)
	var cause error
	for _, arg := range args {
		if err, ok := arg.(error); ok {
			cause = err
			break
		}
	}
	embedsCause := cause != nil
	return FriendlyError{reason, cause, embedsCause}
}

type GraftedError struct {
	primary error
	grafted error
}

func (err GraftedError) Error() string {
	return err.primary.Error()
}

func (err GraftedError) Unwrap() []error {
	// including primary as well preserves compatibility with things like status.Code
	return []error{err.primary, err.grafted}
}

// GraftErrors encapsulates err and wraps toWrap, as if err had wrapped toWrap originally.
// This is primarily useful when an external library does not wrap errors itself, but the
// original error is available via other means
func GraftErrors(err error, toWrap error) error {
	if err == nil {
		return toWrap
	}
	if toWrap == nil {
		return err
	}
	return GraftedError{err, toWrap}
}

func isNothingChangedErr(idempotent bool, e error) bool {
	// If we are not idempotent, we should error on nothing to change
	if !idempotent {
		return false
	}

	s, ok := status.FromError(e)
	if !ok {
		return false
	}
	return s.Code() == codes.InvalidArgument && strings.Contains(s.Message(), "nothing to change")
}

func isNotFoundErr(e error) bool {
	s, ok := status.FromError(e)
	if !ok {
		return false
	}
	return s.Code() == codes.NotFound
}

// ParseRoleARN extracts the IAM role name and AWS account ID from an IAM role
// ARN of the form "arn:aws:iam::<account>:role/<name>".
func ParseRoleARN(roleARN string) (roleName, accountID string, err error) {
	parsed, err := arn.Parse(roleARN)
	if err != nil {
		return "", "", fmt.Errorf("invalid role ARN %q: %w", roleARN, err)
	}
	if parsed.Service != "iam" {
		return "", "", fmt.Errorf("expected an IAM role ARN, got service %q", parsed.Service)
	}
	name, ok := strings.CutPrefix(parsed.Resource, "role/")
	if !ok || name == "" {
		return "", "", fmt.Errorf("expected resource of the form role/<name>, got %q", parsed.Resource)
	}
	return name, parsed.AccountID, nil
}

// ParseServiceAccountEmail extracts the service account ID (local part) and
// GCP project ID from an email of the form "<sa-id>@<project-id>.iam.gserviceaccount.com".
func ParseServiceAccountEmail(email string) (saID, projectID string, err error) {
	const domainSuffix = ".iam.gserviceaccount.com"
	saID, domain, ok := strings.Cut(email, "@")
	if !ok || saID == "" {
		return "", "", fmt.Errorf("invalid service account email %q: expected <sa-id>@<project-id>.iam.gserviceaccount.com", email)
	}
	projectID, ok = strings.CutSuffix(domain, domainSuffix)
	if !ok || projectID == "" {
		return "", "", fmt.Errorf("invalid service account email %q: expected <sa-id>@<project-id>.iam.gserviceaccount.com", email)
	}
	return saID, projectID, nil
}

// loadJSONSpec loads a JSON specification from either a file path (prefixed with '@')
// or treats the input as inline JSON. Returns the parsed data as a byte slice.
func loadJSONSpec(spec string) ([]byte, error) {
	// Check if spec starts with '@' indicating file path
	if filePath, ok := strings.CutPrefix(spec, "@"); ok {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read spec file %q: %w", filePath, err)
		}
		return data, nil
	}

	// Treat as inline JSON
	return []byte(spec), nil
}

func promptApplyResource(cctx *CommandContext, existing, actual proto.Message, verboseDiff bool) error {
	if !cctx.JSONOutput {
		cctx.Printer.PrintDiff(existing, actual, printer.DiffOptions{
			Verbose: verboseDiff,
		})
	}

	yes, err := cctx.promptYes("Apply (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return fmt.Errorf("Aborting apply.")
	}
	return nil
}

// AsyncOperationHandler handles the async operation lifecycle.
type AsyncOperationHandler interface {
	// HandleOperation dispatches a successfully started operation: prints immediately (async)
	// or polls until completion (sync).
	HandleOperation(op *operation.AsyncOperation, id string) error
	// HandleCreateErr handles an error from a create call: swallows AlreadyExists
	// errors when idempotent, propagates all others.
	HandleCreateErr(err error) error

	HandleUpdateErr(err error) error

	HandleDeleteErr(err error) error
}

type Prompter interface {
	PromptApply(old, new proto.Message, verbose bool) error
}

type operationHandler struct {
	cctx       *CommandContext
	asyncOpts  AsyncOperationOptions
	clientOpts ClientOptions
}

func NewOperationHandler(cctx *CommandContext, asyncOpts AsyncOperationOptions, clientOpts ClientOptions) AsyncOperationHandler {
	return &operationHandler{cctx: cctx, asyncOpts: asyncOpts, clientOpts: clientOpts}
}

func (r *operationHandler) HandleOperation(op *operation.AsyncOperation, resourceID string) error {
	if r.asyncOpts.Async {
		return r.cctx.Printer.PrintStructured(MutationResult{AsyncOp: op, ID: resourceID}, printer.StructuredOptions{})
	}
	poller, pollerErr := getPoller(r.cctx, r.clientOpts)
	if pollerErr != nil {
		return pollerErr
	}
	return poller.PollAsyncOperation(r.cctx, op.Id, resourceID)
}

func (r *operationHandler) HandleUpdateErr(err error) error {
	if isNothingChangedErr(r.asyncOpts.Idempotent, err) {
		return r.cctx.Printer.PrintStructured(newUnchangedResult(), printer.StructuredOptions{})
	}
	return err
}

func (r *operationHandler) HandleCreateErr(err error) error {
	if r.asyncOpts.Idempotent {
		if s, ok := status.FromError(err); ok && s.Code() == codes.AlreadyExists {
			return r.cctx.Printer.PrintStructured(newUnchangedResult(), printer.StructuredOptions{})
		}
	}
	return err
}

func (r *operationHandler) HandleDeleteErr(err error) error {
	if r.asyncOpts.Idempotent {
		if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
			return r.cctx.Printer.PrintStructured(newUnchangedResult(), printer.StructuredOptions{})
		}
	}
	return err
}

type prompter struct {
	cctx *CommandContext
}

func newPrompter(cctx *CommandContext) Prompter {
	return &prompter{cctx: cctx}
}

func (p *prompter) PromptApply(old, new proto.Message, verbose bool) error {
	return promptApplyResource(p.cctx, old, new, verbose)
}

// AsyncOperationResponse is implemented by gRPC responses that carry an async operation.
type AsyncOperationResponse interface {
	GetAsyncOperation() *operation.AsyncOperation
}

// wrapUpdateOperation wraps a gRPC call that returns an AsyncOperationResponse,
// delegating result dispatch and error handling to an AsyncOperationHandler.
func wrapUpdateOperation[Req any, Res AsyncOperationResponse](
	fn func(context.Context, Req, ...grpc.CallOption) (Res, error),
	handler AsyncOperationHandler,
	resourceID string,
) func(context.Context, Req) error {
	return func(ctx context.Context, params Req) error {
		res, err := fn(ctx, params)
		if err != nil {
			return handler.HandleUpdateErr(err)
		}

		return handler.HandleOperation(res.GetAsyncOperation(), resourceID)
	}
}

func wrapCreateOperation[Req any, Res AsyncOperationResponse](
	fn func(context.Context, Req, ...grpc.CallOption) (Res, error),
	handler AsyncOperationHandler,
	idFn func(Res) string,
) func(context.Context, Req) error {
	return func(ctx context.Context, params Req) error {
		res, err := fn(ctx, params)
		if err != nil {
			return handler.HandleCreateErr(err)
		}

		resourceID := idFn(res)
		return handler.HandleOperation(res.GetAsyncOperation(), resourceID)
	}
}

func wrapDeleteOperation[Req any, Res AsyncOperationResponse](
	fn func(context.Context, Req, ...grpc.CallOption) (Res, error),
	handler AsyncOperationHandler,
	resourceID string,
) func(context.Context, Req) error {
	return func(ctx context.Context, params Req) error {
		res, err := fn(ctx, params)
		if err != nil {
			return handler.HandleDeleteErr(err)
		}

		return handler.HandleOperation(res.GetAsyncOperation(), resourceID)
	}
}

type AsyncOperationPoller struct {
	CloudClient cloudservice.CloudServiceClient
}

// pollAsyncOperation polls an async operation until it reaches a terminal state.
// It prints status updates every second and returns the final AsyncOperation.
//
// The cloudClient should be pre-built using cctx.BuildCloudClient().
//
// AIDEV-NOTE: This function takes a pre-built cloudClient. Commands should
// build the client using cctx.BuildCloudClient() and pass it directly.
func (p *AsyncOperationPoller) PollAsyncOperation(
	cctx *CommandContext,
	operationID string,
	id string,
) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cctx.Context.Done():
			return fmt.Errorf("operation polling cancelled: %w", cctx.Context.Err())
		case <-ticker.C:
			// Get the current state of the operation
			resp, err := p.CloudClient.GetAsyncOperation(cctx.Context, &cloudservice.GetAsyncOperationRequest{
				AsyncOperationId: operationID,
			})
			if err != nil {
				return fmt.Errorf("failed to get async operation status: %w", err)
			}

			asyncOp := resp.GetAsyncOperation()
			if asyncOp == nil {
				return fmt.Errorf("async operation not found")
			}

			// Print current state
			var progressString string
			switch asyncOp.State {
			case operation.AsyncOperation_STATE_PENDING:
				progressString = fmt.Sprintf("[%s] Operation pending...\n", time.Now().Format("15:04:05"))
			case operation.AsyncOperation_STATE_IN_PROGRESS:
				progressString = fmt.Sprintf("[%s] Operation in progress...\n", time.Now().Format("15:04:05"))
			case operation.AsyncOperation_STATE_FULFILLED:
				progressString = fmt.Sprintf("[%s] Operation completed successfully\n", time.Now().Format("15:04:05"))
				return cctx.Printer.PrintStructured(MutationResult{
					ID:      id,
					AsyncOp: asyncOp,
				}, printer.StructuredOptions{})
			case operation.AsyncOperation_STATE_FAILED:
				progressString = fmt.Sprintf("[%s] Operation failed: %s\n", time.Now().Format("15:04:05"), asyncOp.FailureReason)
				// Print the structured output first, then return error for proper exit code
				if err := cctx.Printer.PrintStructured(MutationResult{
					ID:      id,
					AsyncOp: asyncOp,
				}, printer.StructuredOptions{}); err != nil {
					return err
				}
				return fmt.Errorf("async operation failed: %s", asyncOp.FailureReason)
			case operation.AsyncOperation_STATE_CANCELLED:
				progressString = fmt.Sprintf("[%s] Operation cancelled\n", time.Now().Format("15:04:05"))
				// Print the structured output first, then return error for proper exit code
				if err := cctx.Printer.PrintStructured(MutationResult{
					ID:      id,
					AsyncOp: asyncOp,
				}, printer.StructuredOptions{}); err != nil {
					return err
				}
				return fmt.Errorf("async operation cancelled")
			case operation.AsyncOperation_STATE_REJECTED:
				progressString = fmt.Sprintf("[%s] Operation rejected\n", time.Now().Format("15:04:05"))
				// Print the structured output first, then return error for proper exit code
				if err := cctx.Printer.PrintStructured(MutationResult{
					ID:      id,
					AsyncOp: asyncOp,
				}, printer.StructuredOptions{}); err != nil {
					return err
				}
				return fmt.Errorf("async operation rejected")
			default:
				progressString = fmt.Sprintf("[%s] Operation pending...\n", time.Now().Format("15:04:05"))
			}
			if !cctx.JSONOutput {
				cctx.Printer.Print(progressString)
			}
		}
	}
}

type MutationResult struct {
	AsyncOp *operation.AsyncOperation `json:"asyncOperation"`
	ID      string                    `json:"id"`
}

type Result struct {
	Status string `json:"status"`
}

func newUnchangedResult() Result {
	return Result{
		Status: "unchanged",
	}
}

func getPoller(cctx *CommandContext, opts ClientOptions) (Poller, error) {
	if cctx.Poller != nil {
		return cctx.Poller, nil
	}
	cloudClient, err := cctx.BuildCloudClient(opts)
	if err != nil {
		return nil, err
	}
	return &AsyncOperationPoller{
		CloudClient: cloudClient.CloudService(),
	}, nil
}

// wrapAsyncOperation wraps an async operation function with standard error handling
// and async operation polling. It returns a function that takes the operation parameters
// and executes the operation with:
//   - Idempotent error handling (returns unchanged result if appropriate)
//   - Async mode support (returns operation ID immediately if async is true)
//   - Automatic polling until completion (if async is false)
func wrapAsyncOperation[P any](
	cctx *CommandContext,
	asyncOpts AsyncOperationOptions,
	resourceID string,
	clientOpts ClientOptions,
	fn func(context.Context, P) (*operation.AsyncOperation, error),
) func(P) error {
	return func(params P) error {
		op, err := fn(cctx.Context, params)
		if err != nil {
			if isNothingChangedErr(asyncOpts.Idempotent, err) {
				return cctx.Printer.PrintStructured(newUnchangedResult(), printer.StructuredOptions{})
			}
			return err
		}

		if asyncOpts.Async {
			// Return immediately with the async operation
			return cctx.Printer.PrintStructured(MutationResult{
				AsyncOp: op,
				ID:      resourceID,
			}, printer.StructuredOptions{})
		}

		poller, err := getPoller(cctx, clientOpts)
		if err != nil {
			return err
		}

		return poller.PollAsyncOperation(cctx, op.Id, resourceID)
	}
}
