package async

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	"go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type (
	Poller interface {
		HandleCreateAsyncOperationResponse(ctx context.Context, response RespWithAsyncOp, err error) error
		HandleUpdateOperation(ctx context.Context, response RespWithAsyncOp, err error) error
		HandleDeleteOperation(ctx context.Context, response RespWithAsyncOp, err error) error
	}

	poller struct {
		cloudService cloudservice.CloudServiceClient
		printer      *printer.Printer
		idempotent   bool
		async        bool
		pollInterval time.Duration
	}

	RespWithAsyncOp interface {
		GetAsyncOperation() *operation.AsyncOperation
	}
)

func NewPoller(
	cloudService cloudservice.CloudServiceClient,
	printer *printer.Printer,
	idempotent bool,
	async bool,
	pollInterval time.Duration,
) Poller {
	return &poller{
		cloudService: cloudService,
		printer:      printer,
		idempotent:   idempotent,
		async:        async,
		pollInterval: pollInterval,
	}
}

func (p *poller) HandleCreateAsyncOperationResponse(
	ctx context.Context,
	response RespWithAsyncOp,
	err error,
) error {
	if s, ok := status.FromError(err); ok && s.Code() == codes.AlreadyExists && p.idempotent {
		p.printer.Println("Resource already exists")
		return nil
	} else if err != nil {
		return fmt.Errorf("create operation failed: %w", err)
	}
	return p.handleAsyncOperation(ctx, response)
}

func (p *poller) HandleUpdateOperation(
	ctx context.Context,
	response RespWithAsyncOp,
	err error,
) error {
	if s, ok := status.FromError(err); p.idempotent && ok &&
		s.Code() == codes.InvalidArgument &&
		strings.Contains(s.Message(), "nothing to change") {
		p.printer.Println("Resource already in desired state")
		return nil
	} else if err != nil {
		return fmt.Errorf("update operation failed: %w", err)
	}
	return p.handleAsyncOperation(ctx, response)
}

func (p *poller) HandleDeleteOperation(
	ctx context.Context,
	response RespWithAsyncOp,
	err error,
) error {
	if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound && p.idempotent {
		p.printer.Println("Resource already deleted")
		return nil
	} else if err != nil {
		return fmt.Errorf("delete operation failed: %w", err)
	}
	return p.handleAsyncOperation(ctx, response)
}

func (p *poller) handleAsyncOperation(
	ctx context.Context,
	response RespWithAsyncOp,
) error {
	if p.async {
		// Return immediately with the async operation
		return p.printer.PrintResponseWithAsyncOperation(response, printer.PrintResourceOptions{})
	}
	ao := response.GetAsyncOperation()
	if ao == nil {
		// really should never happen, but if it does, return an error
		return fmt.Errorf("async operation not found in response")
	}

	// Poll the async operation until it completes
	finalOp, err := p.waitForAsyncOperation(ctx, ao.Id)
	if err != nil {
		return fmt.Errorf("failed to wait for async operation: %w", err)
	}
	// Update the response's async operation to the final state before printing,
	// so the user can see the final state in the output and parse it if they want.
	if err := setStructField(response, "AsyncOperation", finalOp); err != nil {
		return fmt.Errorf("failed to set async operation on response: %w", err)
	}
	return p.printer.PrintResponseWithAsyncOperation(response, printer.PrintResourceOptions{})
}

func (p *poller) waitForAsyncOperation(
	ctx context.Context,
	asyncOpID string,
) (*operation.AsyncOperation, error) {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("operation polling cancelled: %w", ctx.Err())
		case <-ticker.C:
			// Get the current state of the operation
			resp, err := p.cloudService.GetAsyncOperation(ctx, &cloudservice.GetAsyncOperationRequest{
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
			p.printer.Print(fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), outString))
			if !continuePolling {
				if outErr {
					return nil, errors.New(outString)
				}
				return asyncOp, nil
			}
		}
	}
}

func setStructField(inputIntf any, fieldName string, value any) error {
	indirectVal := reflect.Indirect(reflect.ValueOf(inputIntf))

	if !indirectVal.CanSet() {
		return fmt.Errorf("input interface is not addressable: %v",
			inputIntf)
	}
	if indirectVal.Kind() != reflect.Struct {
		return fmt.Errorf("input is not an pointer to a struct but of type %v",
			indirectVal.Kind())
	}

	// allocate each of the structs fields
	var err error
	var allocatedField bool
	for i := range indirectVal.NumField() {
		field := indirectVal.Field(i)
		if indirectVal.Type().Field(i).Name == fieldName {
			if field.Type() != reflect.TypeOf(value) {
				return fmt.Errorf("type of value does not match type of struct field: %v vs %v",
					field.Type(), reflect.TypeOf(value))
			}
			field.Set(reflect.ValueOf(value))
			allocatedField = true
		}
	}
	if !allocatedField {
		return fmt.Errorf("field %s not found in struct", fieldName)
	}
	return err
}
