package namespace

import (
	"context"
	"fmt"

	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/grpc"
)

// CloudService defines the interface for cloud service operations needed by the namespace client.
type CloudService interface {
	GetNamespace(ctx context.Context, req *cloudservice.GetNamespaceRequest, opts ...grpc.CallOption) (*cloudservice.GetNamespaceResponse, error)
	GetNamespaces(ctx context.Context, req *cloudservice.GetNamespacesRequest, opts ...grpc.CallOption) (*cloudservice.GetNamespacesResponse, error)
	UpdateNamespace(ctx context.Context, req *cloudservice.UpdateNamespaceRequest, opts ...grpc.CallOption) (*cloudservice.UpdateNamespaceResponse, error)
}

func NewClient(cloudClient cloudservice.CloudServiceClient) *Client {
	return &Client{Cloud: cloudClient}
}

type Client struct {
	Cloud CloudService
}

func (c *Client) GetNamespace(ctx context.Context, namespace string) (*namespacev1.Namespace, error) {
	res, err := c.Cloud.GetNamespace(ctx, &cloudservice.GetNamespaceRequest{
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

type GetNamespacesParams struct {
	PageSize  int32
	PageToken string
	Name      string
}

func (c *Client) GetNamespaces(ctx context.Context, params GetNamespacesParams) ([]*namespacev1.Namespace, string, error) {
	res, err := c.Cloud.GetNamespaces(ctx, &cloudservice.GetNamespacesRequest{
		PageSize:  params.PageSize,
		PageToken: params.PageToken,
		Name:      params.Name,
	})
	if err != nil {
		return nil, "", err
	}

	return res.Namespaces, res.NextPageToken, nil
}

type UpdateNamespaceParams struct {
	Namespace        string
	Spec             *namespacev1.NamespaceSpec
	AsyncOperationID string
	ResourceVersion  string
}

func (c *Client) UpdateNamespace(ctx context.Context, params UpdateNamespaceParams) (*operation.AsyncOperation, error) {
	res, err := c.Cloud.UpdateNamespace(ctx, &cloudservice.UpdateNamespaceRequest{
		AsyncOperationId: params.AsyncOperationID,
		Namespace:        params.Namespace,
		ResourceVersion:  params.ResourceVersion,
		Spec:             params.Spec,
	})
	if err != nil {
		return nil, err
	}

	return res.AsyncOperation, nil
}
