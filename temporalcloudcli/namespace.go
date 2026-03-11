package temporalcloudcli

import (
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
