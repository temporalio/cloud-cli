package asyncoperation

import (
	"context"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
	"google.golang.org/grpc"
)

// CloudService defines the interface for cloud service operations needed by the asyncoperation client.
type CloudService interface {
	GetAsyncOperation(ctx context.Context, req *cloudservice.GetAsyncOperationRequest, opts ...grpc.CallOption) (*cloudservice.GetAsyncOperationResponse, error)
}

func NewClient(cloudClient cloudservice.CloudServiceClient) *Client {
	return &Client{Cloud: cloudClient}
}

type Client struct {
	Cloud CloudService
}

func (c *Client) GetAsyncOperation(ctx context.Context, asyncOperationID string) (*operation.AsyncOperation, error) {
	res, err := c.Cloud.GetAsyncOperation(ctx, &cloudservice.GetAsyncOperationRequest{
		AsyncOperationId: asyncOperationID,
	})
	if err != nil {
		return nil, err
	}
	return res.AsyncOperation, nil
}
