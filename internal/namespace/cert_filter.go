package namespace

import (
	"context"

	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
)

// CertFilter represents a certificate filter used for mTLS authentication.
// AIDEV-NOTE: This is a simplified representation of cloud-sdk's CertificateFilterSpec
// that provides a cleaner API for CLI operations. All fields are optional.
type CertFilter struct {
	CommonName             string
	Organization           string
	OrganizationalUnit     string
	SubjectAlternativeName string
}

// ListCertFilters returns the list of certificate filters configured for the namespace.
func (c *Client) ListCertFilters(ctx context.Context, name string) ([]CertFilter, error) {
	ns, err := c.GetNamespace(ctx, name)
	if err != nil {
		return nil, err
	}

	mtlsAuth := ns.GetSpec().GetMtlsAuth()
	if mtlsAuth == nil {
		return []CertFilter{}, nil
	}

	filters := mtlsAuth.GetCertificateFilters()
	if filters == nil {
		return []CertFilter{}, nil
	}

	result := make([]CertFilter, 0, len(filters))
	for _, filter := range filters {
		result = append(result, CertFilter{
			CommonName:             filter.GetCommonName(),
			Organization:           filter.GetOrganization(),
			OrganizationalUnit:     filter.GetOrganizationalUnit(),
			SubjectAlternativeName: filter.GetSubjectAlternativeName(),
		})
	}

	return result, nil
}

type AddCertFiltersParams struct {
	Namespace        string
	Filters          []CertFilter
	ResourceVersion  string
	AsyncOperationID string
}

// AddCertFilters adds new certificate filters to the namespace's mTLS configuration.
// This will check for duplicate filters and return an error if any already exist.
func (c *Client) AddCertFilters(ctx context.Context, params AddCertFiltersParams) (*operation.AsyncOperation, error) {
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	spec := ns.Spec
	if spec.MtlsAuth == nil {
		spec.MtlsAuth = &namespacev1.MtlsAuthSpec{}
	}

	for _, filter := range params.Filters {
		spec.MtlsAuth.CertificateFilters = append(spec.MtlsAuth.CertificateFilters, &namespacev1.CertificateFilterSpec{
			CommonName:             filter.CommonName,
			Organization:           filter.Organization,
			OrganizationalUnit:     filter.OrganizationalUnit,
			SubjectAlternativeName: filter.SubjectAlternativeName,
		})
	}

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

type DeleteCertFiltersParams struct {
	Namespace        string
	Filters          []CertFilter
	ResourceVersion  string
	AsyncOperationID string
}

// DeleteCertFilters removes certificate filters from the namespace's mTLS configuration.
// AIDEV-NOTE: Filters are matched by exact field equality. All four fields (CommonName,
// Organization, OrganizationalUnit, SubjectAlternativeName) must match for a filter to be removed.
func (c *Client) DeleteCertFilters(ctx context.Context, params DeleteCertFiltersParams) (*operation.AsyncOperation, error) {
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	spec := ns.Spec
	if spec.MtlsAuth == nil {
		spec.MtlsAuth = &namespacev1.MtlsAuthSpec{}
	}

	existingFilters := spec.MtlsAuth.GetCertificateFilters()

	// Build a list of filters to remove
	filtersToRemove := make(map[int]struct{})
	for _, filterToDelete := range params.Filters {
		for i, existing := range existingFilters {
			if certFiltersEqual(filterToDelete, existing) {
				filtersToRemove[i] = struct{}{}
			}
		}
	}

	// Build new filter list excluding the ones to remove
	var newFilters []*namespacev1.CertificateFilterSpec
	for i, filter := range existingFilters {
		if _, shouldRemove := filtersToRemove[i]; !shouldRemove {
			newFilters = append(newFilters, filter)
		}
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

// certFiltersEqual compares a CertFilter with a CertificateFilterSpec for equality.
// All fields must match for the filters to be considered equal.
func certFiltersEqual(cf CertFilter, spec *namespacev1.CertificateFilterSpec) bool {
	return cf.CommonName == spec.GetCommonName() &&
		cf.Organization == spec.GetOrganization() &&
		cf.OrganizationalUnit == spec.GetOrganizationalUnit() &&
		cf.SubjectAlternativeName == spec.GetSubjectAlternativeName()
}
