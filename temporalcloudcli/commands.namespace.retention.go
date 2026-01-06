package temporalcloudcli

import (
	"fmt"

	namespace "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceRetentionSetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := newCloudClient(cctx)
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

	cctx.Printer.PrintDiff(ns.Spec, newSpec, printer.DiffOptions{})
	// Step 5: Confirm apply if not forced
	yes, err := cctx.promptYes("Apply (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		fmt.Fprintln(cctx.Printer.Output, "Aborting apply.")
		return nil
	}

	asyncOp, err := client.applyNamespace(cctx.Context, applyNamespaceParams{
		namespace:        c.Namespace,
		spec:             newSpec,
		asyncOperationID: c.AsyncOperationId,
		resourceVersion:  ns.ResourceVersion,
		idempotent:       c.Idempotent,
	})
	if err != nil {
		return err
	}
	if asyncOp == nil {
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
		return cctx.Printer.PrintStructured(asyncOp, printer.StructuredOptions{})
	}

	// Poll for completion
	return pollAsyncOperation(cctx, asyncOp.Id)
}

func (c *CloudNamespaceRetentionGetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := newCloudClient(cctx)
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
