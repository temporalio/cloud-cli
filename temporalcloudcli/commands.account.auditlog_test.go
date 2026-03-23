package temporalcloudcli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
	auditlogv1 "go.temporal.io/cloud-sdk/api/auditlog/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAuditLogs_Success(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	expected := struct {
		AuditLogs []*auditlogv1.LogRecord `json:"auditLogs"`
		NextPageToken string              `json:"nextPageToken"`
	}{
		AuditLogs: []*auditlogv1.LogRecord{
			{LogId: "log-1", Operation: "CreateNamespace"},
			{LogId: "log-2", Operation: "DeleteNamespace"},
		},
		NextPageToken: "next-token",
	}
	mockCloud.EXPECT().
		GetAuditLogs(context.Background(), &cloudservice.GetAuditLogsRequest{}).
		Return(&cloudservice.GetAuditLogsResponse{
			Logs:          expected.AuditLogs,
			NextPageToken: expected.NextPageToken,
		}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.GetAuditLogs(context.Background(), temporalcloudcli.GetAuditLogsParams{
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)

	var out struct {
		AuditLogs []*auditlogv1.LogRecord `json:"auditLogs"`
		NextPageToken string              `json:"nextPageToken"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, expected, out)
}

func TestGetAuditLogs_WithPagination(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	mockCloud.EXPECT().
		GetAuditLogs(context.Background(), &cloudservice.GetAuditLogsRequest{
			PageSize:  50,
			PageToken: "some-token",
		}).
		Return(&cloudservice.GetAuditLogsResponse{}, nil)

	var buf bytes.Buffer
	err := temporalcloudcli.GetAuditLogs(context.Background(), temporalcloudcli.GetAuditLogsParams{
		PageSize:  50,
		PageToken: "some-token",
		Cloud:     mockCloud,
		Printer:   &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)
}

func TestGetAuditLogs_WithTimeFilters(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)

	startTime, err := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	require.NoError(t, err)
	endTime, err := time.Parse(time.RFC3339, "2024-02-01T00:00:00Z")
	require.NoError(t, err)

	mockCloud.EXPECT().
		GetAuditLogs(context.Background(), &cloudservice.GetAuditLogsRequest{
			StartTimeInclusive: timestamppb.New(startTime),
			EndTimeExclusive:   timestamppb.New(endTime),
		}).
		Return(&cloudservice.GetAuditLogsResponse{}, nil)

	var buf bytes.Buffer
	err = temporalcloudcli.GetAuditLogs(context.Background(), temporalcloudcli.GetAuditLogsParams{
		StartTime: startTime,
		EndTime:   endTime,
		Cloud:     mockCloud,
		Printer:   &printer.Printer{Output: &buf, JSON: true},
	})
	require.NoError(t, err)
}

func TestGetAuditLogs_APIError(t *testing.T) {
	mockCloud := cloudmock.NewMockCloudServiceClient(t)
	apiErr := errors.New("api error")

	mockCloud.EXPECT().
		GetAuditLogs(context.Background(), &cloudservice.GetAuditLogsRequest{}).
		Return(nil, apiErr)

	var buf bytes.Buffer
	err := temporalcloudcli.GetAuditLogs(context.Background(), temporalcloudcli.GetAuditLogsParams{
		Cloud:   mockCloud,
		Printer: &printer.Printer{Output: &buf, JSON: true},
	})
	require.ErrorIs(t, err, apiErr)
	assert.Empty(t, buf.String())
}
