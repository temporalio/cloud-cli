package temporalcloudcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	accountv1 "go.temporal.io/cloud-sdk/api/account/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAuditLogSinks_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	sinks := []*accountv1.AuditLogSink{
		{Name: "sink-1"},
		{Name: "sink-2"},
	}
	mockCloud.EXPECT().
		GetAccountAuditLogSinks(context.Background(), &cloudservice.GetAccountAuditLogSinksRequest{}).
		Return(&cloudservice.GetAccountAuditLogSinksResponse{
			Sinks:         sinks,
			NextPageToken: "next-token",
		}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.ListAuditLogSinks(context.Background(), temporalcloudcli.ListAuditLogSinksParams{
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)

	var out struct {
		Sinks         []*accountv1.AuditLogSink `json:"sinks"`
		NextPageToken string                    `json:"nextPageToken"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "sink-1", out.Sinks[0].Name)
	assert.Equal(t, "sink-2", out.Sinks[1].Name)
	assert.Equal(t, "next-token", out.NextPageToken)
}

func TestListAuditLogSinks_WithPagination(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	mockCloud.EXPECT().
		GetAccountAuditLogSinks(context.Background(), &cloudservice.GetAccountAuditLogSinksRequest{
			PageSize:  50,
			PageToken: "some-token",
		}).
		Return(&cloudservice.GetAccountAuditLogSinksResponse{}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.ListAuditLogSinks(context.Background(), temporalcloudcli.ListAuditLogSinksParams{
		PageSize:  50,
		PageToken: "some-token",
		Cloud:     mockCloud,
		Printer:   &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)
}

func TestListAuditLogSinks_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetAccountAuditLogSinks(context.Background(), &cloudservice.GetAccountAuditLogSinksRequest{}).
		Return(nil, apiErr)

	var buf bytes.Buffer
	err := temporalcloudcli.ListAuditLogSinks(context.Background(), temporalcloudcli.ListAuditLogSinksParams{
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.ErrorIs(t, err, apiErr)
	assert.Empty(t, buf.String())
}
