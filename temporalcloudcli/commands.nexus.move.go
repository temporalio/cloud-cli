package temporalcloudcli

import (
	"errors"
	"fmt"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
)

func (c *CloudNexusEndpointMoveToProjectCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	endpoint, err := resolveNexusEndpoint(cctx, client, c.Name, c.Id)
	if err != nil {
		return err
	}

	names, err := projectDisplayNames(cctx, client, endpoint.GetProjectId(), c.DestinationProjectId)
	if err != nil {
		return err
	}
	if _, ok := names[c.DestinationProjectId]; !ok {
		return fmt.Errorf("destination project %q not found", c.DestinationProjectId)
	}

	yes, err := cctx.GetPrompter().PromptYes(fmt.Sprintf(
		"Move Nexus Endpoint %q from project %s to project %s",
		endpoint.GetSpec().GetName(),
		projectLabel(names, endpoint.GetProjectId()),
		projectLabel(names, c.DestinationProjectId),
	))
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting move.")
	}

	rv := endpoint.GetResourceVersion()
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.MoveNexusEndpointToProject(cctx, &cloudservice.MoveNexusEndpointToProjectRequest{
		EndpointId:              endpoint.GetId(),
		DestinationProjectId:    c.DestinationProjectId,
		ExpectedSourceProjectId: c.SourceProjectId,
		ResourceVersion:         rv,
		AsyncOperationId:        c.AsyncOperationId,
	})
	poller := cctx.GetPoller(client, AsyncOperationOptions{
		AsyncOperationId: c.AsyncOperationId,
		Async:            c.Async,
		PollInterval:     c.PollInterval,
	})
	return poller.HandleIdempotentOperation(cctx, resp, err)
}
