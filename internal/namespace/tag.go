package namespace

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
)

// Tag represents a namespace tag with its key and value.
type Tag struct {
	Key   string
	Value string
}

// ListTags returns the list of tags configured for the namespace, sorted by key.
// AIDEV-NOTE: Tags are key-value pairs used for organization and categorization
// of namespaces. They can be used for filtering and grouping namespaces. Tags are
// stored on the Namespace object itself, not in the NamespaceSpec.
func (c *Client) ListTags(ctx context.Context, name string) ([]Tag, error) {
	ns, err := c.GetNamespace(ctx, name)
	if err != nil {
		return nil, err
	}

	tags := ns.GetTags()
	if tags == nil {
		return []Tag{}, nil
	}

	result := make([]Tag, 0, len(tags))
	for key, value := range tags {
		result = append(result, Tag{
			Key:   key,
			Value: value,
		})
	}

	slices.SortFunc(result, func(a, b Tag) int {
		return cmp.Compare(a.Key, b.Key)
	})

	return result, nil
}

type SetTagsParams struct {
	Namespace        string
	Tags             []Tag
	AsyncOperationID string
}

// SetTags sets (upserts) tags for the namespace.
// AIDEV-NOTE: This uses the UpdateNamespaceTags API which supports upsert operations.
// If a tag with the same key already exists, its value will be updated. If it doesn't
// exist, a new tag will be created.
func (c *Client) SetTags(ctx context.Context, params SetTagsParams) (*operation.AsyncOperation, error) {
	if len(params.Tags) == 0 {
		return nil, fmt.Errorf("at least one tag must be specified")
	}

	tagsToUpsert := make(map[string]string)
	for _, tag := range params.Tags {
		tagsToUpsert[tag.Key] = tag.Value
	}

	res, err := c.Cloud.UpdateNamespaceTags(ctx, &cloudservice.UpdateNamespaceTagsRequest{
		Namespace:        params.Namespace,
		TagsToUpsert:     tagsToUpsert,
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}

	return res.AsyncOperation, nil
}

type DeleteTagsParams struct {
	Namespace        string
	Keys             []string
	AsyncOperationID string
}

// DeleteTags removes tags from the namespace by their keys.
// AIDEV-NOTE: This uses the UpdateNamespaceTags API with the tags_to_remove field.
// Tags that don't exist will be silently ignored (no error returned by the API).
func (c *Client) DeleteTags(ctx context.Context, params DeleteTagsParams) (*operation.AsyncOperation, error) {
	if len(params.Keys) == 0 {
		return nil, fmt.Errorf("at least one tag key must be specified")
	}

	res, err := c.Cloud.UpdateNamespaceTags(ctx, &cloudservice.UpdateNamespaceTagsRequest{
		Namespace:        params.Namespace,
		TagsToRemove:     params.Keys,
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}

	return res.AsyncOperation, nil
}
