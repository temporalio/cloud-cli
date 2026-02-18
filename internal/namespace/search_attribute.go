package namespace

import (
	"context"
	"fmt"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
)

// SearchAttribute represents a custom search attribute with its name and type.
type SearchAttribute struct {
	Name string
	Type namespacev1.NamespaceSpec_SearchAttributeType
}

// ListSearchAttributes returns the list of search attributes configured for the namespace.
// AIDEV-NOTE: This returns custom search attributes from the namespace spec. System search
// attributes are not included in this list.
func (c *Client) ListSearchAttributes(ctx context.Context, name string) ([]SearchAttribute, error) {
	ns, err := c.GetNamespace(ctx, name)
	if err != nil {
		return nil, err
	}

	searchAttrs := ns.GetSpec().GetSearchAttributes()
	if searchAttrs == nil {
		return []SearchAttribute{}, nil
	}

	result := make([]SearchAttribute, 0, len(searchAttrs))
	for name, attrType := range searchAttrs {
		result = append(result, SearchAttribute{
			Name: name,
			Type: attrType,
		})
	}

	return result, nil
}

type CreateSearchAttributeParams struct {
	Namespace        string
	Name             string
	Type             namespacev1.NamespaceSpec_SearchAttributeType
	ResourceVersion  string
	AsyncOperationID string
}

// CreateSearchAttribute adds a new search attribute to the namespace.
// AIDEV-NOTE: This will return an error if a search attribute with the same name already exists.
// Search attributes cannot be deleted once created, only renamed.
func (c *Client) CreateSearchAttribute(ctx context.Context, params CreateSearchAttributeParams) (*operation.AsyncOperation, error) {
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	spec := ns.Spec
	if spec.SearchAttributes == nil {
		spec.SearchAttributes = make(map[string]namespacev1.NamespaceSpec_SearchAttributeType)
	}

	// Check if search attribute already exists
	if _, exists := spec.SearchAttributes[params.Name]; exists {
		return nil, fmt.Errorf("search attribute %q already exists", params.Name)
	}

	spec.SearchAttributes[params.Name] = params.Type

	resourceVersion := ns.ResourceVersion
	if params.ResourceVersion != "" {
		resourceVersion = params.ResourceVersion
	}

	updateParams := UpdateNamespaceParams{
		Namespace:        params.Namespace,
		Spec:             spec,
		ResourceVersion:  resourceVersion,
		AsyncOperationID: params.AsyncOperationID,
	}
	return c.UpdateNamespace(ctx, updateParams)
}

type RenameSearchAttributeParams struct {
	Namespace                         string
	ExistingCustomSearchAttributeName string
	NewCustomSearchAttributeName      string
	ResourceVersion                   string
	AsyncOperationID                  string
}

// RenameSearchAttribute renames an existing custom search attribute.
// AIDEV-NOTE: This uses a dedicated RPC endpoint for renaming search attributes rather than
// updating the namespace spec directly. This is because renaming a search attribute requires
// special handling on the backend to ensure data consistency.
func (c *Client) RenameSearchAttribute(ctx context.Context, params RenameSearchAttributeParams) (*operation.AsyncOperation, error) {
	// First verify the namespace exists and get its current version
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	// Verify the existing search attribute exists
	searchAttrs := ns.GetSpec().GetSearchAttributes()
	if _, exists := searchAttrs[params.ExistingCustomSearchAttributeName]; !exists {
		return nil, fmt.Errorf("search attribute %q does not exist", params.ExistingCustomSearchAttributeName)
	}

	// Verify the new name doesn't already exist
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

type DeleteSearchAttributeParams struct {
	Namespace        string
	Name             string
	ResourceVersion  string
	AsyncOperationID string
}

// DeleteSearchAttribute removes a search attribute from the namespace.
// AIDEV-NOTE: This removes the search attribute from the namespace spec. Note that the data
// associated with this search attribute may still exist in the backend. This operation should
// be used with caution.
func (c *Client) DeleteSearchAttribute(ctx context.Context, params DeleteSearchAttributeParams) (*operation.AsyncOperation, error) {
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	spec := ns.Spec
	if spec.SearchAttributes == nil {
		return nil, fmt.Errorf("no search attributes configured for namespace")
	}

	// Verify the search attribute exists
	if _, exists := spec.SearchAttributes[params.Name]; !exists {
		return nil, fmt.Errorf("search attribute %q does not exist", params.Name)
	}

	// Remove the search attribute
	delete(spec.SearchAttributes, params.Name)

	resourceVersion := ns.ResourceVersion
	if params.ResourceVersion != "" {
		resourceVersion = params.ResourceVersion
	}

	updateParams := UpdateNamespaceParams{
		Namespace:        params.Namespace,
		Spec:             spec,
		ResourceVersion:  resourceVersion,
		AsyncOperationID: params.AsyncOperationID,
	}
	return c.UpdateNamespace(ctx, updateParams)
}
