package temporalcloudcli

import (
	"fmt"

	"go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	n, err := client.GetNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	if c.Spec {
		return cctx.Printer.PrintStructured(n.Spec, printer.StructuredOptions{})
	}
	return cctx.Printer.PrintResource(n, printer.PrintResourceOptions{})
}

func (c *CloudNamespaceEditCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	ns, err := namespaceClient.GetNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	newSpec := &namespacev1.NamespaceSpec{}
	err = runEditorForJSONEditForProtos(ns.Spec, newSpec)
	if err != nil {
		return err
	}

	err = promptApplyResource(cctx, ns.Spec, newSpec, c.VerboseDiff)
	if err != nil {
		return err
	}

	updateNamespace := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, namespaceClient.UpdateNamespace)
	return updateNamespace(namespace.UpdateNamespaceParams{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		AsyncOperationID: c.AsyncOperationId,
		ResourceVersion:  c.ResourceVersion,
	})
}

func (c *CloudNamespaceApplyCommand) run(cctx *CommandContext, _ []string) error {
	specData, err := loadJSONSpec(c.Spec)
	if err != nil {
		return err
	}

	spec := &namespacev1.NamespaceSpec{}
	if err := cctx.UnmarshalProtoJSON(specData, spec); err != nil {
		return fmt.Errorf("failed to parse JSON spec: %w", err)
	}

	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	listRes, _, err := namespaceClient.ListNamespaces(cctx.Context, spec.Name, "", 0)
	if err != nil {
		return err
	}

	var existing *namespacev1.Namespace
	if len(listRes) > 0 {
		if len(listRes) > 1 {
			return fmt.Errorf("expected one namespace to exist, got %d", len(listRes))
		}

		existing, err = namespaceClient.GetNamespace(cctx.Context, listRes[0].Namespace)
		if err != nil {
			if !isNotFoundErr(err) {
				return err
			}
		}
	}

	var existingSpec *namespacev1.NamespaceSpec
	if existing != nil {
		existingSpec = existing.Spec
	}

	err = promptApplyResource(cctx, existingSpec, spec, c.VerboseDiff)
	if err != nil {
		return err
	}

	params := namespace.UpsertNamespaceParams{
		Spec:             spec,
		AsyncOperationID: c.AsyncOperationId,
		ResourceVersion:  c.ResourceVersion,
	}
	if existing != nil {
		params.Namespace = existing.Namespace
	}
	upsertNamespace := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, namespaceClient.UpsertNamespace)
	return upsertNamespace(params)
}

func (c *CloudNamespaceDeleteCommand) run(cctx *CommandContext, _ []string) error {
	client, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	yes, err := cctx.promptYes("Delete (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return fmt.Errorf("Aborting delete.")
	}

	deleteNamespace := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, client.DeleteNamespace)
	return deleteNamespace(namespace.DeleteNamespaceParams{
		Namespace:        c.Namespace,
		AsyncOperationID: c.AsyncOperationId,
		ResourceVersion:  c.ResourceVersion,
	})
}

func (c *CloudNamespaceListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	namespaces, nextPageToken, err := client.ListNamespaces(cctx.Context, c.Name, c.PageToken, int32(c.PageSize))
	if err != nil {
		return err
	}

	return cctx.Printer.PrintResourceList(
		struct {
			Namespaces    []*namespacev1.Namespace
			NextPageToken string
		}{
			Namespaces:    namespaces,
			NextPageToken: nextPageToken,
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Namespace", "State", "CreatedTime"},
			SpecFields: []string{"Regions"},
		},
		printer.TableOptions{},
	)
}

func getNamespaceClient(cctx *CommandContext, opts ClientOptions) (NamespaceClient, error) {
	if cctx.NamespaceClient != nil {
		return cctx.NamespaceClient, nil
	}

	cloudClient, err := getCloudClient(cctx, opts)
	if err != nil {
		return nil, err
	}

	cctx.NamespaceClient = namespace.NewClient(cloudClient)
	return cctx.NamespaceClient, nil
}

func getCloudClient(cctx *CommandContext, opts ClientOptions) (cloudservice.CloudServiceClient, error) {
	if cctx.CloudClient != nil {
		return cctx.CloudClient, nil
	}

	cloudClient, err := cctx.BuildCloudClient(opts)
	if err != nil {
		return nil, err
	}

	cctx.CloudClient = cloudClient.CloudService()
	return cctx.CloudClient, nil
}
