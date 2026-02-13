package temporalcloudcli

import (
	namespace "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceLifecycleGetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))

	ns, err := client.getNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	// Extract lifecycle configuration, handle nil case
	enableDeleteProtection := false
	if ns.Spec.Lifecycle != nil {
		enableDeleteProtection = ns.Spec.Lifecycle.EnableDeleteProtection
	}

	// Create focused output showing only lifecycle information
	result := struct {
		Namespace              string `json:"namespace"`
		EnableDeleteProtection bool   `json:"enableDeleteProtection"`
	}{
		Namespace:              ns.Namespace,
		EnableDeleteProtection: enableDeleteProtection,
	}

	return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
}

func (c *CloudNamespaceLifecycleSetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))

	ns, err := client.getNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	// Clone the spec and update lifecycle
	newSpec := proto.Clone(ns.Spec).(*namespace.NamespaceSpec)

	// Ensure lifecycle is initialized
	if newSpec.Lifecycle == nil {
		newSpec.Lifecycle = &namespace.LifecycleSpec{}
	}
	newSpec.Lifecycle.EnableDeleteProtection = c.EnableDeleteProtection

	// Show diff
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
	return pollAsyncOperation(cctx, cloudClient, res.asyncOp.Id, res.Namespace)
}
