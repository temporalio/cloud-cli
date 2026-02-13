package temporalcloudcli

import (
	namespace "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceRetentionSetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))

	ns, err := client.getNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	newSpec := proto.Clone(ns.Spec).(*namespace.NamespaceSpec)
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

	res, err := client.applyNamespace(cctx.Context, applyNamespaceParams{
		namespace:        c.Namespace,
		spec:             newSpec,
		asyncOperationID: c.AsyncOperationId,
		resourceVersion:  resourceVersion,
		idempotent:       c.Idempotent,
	})
	if err != nil {
		return err
	}
	if res.asyncOp == nil {
		// Nothing changed (idempotent case)
		result := struct {
			Status    string
			Namespace string
		}{
			Status:    "unchanged",
			Namespace: newSpec.Name,
		}
		return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
	}

	// Handle async flag
	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: res.asyncOp,
			ID:      res.Namespace,
		}, printer.StructuredOptions{})
	}

	// Poll for completion
	return PollAsyncOperation(cctx, cloudClient, res.asyncOp.Id, res.Namespace)
}

func (c *CloudNamespaceRetentionGetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))

	ns, err := client.getNamespace(cctx.Context, c.Namespace)
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
