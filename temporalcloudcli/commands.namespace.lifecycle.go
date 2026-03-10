package temporalcloudcli

import (
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceLifecycleGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	ns, err := client.GetNamespace(cctx.Context, c.Namespace)
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

	return cctx.Printer.PrintResource(result, printer.PrintResourceOptions{})
}

func (c *CloudNamespaceLifecycleSetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	ns, err := client.GetNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	// Clone the spec and update lifecycle
	newSpec := proto.Clone(ns.Spec).(*namespacev1.NamespaceSpec)

	// Ensure lifecycle is initialized
	if newSpec.Lifecycle == nil {
		newSpec.Lifecycle = &namespacev1.LifecycleSpec{}
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

	updateNamespace := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, client.UpdateNamespace)
	return updateNamespace(namespace.UpdateNamespaceParams{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		AsyncOperationID: c.AsyncOperationId,
		ResourceVersion:  resourceVersion,
	})
}
