package namespace

import (
	"context"

	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"
)

// GetCodecServer retrieves the codec server configuration for the specified namespace.
func (c *Client) GetCodecServer(ctx context.Context, namespaceName string) (*namespacev1.CodecServerSpec, error) {
	ns, err := c.GetNamespace(ctx, namespaceName)
	if err != nil {
		return nil, err
	}
	return ns.GetSpec().GetCodecServer(), nil
}

// SetCodecParams contains parameters for setting the codec server on a namespace.
type SetCodecParams struct {
	Namespace                        string
	Endpoint                         string
	PassAccessToken                  bool
	IncludeCrossOriginCredentials    bool
	CustomErrorMessageDefaultMessage string
	CustomErrorMessageDefaultLink    string
	ResourceVersion                  string
	AsyncOperationID                 string
}

// SetCodec sets the codec server configuration on the specified namespace.
func (c *Client) SetCodec(ctx context.Context, params SetCodecParams) (*operation.AsyncOperation, error) {
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	newSpec := ns.Spec
	newSpec.CodecServer = &namespacev1.CodecServerSpec{
		Endpoint:                      params.Endpoint,
		PassAccessToken:               params.PassAccessToken,
		IncludeCrossOriginCredentials: params.IncludeCrossOriginCredentials,
	}
	if params.CustomErrorMessageDefaultMessage != "" || params.CustomErrorMessageDefaultLink != "" {
		newSpec.CodecServer.CustomErrorMessage = &namespacev1.CodecServerSpec_CustomErrorMessage{
			Default: &namespacev1.CodecServerSpec_CustomErrorMessage_ErrorMessage{
				Message: params.CustomErrorMessageDefaultMessage,
				Link:    params.CustomErrorMessageDefaultLink,
			},
		}
	}

	resourceVersion := ns.ResourceVersion
	if params.ResourceVersion != "" {
		resourceVersion = params.ResourceVersion
	}

	return c.UpdateNamespace(ctx, UpdateNamespaceParams{
		Namespace:        params.Namespace,
		Spec:             newSpec,
		ResourceVersion:  resourceVersion,
		AsyncOperationID: params.AsyncOperationID,
	})
}

// DeleteCodecParams contains parameters for removing the codec server from a namespace.
type DeleteCodecParams struct {
	Namespace        string
	ResourceVersion  string
	AsyncOperationID string
}

// DeleteCodec removes the codec server configuration from the specified namespace.
func (c *Client) DeleteCodec(ctx context.Context, params DeleteCodecParams) (*operation.AsyncOperation, error) {
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	newSpec := ns.Spec
	newSpec.CodecServer = nil

	resourceVersion := ns.ResourceVersion
	if params.ResourceVersion != "" {
		resourceVersion = params.ResourceVersion
	}

	return c.UpdateNamespace(ctx, UpdateNamespaceParams{
		Namespace:        params.Namespace,
		Spec:             newSpec,
		ResourceVersion:  resourceVersion,
		AsyncOperationID: params.AsyncOperationID,
	})
}
