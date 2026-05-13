package temporalcloudcli

import (
	"errors"
	"fmt"
	"slices"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	connectivityrulev1 "go.temporal.io/cloud-sdk/api/connectivityrule/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceConnectivityListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResourceList(
		struct {
			ConnectivityRules []*connectivityrulev1.ConnectivityRule
		}{
			ConnectivityRules: res.Namespace.GetConnectivityRules(),
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Id", "State"},
			SpecFields: []string{},
		},
		printer.TableOptions{},
	)
}

func (c *CloudNamespaceConnectivityAttachCommand) run(cctx *CommandContext, _ []string) error {
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
	for _, id := range c.ConnectivityRuleId {
		if slices.Contains(ns.Spec.GetConnectivityRuleIds(), id) {
			return fmt.Errorf("connectivity rule %q is already attached to namespace %q", id, c.Namespace)
		}
	}
	newSpec := proto.Clone(ns.Spec).(*namespacev1.NamespaceSpec)
	newSpec.ConnectivityRuleIds = append(newSpec.ConnectivityRuleIds, c.ConnectivityRuleId...)

	yes, err := cctx.GetPrompter().PromptApply(ns.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting attach.")
	}

	rv := ns.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateNamespace(cctx, &cloudservice.UpdateNamespaceRequest{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNamespaceConnectivityDetachCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	ns := res.Namespace
	newSpec := proto.Clone(ns.Spec).(*namespacev1.NamespaceSpec)
	newSpec.ConnectivityRuleIds = slices.DeleteFunc(newSpec.ConnectivityRuleIds, func(id string) bool {
		return slices.Contains(c.ConnectivityRuleId, id)
	})

	yes, err := cctx.GetPrompter().PromptApply(ns.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting detach.")
	}

	rv := ns.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateNamespace(cctx, &cloudservice.UpdateNamespaceRequest{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}
