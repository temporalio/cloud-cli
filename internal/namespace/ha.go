package namespace

import (
	"context"
	"maps"
	"slices"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
)

// RegionStatus pairs a region ID with its proto-generated status.
// AIDEV-NOTE: Region is the map key from Namespace.RegionStatus; it is not a field on the proto.
type RegionStatus struct {
	Region string
	Status namespacev1.NamespaceRegionStatus_State
}

// UpdateHAParams contains parameters for updating HA configuration on a namespace.
type UpdateHAParams struct {
	Namespace           string
	DisableAutoFailover bool
	ResourceVersion     string
	AsyncOperationID    string
}

// AddRegionParams contains parameters for adding a replica region to a namespace.
type AddRegionParams struct {
	Namespace        string
	Region           string
	ResourceVersion  string
	AsyncOperationID string
}

// RemoveRegionParams contains parameters for removing a replica region from a namespace.
type RemoveRegionParams struct {
	Namespace        string
	Region           string
	ResourceVersion  string
	AsyncOperationID string
}

// FailoverParams contains parameters for triggering a failover on a namespace.
type FailoverParams struct {
	Namespace        string
	Region           string
	AsyncOperationID string
}

// ListRegions returns per-region statuses for the specified namespace, sorted by region ID.
func (c *Client) ListRegions(ctx context.Context, namespaceName string) ([]RegionStatus, error) {
	ns, err := c.GetNamespace(ctx, namespaceName)
	if err != nil {
		return nil, err
	}

	m := ns.GetRegionStatus()
	keys := slices.Sorted(maps.Keys(m))
	result := make([]RegionStatus, len(keys))
	for i, k := range keys {
		result[i] = RegionStatus{Region: k, Status: m[k].GetState()}
	}

	return result, nil
}

// UpdateHA updates the HighAvailability spec on the namespace (fetch + mutate + UpdateNamespace).
func (c *Client) UpdateHA(ctx context.Context, params UpdateHAParams) (*operation.AsyncOperation, error) {
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	spec := ns.GetSpec()
	if spec.HighAvailability == nil {
		spec.HighAvailability = &namespacev1.HighAvailabilitySpec{}
	}
	spec.HighAvailability.DisableManagedFailover = params.DisableAutoFailover

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

// AddRegion adds a replica region to the specified namespace.
func (c *Client) AddRegion(ctx context.Context, params AddRegionParams) (*operation.AsyncOperation, error) {
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	resourceVersion := ns.ResourceVersion
	if params.ResourceVersion != "" {
		resourceVersion = params.ResourceVersion
	}

	res, err := c.Cloud.AddNamespaceRegion(ctx, &cloudservice.AddNamespaceRegionRequest{
		Namespace:        params.Namespace,
		Region:           params.Region,
		ResourceVersion:  resourceVersion,
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}
	return res.AsyncOperation, nil
}

// RemoveRegion removes a replica region from the specified namespace.
func (c *Client) RemoveRegion(ctx context.Context, params RemoveRegionParams) (*operation.AsyncOperation, error) {
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	resourceVersion := ns.ResourceVersion
	if params.ResourceVersion != "" {
		resourceVersion = params.ResourceVersion
	}

	res, err := c.Cloud.DeleteNamespaceRegion(ctx, &cloudservice.DeleteNamespaceRegionRequest{
		Namespace:        params.Namespace,
		Region:           params.Region,
		ResourceVersion:  resourceVersion,
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}
	return res.AsyncOperation, nil
}

// Failover triggers a failover for the specified namespace to the given region.
func (c *Client) Failover(ctx context.Context, params FailoverParams) (*operation.AsyncOperation, error) {
	res, err := c.Cloud.FailoverNamespaceRegion(ctx, &cloudservice.FailoverNamespaceRegionRequest{
		Namespace:        params.Namespace,
		Region:           params.Region,
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}
	return res.AsyncOperation, nil
}
