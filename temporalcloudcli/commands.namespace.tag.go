package temporalcloudcli

import (
	"cmp"
	"errors"
	"slices"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

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

	tags := make([]Tag, 0, len(res.Namespace.GetTags()))
	for key, value := range res.Namespace.GetTags() {
		tags = append(tags, Tag{Key: key, Value: value})
	}
	slices.SortFunc(tags, func(a, b Tag) int {
		return cmp.Compare(a.Key, b.Key)
	})

	return cctx.Printer.PrintResourceList(
		struct{ Tags []Tag }{Tags: tags},
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
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudNamespaceTagUpdateCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	existingTags := res.Namespace.GetTags()
	newTags := make(map[string]string, len(existingTags)+1)
	for k, v := range existingTags {
		newTags[k] = v
	}
	newTags[c.Key] = c.Value

	yes, err := cctx.GetPrompter().PromptApply(
		&namespacev1.Namespace{Tags: existingTags},
		&namespacev1.Namespace{Tags: newTags},
		false,
	)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting update.")
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

	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	existingTags := res.Namespace.GetTags()
	newTags := make(map[string]string, len(existingTags))
	for k, v := range existingTags {
		if k != c.Key {
			newTags[k] = v
		}
	}

	yes, err := cctx.GetPrompter().PromptApply(
		&namespacev1.Namespace{Tags: existingTags},
		&namespacev1.Namespace{Tags: newTags},
		false,
	)
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
