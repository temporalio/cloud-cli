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
	CreateNamespace(ctx context.Context, req *cloudservice.CreateNamespaceRequest, opts ...grpc.CallOption) (*cloudservice.CreateNamespaceResponse, error)
	DeleteNamespace(ctx context.Context, req *cloudservice.DeleteNamespaceRequest, opts ...grpc.CallOption) (*cloudservice.DeleteNamespaceResponse, error)
	UpdateNamespace(ctx context.Context, req *cloudservice.UpdateNamespaceRequest, opts ...grpc.CallOption) (*cloudservice.UpdateNamespaceResponse, error)
	RenameCustomSearchAttribute(ctx context.Context, req *cloudservice.RenameCustomSearchAttributeRequest, opts ...grpc.CallOption) (*cloudservice.RenameCustomSearchAttributeResponse, error)
	UpdateNamespaceTags(ctx context.Context, req *cloudservice.UpdateNamespaceTagsRequest, opts ...grpc.CallOption) (*cloudservice.UpdateNamespaceTagsResponse, error)
	AddNamespaceRegion(ctx context.Context, req *cloudservice.AddNamespaceRegionRequest, opts ...grpc.CallOption) (*cloudservice.AddNamespaceRegionResponse, error)
	DeleteNamespaceRegion(ctx context.Context, req *cloudservice.DeleteNamespaceRegionRequest, opts ...grpc.CallOption) (*cloudservice.DeleteNamespaceRegionResponse, error)
	FailoverNamespaceRegion(ctx context.Context, req *cloudservice.FailoverNamespaceRegionRequest, opts ...grpc.CallOption) (*cloudservice.FailoverNamespaceRegionResponse, error)
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

type UpsertNamespaceParams struct {
	Namespace        string
	Spec             *namespacev1.NamespaceSpec
	AsyncOperationID string
	ResourceVersion  string
}

func (c *Client) UpsertNamespace(ctx context.Context, params UpsertNamespaceParams) (*operation.AsyncOperation, error) {
	if params.Namespace == "" {
		res, err := c.Cloud.CreateNamespace(ctx, &cloudservice.CreateNamespaceRequest{
			Spec:             params.Spec,
			AsyncOperationId: params.AsyncOperationID,
		}, nil)
		if err != nil {
			return nil, err
		}

		return res.AsyncOperation, nil
	}

	res, err := c.Cloud.UpdateNamespace(ctx, &cloudservice.UpdateNamespaceRequest{
		Namespace:       params.Namespace,
		Spec:            params.Spec,
		ResourceVersion: params.ResourceVersion,
	})
	if err != nil {
		return nil, err
	}

	return res.AsyncOperation, nil
}

type DeleteNamespaceParams struct {
	Namespace        string
	AsyncOperationID string
	ResourceVersion  string
}

func (c *Client) DeleteNamespace(ctx context.Context, params DeleteNamespaceParams) (*operation.AsyncOperation, error) {
	getRes, err := c.Cloud.GetNamespace(ctx, &cloudservice.GetNamespaceRequest{Namespace: params.Namespace})
	if err != nil {
		return nil, err
	}

	if params.ResourceVersion == "" {
		params.ResourceVersion = getRes.Namespace.ResourceVersion
	}

	deleteRes, err := c.Cloud.DeleteNamespace(ctx, &cloudservice.DeleteNamespaceRequest{
		Namespace:        params.Namespace,
		ResourceVersion:  params.ResourceVersion,
		AsyncOperationId: params.AsyncOperationID,
	})
	if err != nil {
		return nil, err
	}

	return deleteRes.AsyncOperation, nil
}

func (c *Client) ListNamespaces(ctx context.Context, name, pageToken string, pageSize int32) ([]*namespacev1.Namespace, string, error) {
	res, err := c.Cloud.GetNamespaces(ctx, &cloudservice.GetNamespacesRequest{
		PageSize:  pageSize,
		PageToken: pageToken,
		Name:      name,
	})
	if err != nil {
		return nil, "", err
	}

	return res.Namespaces, res.NextPageToken, nil
}
