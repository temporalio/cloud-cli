package temporalcloudcli

import (
	"errors"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceTagListCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	tags, err := namespaceClient.ListTags(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	return cctx.Printer.PrintResourceList(
		struct{ Tags []namespace.Tag }{Tags: tags},
		printer.PrintResourceOptions{},
		printer.TableOptions{},
	)
}

func (c *CloudNamespaceTagCreateCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	yes, err := cctx.promptYes("Create (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return errors.New("Aborting create.")
	}

	setTag := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, namespaceClient.SetTag)
	return setTag(namespace.SetTagParams{
		Namespace:        c.Namespace,
		Key:              c.Key,
		Value:            c.Value,
		AsyncOperationID: c.AsyncOperationId,
	})
}

func (c *CloudNamespaceTagUpdateCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	setTag := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, namespaceClient.SetTag)
	return setTag(namespace.SetTagParams{
		Namespace:        c.Namespace,
		Key:              c.Key,
		Value:            c.Value,
		AsyncOperationID: c.AsyncOperationId,
	})
}

func (c *CloudNamespaceTagDeleteCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	yes, err := cctx.promptYes("Delete (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return errors.New("Aborting delete.")
	}

	deleteTags := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, namespaceClient.DeleteTags)
	return deleteTags(namespace.DeleteTagsParams{
		Namespace:        c.Namespace,
		Keys:             []string{c.Key},
		AsyncOperationID: c.AsyncOperationId,
	})
}
