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

	asyncOp, err := client.updateNamespace(cctx.Context, newSpec, updateNamespaceParams{
		asyncOperationID: c.AsyncOperationId,
		resourceVersion:  ns.ResourceVersion,
		namespace:        c.Namespace,
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

	// Step 4: Apply the namespace (create or update)
	params := applyNamespaceParams{
		asyncOperationID: c.AsyncOperationId, // Use the flag value if provided
		idempotent:       c.Idempotent,       // Use the flag value
	}

	asyncOp, err := client.applyNamespace(cctx.Context, c.Namespace, spec, params)
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
			Namespace: spec.Name,
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

func (c *CloudNamespaceDiffCommand) run(cctx *CommandContext, _ []string) error {
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

	// Step 5: Retrieve existing namespace
	existing, err := client.getNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}
	// Step 6: Compute and print diff
	return cctx.Printer.PrintDiff(existing.Spec, spec, printer.DiffOptions{
		Verbose: c.Verbose,
	})
}
