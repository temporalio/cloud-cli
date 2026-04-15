package temporalcloudcli

import (
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudAsyncOperationGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetAsyncOperation(cctx, &cloudservice.GetAsyncOperationRequest{
		AsyncOperationId: c.AsyncOperationId,
	})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResource(res.AsyncOperation, printer.PrintResourceOptions{})
}

func (c *CloudAsyncOperationAwaitCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	asyncOpts := AsyncOperationOptions{}
	asyncOpts.PollInterval = c.PollInterval
	return cctx.GetPoller(client, asyncOpts).AwaitAsyncOperation(cctx, c.AsyncOperationId)
}
