package temporalcloudcli

import (
	"fmt"

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

	return cctx.Printer.PrintStructured(tags, printer.StructuredOptions{})
}

func (c *CloudNamespaceTagCreateCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	if c.Key == "" {
		return fmt.Errorf("key is required")
	}

	if c.Value == "" {
		return fmt.Errorf("value is required")
	}

	// Check if the tag already exists
	existingTags, err := namespaceClient.ListTags(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	for _, tag := range existingTags {
		if tag.Key == c.Key {
			return fmt.Errorf("tag with key %q already exists", c.Key)
		}
	}

	yes, err := cctx.promptYes("Create tag (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return fmt.Errorf("Aborting create.")
	}

	params := namespace.SetTagsParams{
		Namespace: c.Namespace,
		Tags: []namespace.Tag{
			{
				Key:   c.Key,
				Value: c.Value,
			},
		},
		AsyncOperationID: c.AsyncOperationId,
	}
	op, err := namespaceClient.SetTags(cctx.Context, params)
	if err != nil {
		if isNothingChangedErr(c.Idempotent, err) {
			result := struct {
				Status    string
				Namespace string
			}{
				Status:    "unchanged",
				Namespace: c.Namespace,
			}
			return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
		}

		return err
	}

	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: op,
			ID:      c.Namespace,
		}, printer.StructuredOptions{})
	}

	poller, err := getPoller(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	return poller.Poll(cctx.Context, op.Id, c.Namespace)
}

func (c *CloudNamespaceTagUpdateCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	if c.Key == "" {
		return fmt.Errorf("key is required")
	}

	if c.Value == "" {
		return fmt.Errorf("value is required")
	}

	// Check if the tag exists
	existingTags, err := namespaceClient.ListTags(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	tagExists := false
	for _, tag := range existingTags {
		if tag.Key == c.Key {
			tagExists = true
			break
		}
	}

	if !tagExists {
		return fmt.Errorf("tag with key %q does not exist", c.Key)
	}

	yes, err := cctx.promptYes("Update tag (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return fmt.Errorf("Aborting update.")
	}

	params := namespace.SetTagsParams{
		Namespace: c.Namespace,
		Tags: []namespace.Tag{
			{
				Key:   c.Key,
				Value: c.Value,
			},
		},
		AsyncOperationID: c.AsyncOperationId,
	}
	op, err := namespaceClient.SetTags(cctx.Context, params)
	if err != nil {
		if isNothingChangedErr(c.Idempotent, err) {
			result := struct {
				Status    string
				Namespace string
			}{
				Status:    "unchanged",
				Namespace: c.Namespace,
			}
			return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
		}

		return err
	}

	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: op,
			ID:      c.Namespace,
		}, printer.StructuredOptions{})
	}

	poller, err := getPoller(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	return poller.Poll(cctx.Context, op.Id, c.Namespace)
}

func (c *CloudNamespaceTagDeleteCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	if c.Key == "" {
		return fmt.Errorf("key is required")
	}

	yes, err := cctx.promptYes("Delete tag (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return fmt.Errorf("Aborting delete.")
	}

	params := namespace.DeleteTagsParams{
		Namespace:        c.Namespace,
		Keys:             []string{c.Key},
		AsyncOperationID: c.AsyncOperationId,
	}
	op, err := namespaceClient.DeleteTags(cctx.Context, params)
	if err != nil {
		if isNothingChangedErr(c.Idempotent, err) {
			result := struct {
				Status    string
				Namespace string
			}{
				Status:    "unchanged",
				Namespace: c.Namespace,
			}
			return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
		}

		return err
	}

	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: op,
			ID:      c.Namespace,
		}, printer.StructuredOptions{})
	}

	poller, err := getPoller(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	return poller.Poll(cctx.Context, op.Id, c.Namespace)
}
