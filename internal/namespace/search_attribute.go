package namespace

import (
	"context"
	"fmt"
	"sort"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
)

// SearchAttribute represents a custom search attribute with its name and type.
type SearchAttribute struct {
	Name string
	Type namespacev1.NamespaceSpec_SearchAttributeType
}

// ListSearchAttributes retrieves all custom search attributes configured for the namespace.
func (c *Client) ListSearchAttributes(ctx context.Context, name string) ([]SearchAttribute, error) {
	ns, err := c.GetNamespace(ctx, name)
	if err != nil {
		return nil, err
	}

	searchAttrs := ns.GetSpec().GetSearchAttributes()
	if searchAttrs == nil {
		return []SearchAttribute{}, nil
	}

	keys := make([]string, 0, len(searchAttrs))
	for attrName := range searchAttrs {
		keys = append(keys, attrName)
	}
	sort.Strings(keys)

	result := make([]SearchAttribute, len(keys))
	for i, attrName := range keys {
		result[i] = SearchAttribute{
			Name: attrName,
			Type: searchAttrs[attrName],
		}
	}

	return result, nil
}

// CreateSearchAttributeParams contains parameters for adding a search attribute to a namespace.
type CreateSearchAttributeParams struct {
	Namespace        string
	Name             string
	Type             namespacev1.NamespaceSpec_SearchAttributeType
	ResourceVersion  string
	AsyncOperationID string
}

// CreateSearchAttribute adds a new custom search attribute to the namespace spec.
// Returns an error if a search attribute with the same name already exists.
func (c *Client) CreateSearchAttribute(ctx context.Context, params CreateSearchAttributeParams) (*operation.AsyncOperation, error) {
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	spec := ns.GetSpec()
	if spec.SearchAttributes == nil {
		spec.SearchAttributes = make(map[string]namespacev1.NamespaceSpec_SearchAttributeType)
	}

	if _, exists := spec.SearchAttributes[params.Name]; exists {
		return nil, fmt.Errorf("search attribute %q already exists", params.Name)
	}

	spec.SearchAttributes[params.Name] = params.Type

	resourceVersion := ns.ResourceVersion
	if params.ResourceVersion != "" {
		resourceVersion = params.ResourceVersion
	}

	return c.UpdateNamespace(ctx, UpdateNamespaceParams{
		Namespace:        params.Namespace,
		Spec:             spec,
		ResourceVersion:  resourceVersion,
		AsyncOperationID: params.AsyncOperationID,
	})
}

// RenameSearchAttributeParams contains parameters for renaming a custom search attribute.
type RenameSearchAttributeParams struct {
	Namespace                         string
	ExistingCustomSearchAttributeName string
	NewCustomSearchAttributeName      string
	ResourceVersion                   string
	AsyncOperationID                  string
}

// RenameSearchAttribute renames an existing custom search attribute.
// AIDEV-NOTE: Uses the dedicated RenameCustomSearchAttribute RPC rather than UpdateNamespace
// because the backend must migrate existing workflow data to the new attribute name.
func (c *Client) RenameSearchAttribute(ctx context.Context, params RenameSearchAttributeParams) (*operation.AsyncOperation, error) {
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	searchAttrs := ns.GetSpec().GetSearchAttributes()
	if _, exists := searchAttrs[params.ExistingCustomSearchAttributeName]; !exists {
		return nil, fmt.Errorf("search attribute %q does not exist", params.ExistingCustomSearchAttributeName)
	}

	if _, exists := searchAttrs[params.NewCustomSearchAttributeName]; exists {
		return nil, fmt.Errorf("search attribute %q already exists", params.NewCustomSearchAttributeName)
	}

	resourceVersion := ns.ResourceVersion
	if params.ResourceVersion != "" {
		resourceVersion = params.ResourceVersion
	}

	res, err := c.Cloud.RenameCustomSearchAttribute(ctx, &cloudservice.RenameCustomSearchAttributeRequest{
		Namespace:                         params.Namespace,
		ExistingCustomSearchAttributeName: params.ExistingCustomSearchAttributeName,
		NewCustomSearchAttributeName:      params.NewCustomSearchAttributeName,
		ResourceVersion:                   resourceVersion,
		AsyncOperationId:                  params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}

	return res.AsyncOperation, nil
}
