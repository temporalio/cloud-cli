package temporalcloudcli

import (
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceRetentionSetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	ns, err := client.GetNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	newSpec := proto.Clone(ns.Spec).(*namespacev1.NamespaceSpec)
	newSpec.RetentionDays = int32(c.RetentionDays)

	err = promptApplyResource(cctx, ns.Spec, newSpec, c.VerboseDiff)
	if err != nil {
		return err
	}

	// Use provided resource version, or fetch from current namespace
	resourceVersion := c.ResourceVersion
	if resourceVersion == "" {
		resourceVersion = ns.ResourceVersion
	}

	updateNamespace := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, client.UpdateNamespace)
	return updateNamespace(namespace.UpdateNamespaceParams{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		AsyncOperationID: c.AsyncOperationId,
		ResourceVersion:  resourceVersion,
	})
}

func (c *CloudNamespaceRetentionGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	ns, err := client.GetNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	// Create focused output showing only retention information
	result := struct {
		Namespace     string `json:"namespace"`
		RetentionDays int32  `json:"retentionDays"`
	}{
		Namespace:     ns.Namespace,
		RetentionDays: ns.Spec.RetentionDays,
	}

	return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
}
