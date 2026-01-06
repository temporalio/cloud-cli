package temporalcloudcli

import (
	"fmt"

	namespace "go.temporal.io/cloud-sdk/api/namespace/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceGetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := newCloudClient(cctx)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))

	n, err := client.getNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	if c.Spec {
		return cctx.Printer.PrintStructured(n.Spec, printer.StructuredOptions{})
	}
	return cctx.Printer.PrintStructured(n, printer.StructuredOptions{})
}

func (c *CloudNamespaceEditCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := newCloudClient(cctx)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))

	ns, err := client.getNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	newSpec := &namespace.NamespaceSpec{}
	err = runEditorForJSONEditForProtos(ns.Spec, newSpec)
	if err != nil {
		return err
	}

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

	// TODO: (gmankes) remove this -- clean up and make shareable
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

func (c *CloudNamespaceApplyCommand) run(cctx *CommandContext, _ []string) error {
	// Step 1: Load spec from file or inline
	specData, err := loadJSONSpec(c.Spec)
	if err != nil {
		return err
	}

	// Step 2: Parse JSON into NamespaceSpec using protobuf JSON unmarshaling
	spec := &namespace.NamespaceSpec{}
	if err := cctx.UnmarshalProtoJSON(specData, spec); err != nil {
		return fmt.Errorf("failed to parse JSON spec: %w", err)
	}

	// Step 3: Create cloud and namespace clients
	cloudClient, err := newCloudClient(cctx)
	if err != nil {
		return err
	}
	client := newNamespaceClient(withCloudClient(cloudClient))

	// Step 4: Retrieve existing namespace
	existing, err := client.getNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}
	cctx.Printer.PrintDiff(existing.Spec, spec, printer.DiffOptions{
		Verbose: c.VerboseDiff,
	})

	// Step 5: Confirm apply if not forced
	yes, err := cctx.promptYes("Apply (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		fmt.Fprintln(cctx.Printer.Output, "Aborting apply.")
		return nil
	}

	// Step 5: Apply the namespace (create or update)
	params := applyNamespaceParams{
		namespace: c.Namespace,
		spec:      spec,

		resourceVersion:  existing.ResourceVersion,
		asyncOperationID: c.AsyncOperationId, // Use the flag value if provided
		idempotent:       c.Idempotent,       // Use the flag value
	}

	asyncOp, err := client.applyNamespace(cctx.Context, params)
	if err != nil {
		return fmt.Errorf("failed to apply namespace: %w", err)
	}

	// Step 5: Handle result
	if asyncOp == nil {
		// Nothing changed (idempotent case)
		result := struct {
			Status    string
			Namespace string
		}{
			Status:    "unchanged",
			Namespace: c.Namespace,
		}
		return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
	}

	// Step 6: Handle async flag
	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(asyncOp, printer.StructuredOptions{})
	}

	// Step 7: Poll for completion
	return pollAsyncOperation(cctx, asyncOp.Id)
}

func (c *CloudNamespaceDeleteCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := newCloudClient(cctx)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))

	asyncOp, err := client.deleteNamespace(cctx.Context, deleteNamespaceParams{
		namespace:        c.Namespace,
		idempotent:       c.Idempotent,
		asyncOperationID: c.AsyncOperationId,
	})
	if err != nil {
		return err
	}
	// Handle async flag
	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(asyncOp, printer.StructuredOptions{})
	}

	// Poll for completion
	return pollAsyncOperation(cctx, asyncOp.Id)
}

func (c *CloudNamespaceListCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := newCloudClient(cctx)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))

	namespaces, nextPageToken, err := client.getNamespaces(cctx.Context, getNamespacesParams{
		pageSize:  int32(c.PageSize),
		pageToken: c.PageToken,
		name:      c.Name,
	})
	if err != nil {
		return err
	}

	return cctx.Printer.PrintStructured(
		struct {
			Namespaces    []*namespace.Namespace
			NextPageToken string
		}{
			Namespaces:    namespaces,
			NextPageToken: nextPageToken,
		},
		printer.StructuredOptions{},
	)
}

