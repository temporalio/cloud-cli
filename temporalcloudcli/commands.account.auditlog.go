package temporalcloudcli

import (
	"context"
	"time"

	auditlogv1 "go.temporal.io/cloud-sdk/api/auditlog/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type GetAuditLogsParams struct {
	PageSize  int32
	PageToken string
	StartTime time.Time
	EndTime   time.Time

	Cloud   cloudservice.CloudServiceClient
	Printer *printer.Printer
}

func GetAuditLogs(ctx context.Context, params GetAuditLogsParams) error {
	req := &cloudservice.GetAuditLogsRequest{
		PageSize:  params.PageSize,
		PageToken: params.PageToken,
	}
	if !params.StartTime.IsZero() {
		req.StartTimeInclusive = timestamppb.New(params.StartTime)
	}
	if !params.EndTime.IsZero() {
		req.EndTimeExclusive = timestamppb.New(params.EndTime)
	}
	res, err := params.Cloud.GetAuditLogs(ctx, req)
	if err != nil {
		return err
	}
	return params.Printer.PrintResourceList(
		struct {
			AuditLogs     []*auditlogv1.LogRecord
			NextPageToken string
		}{
			AuditLogs:     res.GetLogs(),
			NextPageToken: res.GetNextPageToken(),
		},
		printer.PrintResourceOptions{
			Fields: []string{"EmitTime", "Operation", "Status", "Principal"},
		},
		printer.TableOptions{},
	)
}

func (c *CloudAccountAuditLogGetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return GetAuditLogs(cctx.Context, GetAuditLogsParams{
		PageSize:  int32(c.PageSize),
		PageToken: c.PageToken,
		StartTime: c.StartTime.Time(),
		EndTime:   c.EndTime.Time(),
		Cloud:     cloudClient.CloudService(),
		Printer:   cctx.Printer,
	})
}
