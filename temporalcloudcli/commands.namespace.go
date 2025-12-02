package temporalcloudcli

import (
	"fmt"

	namespace "go.temporal.io/cloud-sdk/api/namespace/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceGetCommand) run(cctx *CommandContext, args []string) error {
	cloudClient, err := newCloudClient(cctx)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))

	n, err := client.getNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	return cctx.Printer.PrintStructured(n, printer.StructuredOptions{})
}

func (c *CloudNamespaceEditCommand) run(cctx *CommandContext, args []string) error {
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

	asyncOp, err := client.updateNamespace(cctx.Context, newSpec, updateNamespaceParams{
		asyncOperationID: c.AsyncOperationId,
		resourceVersion:  ns.ResourceVersion,
		namespace:        c.Namespace,
		idempotent:       c.Idemptotent,
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

func (c *CloudNamespaceApplyCommand) run(cctx *CommandContext, args []string) error {
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

	// Validate namespace name is present
	if spec.Name == "" {
		return fmt.Errorf("namespace name must be provided either via --namespace flag or in the spec")
	}

	// Step 4: Create cloud and namespace clients
	cloudClient, err := newCloudClient(cctx)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))

	// Step 5: Handle dry-run mode
	if c.DryRun {
		return c.performDryRun(cctx, client, spec)
	}

	// Step 6: Apply the namespace (create or update)
	params := applyNamespaceParams{
		asyncOperationID: c.AsyncOperationId, // Use the flag value if provided
		idempotent:       c.Idemptotent,      // Use the flag value
	}

	asyncOp, err := client.applyNamespace(cctx.Context, spec, params)
	if err != nil {
		return fmt.Errorf("failed to apply namespace: %w", err)
	}

	// Step 7: Handle result
	if asyncOp == nil {
		// Nothing changed (idempotent case)
		result := struct {
			Status    string
			Namespace string
		}{
			Status:    "unchanged",
			Namespace: spec.Name,
		}
		return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
	}

	// Step 8: Handle async flag
	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(asyncOp, printer.StructuredOptions{})
	}

	// Step 9: Poll for completion
	return pollAsyncOperation(cctx, asyncOp.Id)
}

func (c *CloudNamespaceApplyCommand) performDryRun(cctx *CommandContext, client *namespaceClient, spec *namespace.NamespaceSpec) error {
	// Try to get existing namespace
	namespaces, err := client.listNamespacesWithName(cctx.Context, spec.Name)
	if err != nil {
		return err
	} else if len(namespaces) > 1 {
		return fmt.Errorf("multiple namespaces match namespace name: %s", spec.GetName())
	} else if len(namespaces) == 0 {
		// Namespace doesn't exist - would create
		result := struct {
			DryRun    bool
			Action    string
			Namespace string
			Spec      *namespace.NamespaceSpec
		}{
			DryRun:    true,
			Action:    "create",
			Namespace: spec.Name,
			Spec:      spec,
		}
		return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
	}

	existing := namespaces[0]

	// Namespace exists - would update
	result := struct {
		DryRun          bool
		Action          string
		Namespace       string
		ResourceVersion string
		Spec            *namespace.NamespaceSpec
	}{
		DryRun:          true,
		Action:          "update",
		Namespace:       spec.Name,
		ResourceVersion: existing.ResourceVersion,
		Spec:            spec,
	}
	return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
}
