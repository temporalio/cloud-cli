package temporalcloudcli

import (
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

	yes, err := cctx.GetPrompter().PromptYes(fmt.Sprintf(
		"Move namespace %q from project %q to project %q",
		c.Namespace, ns.GetProjectId(), c.DestinationProjectId,
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
	return poller.HandleOperation(cctx, resp, err)
}
