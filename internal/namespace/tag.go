package namespace

import (
	"cmp"
	"context"
	"slices"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
)

// Tag represents a namespace tag with its key and value.
// AIDEV-NOTE: Tags are stored on the Namespace object directly (not in NamespaceSpec).
// They are managed via the UpdateNamespaceTags API (separate from UpdateNamespace).
type Tag struct {
	Key   string
	Value string
}

// ListTags returns the list of tags configured for the namespace, sorted by key.
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
		result = append(result, Tag{Key: key, Value: value})
	}

	slices.SortFunc(result, func(a, b Tag) int {
		return cmp.Compare(a.Key, b.Key)
	})

	return result, nil
}

// SetTagParams contains parameters for setting (creating or updating) a namespace tag.
type SetTagParams struct {
	Namespace        string
	Key              string
	Value            string
	AsyncOperationID string
}

// SetTag creates or updates a tag on the namespace.
func (c *Client) SetTag(ctx context.Context, params SetTagParams) (*operation.AsyncOperation, error) {
	res, err := c.Cloud.UpdateNamespaceTags(ctx, &cloudservice.UpdateNamespaceTagsRequest{
		Namespace:        params.Namespace,
		TagsToUpsert:     map[string]string{params.Key: params.Value},
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}

	return res.AsyncOperation, nil
}

// DeleteTagsParams contains parameters for deleting namespace tags.
type DeleteTagsParams struct {
	Namespace        string
	Keys             []string
	AsyncOperationID string
}

// DeleteTags removes tags from the namespace by their keys. Keys that do not
// exist are silently ignored by the API.
func (c *Client) DeleteTags(ctx context.Context, params DeleteTagsParams) (*operation.AsyncOperation, error) {
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
