package temporalcloudcli

import (
	"cmp"
	"errors"
	"slices"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

// Tag represents a namespace tag key-value pair for display.
type Tag struct {
	Key   string
	Value string
}

func (c *CloudNamespaceTagListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	tags := res.Namespace.GetTags()
	result := make([]Tag, 0, len(tags))
	for k, v := range tags {
		result = append(result, Tag{Key: k, Value: v})
	}
	slices.SortFunc(result, func(a, b Tag) int {
		return cmp.Compare(a.Key, b.Key)
	})
	return cctx.Printer.PrintResourceList(
		struct{ Tags []Tag }{Tags: result},
		printer.PrintResourceOptions{},
		printer.TableOptions{},
	)
}

func (c *CloudNamespaceTagCreateCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	yes, err := cctx.GetPrompter().PromptYes("Create")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
	}
	resp, err := client.UpdateNamespaceTags(cctx, &cloudservice.UpdateNamespaceTagsRequest{
		Namespace:        c.Namespace,
		TagsToUpsert:     map[string]string{c.Key: c.Value},
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNamespaceTagUpdateCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	resp, err := client.UpdateNamespaceTags(cctx, &cloudservice.UpdateNamespaceTagsRequest{
		Namespace:        c.Namespace,
		TagsToUpsert:     map[string]string{c.Key: c.Value},
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNamespaceTagDeleteCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	yes, err := cctx.GetPrompter().PromptYes("Delete")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting delete.")
	}
	resp, err := client.UpdateNamespaceTags(cctx, &cloudservice.UpdateNamespaceTagsRequest{
		Namespace:        c.Namespace,
		TagsToRemove:     []string{c.Key},
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleDeleteOperation(cctx, resp, err)
}
