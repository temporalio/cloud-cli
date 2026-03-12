package temporalcloudcli

import (
	"context"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type (
	GetLifecycleParams struct {
		Namespace string

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
	}

	SetLifecycleParams struct {
		Namespace              string
		EnableDeleteProtection bool
		ResourceVersion        string
		AsyncOperationID       string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}
)

func GetLifecycle(ctx context.Context, params GetLifecycleParams) error {
	res, err := params.Cloud.GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: params.Namespace})
	if err != nil {
		return err
	}
	ns := res.Namespace
	enableDeleteProtection := false
	if ns.Spec.Lifecycle != nil {
		enableDeleteProtection = ns.Spec.Lifecycle.EnableDeleteProtection
	}
	return params.Printer.PrintStructured(struct {
		Namespace              string `json:"namespace"`
		EnableDeleteProtection bool   `json:"enableDeleteProtection"`
	}{
		Namespace:              ns.Namespace,
		EnableDeleteProtection: enableDeleteProtection,
	}, printer.StructuredOptions{})
}

func SetLifecycle(ctx context.Context, params SetLifecycleParams) error {
	res, err := params.Cloud.GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: params.Namespace})
	if err != nil {
		return err
	}
	ns := res.Namespace
	newSpec := proto.Clone(ns.Spec).(*namespacev1.NamespaceSpec)
	if newSpec.Lifecycle == nil {
		newSpec.Lifecycle = &namespacev1.LifecycleSpec{}
	}
	newSpec.Lifecycle.EnableDeleteProtection = params.EnableDeleteProtection

	if err := params.Prompter.PromptApply(ns.Spec, newSpec, false); err != nil {
		return err
	}

	rv := ns.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}
	updateNamespace := runAsyncOperation(params.Cloud.UpdateNamespace, params.OperationHandler)
	return updateNamespace(ctx, &cloudservice.UpdateNamespaceRequest{
		Namespace:        params.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func (c *CloudNamespaceLifecycleGetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return GetLifecycle(cctx.Context, GetLifecycleParams{
		Namespace: c.Namespace,
		Cloud:     cloudClient.CloudService(),
		Printer:   cctx.Printer,
	})
}

func (c *CloudNamespaceLifecycleSetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return SetLifecycle(cctx.Context, SetLifecycleParams{
		Namespace:              c.Namespace,
		EnableDeleteProtection: c.EnableDeleteProtection,
		ResourceVersion:        c.ResourceVersion,
		AsyncOperationID:       c.AsyncOperationId,
		Cloud:                  cloudClient.CloudService(),
		Prompter:               newPrompter(cctx),
		OperationHandler:       NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions),
	})
}
