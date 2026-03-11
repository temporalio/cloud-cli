package temporalcloudcli

import (
	"context"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type (
	GetRetentionParams struct {
		Namespace string

		Cloud   namespace.CloudService
		Printer *printer.Printer
	}

	SetRetentionParams struct {
		Namespace        string
		RetentionDays    int32
		ResourceVersion  string
		AsyncOperationID string

		Cloud            namespace.CloudService
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}
)

func GetRetention(ctx context.Context, params GetRetentionParams) error {
	res, err := params.Cloud.GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: params.Namespace})
	if err != nil {
		return err
	}
	return params.Printer.PrintStructured(struct {
		Namespace     string `json:"namespace"`
		RetentionDays int32  `json:"retentionDays"`
	}{
		Namespace:     res.Namespace.Namespace,
		RetentionDays: res.Namespace.Spec.RetentionDays,
	}, printer.StructuredOptions{})
}

func SetRetention(ctx context.Context, params SetRetentionParams) error {
	res, err := params.Cloud.GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: params.Namespace})
	if err != nil {
		return err
	}
	ns := res.Namespace
	newSpec := proto.Clone(ns.Spec).(*namespacev1.NamespaceSpec)
	newSpec.RetentionDays = params.RetentionDays

	if err := params.Prompter.PromptApply(ns.Spec, newSpec, false); err != nil {
		return err
	}

	rv := ns.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}
	updateNamespace := runAsyncOperation(params.Cloud.UpdateNamespace, params.OperationHandler)
	updateParams := &cloudservice.UpdateNamespaceRequest{
		Namespace:        params.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	}
	return updateNamespace(ctx, updateParams)
}

func (c *CloudNamespaceRetentionGetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return GetRetention(cctx.Context, GetRetentionParams{
		Namespace: c.Namespace,
		Cloud:     cloudClient.CloudService(),
		Printer:   cctx.Printer,
	})
}

func (c *CloudNamespaceRetentionSetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return SetRetention(cctx.Context, SetRetentionParams{
		Namespace:        c.Namespace,
		RetentionDays:    int32(c.RetentionDays),
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewAsyncOperationHandler(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions),
	})
}
