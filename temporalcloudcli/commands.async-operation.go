package temporalcloudcli

import (
	"github.com/temporalio/cloud-cli/internal/asyncoperation"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudAsyncOperationGetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	client := asyncoperation.NewClient(cloudClient.CloudService())

	op, err := client.GetAsyncOperation(cctx.Context, c.AsyncOperationId)
	if err != nil {
		return err
	}

	return cctx.Printer.PrintStructured(op, printer.StructuredOptions{})
}

func (c *CloudAsyncOperationAwaitCommand) run(cctx *CommandContext, _ []string) error {
	poller, err := getPoller(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	return poller.PollAsyncOperation(cctx, c.AsyncOperationId, c.AsyncOperationId)
}
