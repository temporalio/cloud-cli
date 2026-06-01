package temporalcloudcli

import (
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceFairnessGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}

	enabled := false
	if res.Namespace.Spec.Fairness != nil {
		enabled = res.Namespace.Spec.Fairness.TaskQueueFairnessEnabled
	}

	result := struct {
		Namespace                string `json:"namespace"`
		TaskQueueFairnessEnabled bool   `json:"taskQueueFairnessEnabled"`
	}{
		Namespace:                res.Namespace.Namespace,
		TaskQueueFairnessEnabled: enabled,
	}
	return cctx.Printer.PrintResource(result, printer.PrintResourceOptions{})
}

func (c *CloudNamespaceFairnessSetCommand) run(cctx *CommandContext, _ []string) error {
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
	if newSpec.Fairness == nil {
		newSpec.Fairness = &namespacev1.FairnessSpec{}
	}
	newSpec.Fairness.TaskQueueFairnessEnabled = c.EnableTaskQueueFairness

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
