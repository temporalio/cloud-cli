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

type updateNamespaceParams struct {
	asyncOperationID string
	idempotent       bool
	resourceVersion  string
	// namespace is the full name of the namespace including the account
	namespace string
}

func (c *namespaceClient) updateNamespace(ctx context.Context, n *namespace.NamespaceSpec, params updateNamespaceParams) (*operation.AsyncOperation, error) {
	res, err := c.client.CloudService().UpdateNamespace(ctx, &cloudservice.UpdateNamespaceRequest{
		AsyncOperationId: params.asyncOperationID,
		Namespace:        params.namespace,
		ResourceVersion:  params.resourceVersion,
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

type applyNamespaceParams struct {
	asyncOperationID string
	idempotent       bool
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

func (c *namespaceClient) applyNamespace(ctx context.Context, n *namespace.NamespaceSpec, params applyNamespaceParams) (*operation.AsyncOperation, error) {
	// Try to get the existing namespace
	namespaces, err := c.listNamespacesWithName(ctx, n.GetName())
	if err != nil {
		return nil, err
	} else if len(namespaces) > 1 {
		return nil, fmt.Errorf("multiple namespaces match namespace name: %s", n.GetName())
	} else if len(namespaces) == 0 {
		return c.createNamespace(ctx, n, params)
	}

	existing := namespaces[0]

	// update
	// Namespace exists, update it using the current resource version
	updateParams := updateNamespaceParams{
		asyncOperationID: params.asyncOperationID,
		idempotent:       params.idempotent,
		resourceVersion:  existing.ResourceVersion,
		namespace:        existing.Namespace,
	}

	return c.updateNamespace(ctx, n, updateParams)
}

func (c *namespaceClient) listNamespacesWithName(ctx context.Context, name string) ([]*namespace.Namespace, error) {
	namespaces := []*namespace.Namespace{}
	pageToken := ""
	for {
		res, err := c.client.CloudService().GetNamespaces(ctx, &cloudservice.GetNamespacesRequest{
			Name:      name,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, err
		}
		namespaces = append(namespaces, res.Namespaces...)
		// Check if we should continue paging
		pageToken = res.NextPageToken
		if len(pageToken) == 0 {
			break
		}
	}
	return namespaces, nil
}
