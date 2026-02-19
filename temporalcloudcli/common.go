package temporalcloudcli

import (
	"bytes"
	"cmp"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"go.temporal.io/cloud-sdk/cloudclient"
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

// pollAsyncOperation polls an async operation until it reaches a terminal state.
// It prints status updates every second and returns the final AsyncOperation.
//
// The cloudClient should be pre-built using cctx.BuildCloudClient().
//
// AIDEV-NOTE: This function takes a pre-built cloudClient. Commands should
// build the client using cctx.BuildCloudClient() and pass it directly.
func PollAsyncOperation(
	cctx *CommandContext,
	cloudClient *cloudclient.Client,
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
			resp, err := cloudClient.CloudService().GetAsyncOperation(cctx.Context, &cloudservice.GetAsyncOperationRequest{
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
