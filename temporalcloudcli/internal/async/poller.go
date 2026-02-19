package async

import (
	"context"
	"fmt"
	"time"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
)

type Poller struct {
	Cloud      cloudservice.CloudServiceClient
	Printer    *printer.Printer
	JSONOutput bool
}

func (p *Poller) Poll(ctx context.Context, operationID string, id string) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation polling cancelled: %w", ctx.Err())
		case <-ticker.C:
			// Get the current state of the operation
			resp, err := p.Cloud.GetAsyncOperation(ctx, &cloudservice.GetAsyncOperationRequest{
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
				return p.Printer.PrintStructured(MutationResult{
					ID:      id,
					AsyncOp: asyncOp,
				}, printer.StructuredOptions{})
			case operation.AsyncOperation_STATE_FAILED:
				progressString = fmt.Sprintf("[%s] Operation failed: %s\n", time.Now().Format("15:04:05"), asyncOp.FailureReason)
				// Print the structured output first, then return error for proper exit code
				if err := p.Printer.PrintStructured(MutationResult{
					ID:      id,
					AsyncOp: asyncOp,
				}, printer.StructuredOptions{}); err != nil {
					return err
				}
				return fmt.Errorf("async operation failed: %s", asyncOp.FailureReason)
			case operation.AsyncOperation_STATE_CANCELLED:
				progressString = fmt.Sprintf("[%s] Operation cancelled\n", time.Now().Format("15:04:05"))
				// Print the structured output first, then return error for proper exit code
				if err := p.Printer.PrintStructured(MutationResult{
					ID:      id,
					AsyncOp: asyncOp,
				}, printer.StructuredOptions{}); err != nil {
					return err
				}
				return fmt.Errorf("async operation cancelled")
			case operation.AsyncOperation_STATE_REJECTED:
				progressString = fmt.Sprintf("[%s] Operation rejected\n", time.Now().Format("15:04:05"))
				// Print the structured output first, then return error for proper exit code
				if err := p.Printer.PrintStructured(MutationResult{
					ID:      id,
					AsyncOp: asyncOp,
				}, printer.StructuredOptions{}); err != nil {
					return err
				}
				return fmt.Errorf("async operation rejected")
			default:
				progressString = fmt.Sprintf("[%s] Operation pending...\n", time.Now().Format("15:04:05"))
			}
			if !p.JSONOutput {
				p.Printer.Print(progressString)
			}
		}
	}
}

type MutationResult struct {
	AsyncOp *operation.AsyncOperation `json:"asyncOperation"`
	ID      string                    `json:"id"`
}
