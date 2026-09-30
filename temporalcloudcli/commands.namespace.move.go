package temporalcloudcli

import (
	"context"
	"errors"
	"fmt"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
)

func (c *CloudNamespaceMoveToProjectCommand) run(cctx *CommandContext, _ []string) error {
	if c.NoConnectivityRules && len(c.ConnectivityRuleId) > 0 {
		return errors.New("--connectivity-rule-id and --no-connectivity-rules are mutually exclusive")
	}
	seen := make(map[string]struct{}, len(c.ConnectivityRuleId))
	for _, id := range c.ConnectivityRuleId {
		if _, dup := seen[id]; dup {
			return fmt.Errorf("connectivity rule ID %q specified more than once", id)
		}
		seen[id] = struct{}{}
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	ns := res.Namespace

	names, err := projectDisplayNames(cctx, client, ns.GetProjectId(), c.DestinationProjectId)
	if err != nil {
		return err
	}
	if _, ok := names[c.DestinationProjectId]; !ok {
		return fmt.Errorf("destination project %q not found", c.DestinationProjectId)
	}

	yes, err := cctx.GetPrompter().PromptYes(fmt.Sprintf(
		"Move namespace %q from project %s to project %s",
		c.Namespace, projectLabel(names, ns.GetProjectId()), projectLabel(names, c.DestinationProjectId),
	))
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting move.")
	}

	rv := ns.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	req := &cloudservice.MoveNamespaceToProjectRequest{
		Namespace:               c.Namespace,
		DestinationProjectId:    c.DestinationProjectId,
		ExpectedSourceProjectId: c.SourceProjectId,
		ResourceVersion:         rv,
		AsyncOperationId:        c.AsyncOperationId,
	}
	// Leaving the oneof unset asserts the namespace has no rules to carry over; the server
	// rejects that when it does. It is not the same as selecting no rules.
	switch {
	case c.NoConnectivityRules:
		req.DestinationConnectivityRules = &cloudservice.MoveNamespaceToProjectRequest_Unrestricted{
			Unrestricted: &cloudservice.NoConnectivityRules{},
		}
	case len(c.ConnectivityRuleId) > 0:
		req.DestinationConnectivityRules = &cloudservice.MoveNamespaceToProjectRequest_RuleIds{
			RuleIds: &cloudservice.ConnectivityRuleIDs{ConnectivityRuleIds: c.ConnectivityRuleId},
		}
	}

	resp, err := client.MoveNamespaceToProject(cctx, req)
	poller := cctx.GetPoller(client, AsyncOperationOptions{
		AsyncOperationId: c.AsyncOperationId,
		Async:            c.Async,
		PollInterval:     c.PollInterval,
	})
	return poller.HandleIdempotentOperation(cctx, resp, err)
}

// projectDisplayNames maps project IDs to display names. A project that does not exist, or that
// the caller cannot see, is absent from the result rather than reported as an error.
func projectDisplayNames(
	ctx context.Context,
	client cloudservice.CloudServiceClient,
	ids ...string,
) (map[string]string, error) {
	res, err := client.GetProjects(ctx, &cloudservice.GetProjectsRequest{ProjectIds: ids})
	if err != nil {
		return nil, err
	}
	names := make(map[string]string, len(res.GetProjects()))
	for _, p := range res.GetProjects() {
		names[p.GetId()] = p.GetSpec().GetDisplayName()
	}
	return names, nil
}

func projectLabel(names map[string]string, id string) string {
	if name, ok := names[id]; ok {
		return fmt.Sprintf("%q (%s)", name, id)
	}
	return fmt.Sprintf("%q", id)
}
