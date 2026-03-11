package temporalcloudcli

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

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

func runEditorForJSONEditForProtos(existing, value proto.Message) error {
	marshaler := protojson.MarshalOptions{
		EmitUnpopulated: true,
		Indent:          "    ",
	}
	existingBytes, err := marshaler.Marshal(existing)
	if err != nil {
		return fmt.Errorf("unable to convert existing object to json: %v", err)
	}
	updatedBytes, err := runEditor(existingBytes)
	if err != nil {
		return err
	}
	unmarshaller := protojson.UnmarshalOptions{}
	return unmarshaller.Unmarshal(updatedBytes, value)
}

func runEditor(existing []byte) ([]byte, error) {
	f, err := os.CreateTemp("", "cloud-cli-edit-*.json")
	if err != nil {
		return nil, fmt.Errorf("unable to create temp file for editing: %v", err)
	}

	defer func() {
		// Clean up temp file.
		_ = os.Remove(f.Name())
	}()

	if _, err := f.Write(existing); err != nil {
		return nil, fmt.Errorf("unable to write existing data to temp file for editing: %v", err)
	}

	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("unable to close temp file: %v", err)
	}

	editor := strings.Split(cmp.Or(os.Getenv("VISUAL"), os.Getenv("EDITOR"), "vim"), " ")
	program, args := editor[0], editor[1:]

	cmd := exec.Command(program, append(args, f.Name())...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error executing %q: %v", editor, err)
	}

	updated, err := os.ReadFile(f.Name())
	if err != nil {
		return nil, fmt.Errorf("unable to read updated data from temp file: %v", err)
	}

	if bytes.Equal(existing, updated) {
		return nil, fmt.Errorf("no changes detected")
	}
	return updated, nil
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

type AsyncOperationPoller struct {
	CloudClient cloudservice.CloudServiceClient
}

func setStructField(inputIntf any, fieldName string, value any) error {
	indirectVal := reflect.Indirect(reflect.ValueOf(inputIntf))

	if !indirectVal.CanSet() {
		return fmt.Errorf("input interface is not addressable (can't Set the memory address): %#v",
			inputIntf)
	}
	if indirectVal.Kind() != reflect.Struct {
		return fmt.Errorf("input is not an pointer to a struct but of type %v",
			indirectVal.Kind())
	}

	// allocate each of the structs fields
	var err error
	for i := range indirectVal.NumField() {
		field := indirectVal.Field(i)
		if indirectVal.Type().Field(i).Name == fieldName {
			if field.Type() != reflect.TypeOf(value) {
				return fmt.Errorf("field type mismatch: expected %v, got %v",
					field.Type(), reflect.TypeOf(value))
			}
			field.Set(reflect.ValueOf(value))
		}
	}
	return err
}

func setStructFieldToNil(inputIntf any, fieldName string) error {
	indirectVal := reflect.Indirect(reflect.ValueOf(inputIntf))

	if !indirectVal.CanSet() {
		return fmt.Errorf("input interface is not addressable (can't Set the memory address): %#v",
			inputIntf)
	}
	if indirectVal.Kind() != reflect.Struct {
		return fmt.Errorf("input is not an pointer to a struct but of type %v",
			indirectVal.Kind())
	}

	for i := range indirectVal.NumField() {
		field := indirectVal.Field(i)
		if indirectVal.Type().Field(i).Name == fieldName {
			if field.Kind() != reflect.Pointer {
				return fmt.Errorf("field is not a pointer type: %v", field.Type())
			}
			field.Set(reflect.Zero(field.Type()))
		}
	}
	return nil
}

// Given a response from the cloud ops api,
// the PollAsyncOperation polls on the containing async operation until it reaches a terminal state.
// After which it prints the response with the final state of the async operation.
// If the async operation fails, it returns an error with the failure reason.
//
// The cloudClient should be pre-built using cctx.BuildCloudClient().
//
// AIDEV-NOTE: This function takes a pre-built cloudClient. Commands should
// build the client using cctx.BuildCloudClient() and pass it directly.
func (p *AsyncOperationPoller) PollAsyncOperation(
	cctx *CommandContext,
	response ResponseWithAsyncOp,
) error {
	asyncOpID := response.GetAsyncOperation().GetId()
	if asyncOpID == "" {
		return fmt.Errorf("response does not contain a valid async operation ID, response: '%s'", protojson.MarshalOptions{}.Format(response))
	}
	asyncOp, err := p.PollAsyncOperationByID(cctx, asyncOpID)
	if err != nil {
		return err
	}
	if err := setStructField(response, "AsyncOperation", asyncOp); err != nil {
		return fmt.Errorf("failed to set async operation on response: %w", err)
	}
	// If we have a final operation state, print the details in structured format
	if err := cctx.Printer.PrintStructured(response, printer.StructuredOptions{}); err != nil {
		return err
	}
	return nil
}

func (p *AsyncOperationPoller) PollAsyncOperationByID(
	cctx *CommandContext,
	asyncOpID string,
) (*operation.AsyncOperation, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-cctx.Context.Done():
			return nil, fmt.Errorf("operation polling cancelled: %w", cctx.Context.Err())
		case <-ticker.C:
			// Get the current state of the operation
			resp, err := p.CloudClient.GetAsyncOperation(cctx.Context, &cloudservice.GetAsyncOperationRequest{
				AsyncOperationId: asyncOpID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get async operation status: %w", err)
			}

			asyncOp := resp.GetAsyncOperation()
			if asyncOp == nil {
				// really should never happen, but if it does, return an error
				return nil, fmt.Errorf("async operation not found in get-async-operation response")
			}

			var (
				outString       string // user-friendly string to print at the end of each poll
				outErr          bool   // whether we should return an error
				continuePolling bool   // whether we should continue polling
			)
			switch asyncOp.State {
			case operation.AsyncOperation_STATE_PENDING:
				outString = "Operation pending..."
				continuePolling = true
			case operation.AsyncOperation_STATE_IN_PROGRESS:
				outString = "Operation in progress..."
				continuePolling = true
			case operation.AsyncOperation_STATE_FULFILLED:
				outString = "Operation completed successfully"
			case operation.AsyncOperation_STATE_FAILED:
				outString = fmt.Sprintf("Operation failed: %s", asyncOp.FailureReason)
				outErr = true
			case operation.AsyncOperation_STATE_CANCELLED:
				outString = "Operation cancelled"
				outErr = true
			case operation.AsyncOperation_STATE_REJECTED:
				outString = "Operation rejected"
				outErr = true
			default:
				// This should never happen, but if we get an unknown state, print it and continue polling
				outString = fmt.Sprintf("Unknown operation state: %s, trying again...", asyncOp.State.String())
				continuePolling = true
			}
			if !cctx.JSONOutput {
				cctx.Printer.Print(fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), outString))
			}
			if outErr {
				return nil, errors.New(outString)
			}
			if !continuePolling {
				return asyncOp, nil
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
	ID     string `json:"id"`
}

func newUnchangedResult(id string) Result {
	return Result{
		Status: "unchanged",
		ID:     id,
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
func wrapAsyncOperation[Req any, Resp ResponseWithAsyncOp](
	cctx *CommandContext,
	asyncOpts AsyncOperationOptions,
	clientOpts ClientOptions,
	fn func(context.Context, Req, ...grpc.CallOption) (Resp, error),
) func(Req) error {
	return func(request Req) error {
		resp, err := fn(cctx.Context, request)
		if err != nil {
			if isNothingChangedErr(asyncOpts.Idempotent, err) {
				if err := setStructFieldToNil(resp, "AsyncOperation"); err != nil {
					return fmt.Errorf("failed to set async operation to nil on response: %w", err)
				}
				return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
			}
			return err
		}
		if asyncOpts.Async {
			// For responses without async operation or if async flag is set, print and return immediately
			return cctx.Printer.PrintStructured(resp, printer.StructuredOptions{})
		}

		poller, err := getPoller(cctx, clientOpts)
		if err != nil {
			return err
		}

		return poller.PollAsyncOperation(cctx, resp)
	}
}
