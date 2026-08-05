package namespace_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"

	"github.com/temporalio/cloud-cli/internal/namespace"
	namespacemock "github.com/temporalio/cloud-cli/internal/namespace/mock"
)

func TestClientGetNamespaces_ProjectID(t *testing.T) {
	cloud := namespacemock.NewMockCloudService(t)
	client := &namespace.Client{Cloud: cloud}

	cloud.EXPECT().
		GetNamespaces(mock.Anything, &cloudservice.GetNamespacesRequest{
			PageSize:  50,
			PageToken: "page-1",
			Name:      "payments",
			ProjectId: "project-123",
		}, mock.Anything).
		Return(&cloudservice.GetNamespacesResponse{
			Namespaces: []*namespacev1.Namespace{{Namespace: "payments.account"}},
		}, nil)

	namespaces, nextPageToken, err := client.GetNamespaces(context.Background(), namespace.GetNamespacesParams{
		PageSize:  50,
		PageToken: "page-1",
		Name:      "payments",
		ProjectID: "project-123",
	})
	require.NoError(t, err)
	require.Empty(t, nextPageToken)
	require.Len(t, namespaces, 1)
}
