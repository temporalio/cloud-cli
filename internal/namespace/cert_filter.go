package namespace

import (
	"context"
	"slices"

	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/protobuf/proto"
)

// ListCertFilters retrieves all certificate filters configured for mTLS authentication
// on the specified namespace.
func (c *Client) ListCertFilters(ctx context.Context, name string) ([]*namespacev1.CertificateFilterSpec, error) {
	ns, err := c.GetNamespace(ctx, name)
	if err != nil {
		return nil, err
	}

	filters := ns.GetSpec().GetMtlsAuth().GetCertificateFilters()
	if filters == nil {
		return []*namespacev1.CertificateFilterSpec{}, nil
	}

	return filters, nil
}

// AddCertFiltersParams contains parameters for adding certificate filters to a namespace.
type AddCertFiltersParams struct {
	Namespace        string
	Filters          []*namespacev1.CertificateFilterSpec
	ResourceVersion  string
	AsyncOperationID string
}

// AddCertFilters adds certificate filters to the namespace's mTLS authentication configuration.
// The server will handle duplicate detection and return an appropriate error if needed.
func (c *Client) AddCertFilters(ctx context.Context, params AddCertFiltersParams) (*operation.AsyncOperation, error) {
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	spec := ns.GetSpec()
	// Ensure MtlsAuth is initialized
	if spec.MtlsAuth == nil {
		spec.MtlsAuth = &namespacev1.MtlsAuthSpec{Enabled: true}
	}

	existingFilters := spec.MtlsAuth.GetCertificateFilters()

	// Append new filters to existing ones
	newFilters := append(existingFilters, params.Filters...)
	spec.MtlsAuth.CertificateFilters = newFilters

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

// DeleteCertFiltersParams contains parameters for deleting certificate filters from a namespace.
type DeleteCertFiltersParams struct {
	Namespace        string
	Filters          []*namespacev1.CertificateFilterSpec
	ResourceVersion  string
	AsyncOperationID string
}

// DeleteCertFilters removes certificate filters from the namespace's mTLS authentication configuration.
// Filters are matched by exact field equality.
func (c *Client) DeleteCertFilters(ctx context.Context, params DeleteCertFiltersParams) (*operation.AsyncOperation, error) {
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	spec := ns.GetSpec()
	existingFilters := spec.GetMtlsAuth().GetCertificateFilters()

	// Build new filter list excluding filters that match ones being deleted
	var newFilters []*namespacev1.CertificateFilterSpec
	for _, existing := range existingFilters {
		shouldRemove := slices.ContainsFunc(params.Filters, func(toRemove *namespacev1.CertificateFilterSpec) bool {
			return proto.Equal(existing, toRemove)
		})
		if shouldRemove {
			continue
		}

		newFilters = append(newFilters, existing)
	}

	// Update the spec with the new filter list
	if spec.MtlsAuth == nil {
		spec.MtlsAuth = &namespacev1.MtlsAuthSpec{Enabled: true}
	}
	spec.MtlsAuth.CertificateFilters = newFilters

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
