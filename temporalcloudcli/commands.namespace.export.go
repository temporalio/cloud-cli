package temporalcloudcli

import (
	"errors"
	"fmt"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	sinkv1 "go.temporal.io/cloud-sdk/api/sink/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceExportGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespaceExportSink(cctx, &cloudservice.GetNamespaceExportSinkRequest{
		Namespace: c.Namespace,
		Name:      c.SinkName,
	})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResource(struct {
		Namespace string
		Spec      *namespacev1.ExportSinkSpec
	}{
		Namespace: c.Namespace,
		Spec:      res.Sink.Spec,
	}, printer.PrintResourceOptions{})
}

func (c *CloudNamespaceExportListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	var sinks []*namespacev1.ExportSink
	pageToken := ""
	for {
		res, err := client.GetNamespaceExportSinks(cctx, &cloudservice.GetNamespaceExportSinksRequest{
			Namespace: c.Namespace,
			PageToken: pageToken,
		})
		if err != nil {
			return err
		}
		sinks = append(sinks, res.Sinks...)
		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}
	return cctx.Printer.PrintResourceList(
		struct {
			Sinks []*namespacev1.ExportSink
		}{Sinks: sinks},
		printer.PrintResourceOptions{
			Fields:     []string{"Name", "State", "Health"},
			SpecFields: []string{"Enabled"},
		},
		printer.TableOptions{},
	)
}

func (c *CloudNamespaceExportDeleteCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespaceExportSink(cctx, &cloudservice.GetNamespaceExportSinkRequest{
		Namespace: c.Namespace,
		Name:      c.SinkName,
	})
	if err != nil {
		return err
	}
	sink := res.Sink

	yes, err := cctx.GetPrompter().PromptApply(sink.Spec, &namespacev1.ExportSinkSpec{}, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting delete.")
	}

	rv := sink.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.DeleteNamespaceExportSink(cctx, &cloudservice.DeleteNamespaceExportSinkRequest{
		Namespace:        c.Namespace,
		Name:             c.SinkName,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleDeleteOperation(cctx, resp, err)
}

func (c *CloudNamespaceExportEnableCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespaceExportSink(cctx, &cloudservice.GetNamespaceExportSinkRequest{
		Namespace: c.Namespace,
		Name:      c.SinkName,
	})
	if err != nil {
		return err
	}
	sink := res.Sink
	newSpec := proto.Clone(sink.Spec).(*namespacev1.ExportSinkSpec)
	newSpec.Enabled = true

	yes, err := cctx.GetPrompter().PromptApply(sink.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting enable.")
	}

	rv := sink.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateNamespaceExportSink(cctx, &cloudservice.UpdateNamespaceExportSinkRequest{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNamespaceExportDisableCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespaceExportSink(cctx, &cloudservice.GetNamespaceExportSinkRequest{
		Namespace: c.Namespace,
		Name:      c.SinkName,
	})
	if err != nil {
		return err
	}
	sink := res.Sink
	newSpec := proto.Clone(sink.Spec).(*namespacev1.ExportSinkSpec)
	newSpec.Enabled = false

	yes, err := cctx.GetPrompter().PromptApply(sink.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting disable.")
	}

	rv := sink.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateNamespaceExportSink(cctx, &cloudservice.UpdateNamespaceExportSinkRequest{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNamespaceExportS3CreateCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	spec := &namespacev1.ExportSinkSpec{
		Name:    c.SinkName,
		Enabled: true,
		S3: &sinkv1.S3Spec{
			RoleName:     c.RoleName,
			BucketName:   c.BucketName,
			Region:       c.Region,
			AwsAccountId: c.AwsAccountId,
			KmsArn:       c.KmsArn,
		},
	}
	yes, err := cctx.GetPrompter().PromptApply(&namespacev1.ExportSinkSpec{}, spec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
	}

	resp, err := client.CreateNamespaceExportSink(cctx, &cloudservice.CreateNamespaceExportSinkRequest{
		Namespace:        c.Namespace,
		Spec:             spec,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudNamespaceExportS3UpdateCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespaceExportSink(cctx, &cloudservice.GetNamespaceExportSinkRequest{
		Namespace: c.Namespace,
		Name:      c.SinkName,
	})
	if err != nil {
		return err
	}
	sink := res.Sink

	newSpec := &namespacev1.ExportSinkSpec{
		Name:    c.SinkName,
		Enabled: sink.Spec.GetEnabled(),
		S3: &sinkv1.S3Spec{
			RoleName:     c.RoleName,
			BucketName:   c.BucketName,
			Region:       sink.Spec.GetS3().GetRegion(),
			AwsAccountId: c.AwsAccountId,
			KmsArn:       c.KmsArn,
		},
	}

	yes, err := cctx.GetPrompter().PromptApply(sink.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting update.")
	}

	rv := sink.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateNamespaceExportSink(cctx, &cloudservice.UpdateNamespaceExportSinkRequest{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNamespaceExportS3ValidateCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	spec := &namespacev1.ExportSinkSpec{
		Name: c.SinkName,
		S3: &sinkv1.S3Spec{
			RoleName:     c.RoleName,
			BucketName:   c.BucketName,
			Region:       c.Region,
			AwsAccountId: c.AwsAccountId,
			KmsArn:       c.KmsArn,
		},
	}
	if _, err := client.ValidateNamespaceExportSink(cctx, &cloudservice.ValidateNamespaceExportSinkRequest{
		Namespace: c.Namespace,
		Spec:      spec,
	}); err != nil {
		return err
	}
	return cctx.Printer.PrintStructured(
		struct{ Status string }{Status: fmt.Sprintf("Export sink %q configuration is valid.", c.SinkName)},
		printer.StructuredOptions{},
	)
}

func (c *CloudNamespaceExportGcsCreateCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	spec := &namespacev1.ExportSinkSpec{
		Name:    c.SinkName,
		Enabled: true,
		Gcs: &sinkv1.GCSSpec{
			SaId:         c.SaId,
			BucketName:   c.BucketName,
			GcpProjectId: c.GcpProjectId,
			Region:       c.Region,
		},
	}
	yes, err := cctx.GetPrompter().PromptApply(&namespacev1.ExportSinkSpec{}, spec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
	}

	resp, err := client.CreateNamespaceExportSink(cctx, &cloudservice.CreateNamespaceExportSinkRequest{
		Namespace:        c.Namespace,
		Spec:             spec,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudNamespaceExportGcsUpdateCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespaceExportSink(cctx, &cloudservice.GetNamespaceExportSinkRequest{
		Namespace: c.Namespace,
		Name:      c.SinkName,
	})
	if err != nil {
		return err
	}
	sink := res.Sink

	newSpec := &namespacev1.ExportSinkSpec{
		Name:    c.SinkName,
		Enabled: sink.Spec.GetEnabled(),
		Gcs: &sinkv1.GCSSpec{
			SaId:         c.SaId,
			BucketName:   c.BucketName,
			GcpProjectId: c.GcpProjectId,
			Region:       sink.Spec.GetGcs().GetRegion(),
		},
	}

	yes, err := cctx.GetPrompter().PromptApply(sink.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting update.")
	}

	rv := sink.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateNamespaceExportSink(cctx, &cloudservice.UpdateNamespaceExportSinkRequest{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNamespaceExportGcsValidateCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	spec := &namespacev1.ExportSinkSpec{
		Name: c.SinkName,
		Gcs: &sinkv1.GCSSpec{
			SaId:         c.SaId,
			BucketName:   c.BucketName,
			GcpProjectId: c.GcpProjectId,
			Region:       c.Region,
		},
	}
	if _, err := client.ValidateNamespaceExportSink(cctx, &cloudservice.ValidateNamespaceExportSinkRequest{
		Namespace: c.Namespace,
		Spec:      spec,
	}); err != nil {
		return err
	}
	return cctx.Printer.PrintStructured(
		struct{ Status string }{Status: fmt.Sprintf("Export sink %q configuration is valid.", c.SinkName)},
		printer.StructuredOptions{},
	)
}
