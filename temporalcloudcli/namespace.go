package temporalcloudcli

import (
	"context"
	"fmt"

	"go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespace "go.temporal.io/cloud-sdk/api/namespace/v1"
	"go.temporal.io/cloud-sdk/api/operation/v1"
	"go.temporal.io/cloud-sdk/cloudclient"
)

type namespaceClient struct {
	client *cloudclient.Client
}

type namespaceOpt func(*namespaceClient)

func withCloudClient(cloudClient *cloudclient.Client) namespaceOpt {
	return func(nc *namespaceClient) {
		nc.client = cloudClient
	}
}

func newNamespaceClient(opts ...namespaceOpt) *namespaceClient {
	namespaceClient := &namespaceClient{}

	for _, opt := range opts {
		opt(namespaceClient)
	}

	return namespaceClient
}

func (c *namespaceClient) getNamespace(ctx context.Context, namespace string) (*namespace.Namespace, error) {
	res, err := c.client.CloudService().GetNamespace(ctx, &cloudservice.GetNamespaceRequest{
		Namespace: namespace,
	})
	if err != nil {
		return nil, err
	}

	if res.Namespace == nil || res.Namespace.Namespace == "" {
		// this should never happen, the server should return an error when the namespace is not found
		return nil, fmt.Errorf("invalid namespace returned by server")
	}
	return res.Namespace, nil
}

type getNamespacesParams struct {
	pageSize  int32
	pageToken string
	name      string // optional, if set, will filter by name
}

func (c *namespaceClient) getNamespaces(ctx context.Context, params getNamespacesParams) ([]*namespace.Namespace, string, error) {
	res, err := c.client.CloudService().GetNamespaces(ctx, &cloudservice.GetNamespacesRequest{
		PageSize:  params.pageSize,
		PageToken: params.pageToken,
		Name:      params.name,
	})
	if err != nil {
		return nil, "", err
	}

	return res.Namespaces, res.NextPageToken, nil
}

type updateNamespaceParams struct {
	namespace string
	spec      *namespace.NamespaceSpec

	asyncOperationID string
	idempotent       bool
	resourceVersion  string
}

func (c *namespaceClient) updateNamespace(ctx context.Context, params updateNamespaceParams) (*operation.AsyncOperation, error) {
	res, err := c.client.CloudService().UpdateNamespace(ctx, &cloudservice.UpdateNamespaceRequest{
		AsyncOperationId: params.asyncOperationID,
		Namespace:        params.namespace,
		ResourceVersion:  params.resourceVersion,
		Spec:             params.spec,
	})
	if err != nil {
		if isNothingChangedErr(params.idempotent, err) {
			return nil, nil
		}
		return nil, err
	}

	return res.AsyncOperation, nil
}

func (c *namespaceClient) createNamespace(ctx context.Context, n *namespace.NamespaceSpec, params applyNamespaceParams) (*operation.AsyncOperation, error) {
	res, err := c.client.CloudService().CreateNamespace(ctx, &cloudservice.CreateNamespaceRequest{
		AsyncOperationId: params.asyncOperationID,
		Spec:             n,
	})
	if err != nil {
		if isNothingChangedErr(params.idempotent, err) {
			return nil, nil
		}
		return nil, err
	}

	return res.AsyncOperation, nil
}

type deleteNamespaceParams struct {
	namespace string

	resourceVersion  string // optional, if empty, will be fetched
	asyncOperationID string
	idempotent       bool
}

func (c *namespaceClient) deleteNamespace(ctx context.Context, params deleteNamespaceParams) (*operation.AsyncOperation, error) {
	res, err := c.client.CloudService().DeleteNamespace(ctx, &cloudservice.DeleteNamespaceRequest{
		AsyncOperationId: params.asyncOperationID,
		Namespace:        params.namespace,
	})
	if err != nil {
		if isNotFoundErr(err) && params.idempotent {
			return nil, nil
		}
		return nil, err
	}

	return res.AsyncOperation, nil
}

type applyNamespaceParams struct {
	namespace string
	spec      *namespace.NamespaceSpec

	resourceVersion  string // optional, if empty, will be fetched
	asyncOperationID string
	idempotent       bool
}

func (c *namespaceClient) applyNamespace(ctx context.Context, params applyNamespaceParams) (*operation.AsyncOperation, error) {
	if params.resourceVersion == "" {
		// Try to get the existing namespace
		existing, err := c.getNamespace(ctx, params.namespace)
		if err != nil {
			return nil, err
		}
		params.resourceVersion = existing.ResourceVersion
	}

	// update
	// Namespace exists, update it using the current resource version
	updateParams := updateNamespaceParams{
		namespace: params.namespace,
		spec:      params.spec,

		asyncOperationID: params.asyncOperationID,
		idempotent:       params.idempotent,
		resourceVersion:  params.resourceVersion,
	}

	return c.updateNamespace(ctx, updateParams)
}
