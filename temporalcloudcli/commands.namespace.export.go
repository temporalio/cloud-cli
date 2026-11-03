package temporalcloudcli

import (
	"context"
	"fmt"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	sinkv1 "go.temporal.io/cloud-sdk/api/sink/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

type (
	GetExportSinkParams struct {
		Namespace string
		SinkName  string

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
	}

	ListExportSinksParams struct {
		Namespace string

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
	}

	DeleteExportSinkParams struct {
		Namespace        string
		SinkName         string
		ResourceVersion  string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	EnableExportSinkParams struct {
		Namespace        string
		SinkName         string
		ResourceVersion  string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	DisableExportSinkParams struct {
		Namespace        string
		SinkName         string
		ResourceVersion  string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}
)

func GetExportSink(ctx context.Context, params GetExportSinkParams) error {
	res, err := params.Cloud.GetNamespaceExportSink(ctx, &cloudservice.GetNamespaceExportSinkRequest{
		Namespace: params.Namespace,
		Name:      params.SinkName,
	})
	if err != nil {
		return err
	}
	return params.Printer.PrintResource(struct {
		Namespace string
		Spec      *namespacev1.ExportSinkSpec
	}{
		Namespace: params.Namespace,
		Spec:      res.Sink.Spec,
	}, printer.PrintResourceOptions{})
}

func ListExportSinks(ctx context.Context, params ListExportSinksParams) error {
	var sinks []*namespacev1.ExportSink
	pageToken := ""
	for {
		res, err := params.Cloud.GetNamespaceExportSinks(ctx, &cloudservice.GetNamespaceExportSinksRequest{
			Namespace: params.Namespace,
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
	return params.Printer.PrintResourceList(
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

func DeleteExportSink(ctx context.Context, params DeleteExportSinkParams) error {
	sinkRes, err := params.Cloud.GetNamespaceExportSink(ctx, &cloudservice.GetNamespaceExportSinkRequest{
		Namespace: params.Namespace,
		Name:      params.SinkName,
	})
	if err != nil {
		return err
	}
	sink := sinkRes.Sink

	if err := params.Prompter.PromptApply(sink.Spec, &namespacev1.ExportSinkSpec{}, false); err != nil {
		return err
	}

	rv := sink.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}

	deleteSink := wrapDeleteOperation(params.Cloud.DeleteNamespaceExportSink, params.OperationHandler, params.SinkName)
	return deleteSink(ctx, &cloudservice.DeleteNamespaceExportSinkRequest{
		Namespace:        params.Namespace,
		Name:             params.SinkName,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func EnableExportSink(ctx context.Context, params EnableExportSinkParams) error {
	sinkRes, err := params.Cloud.GetNamespaceExportSink(ctx, &cloudservice.GetNamespaceExportSinkRequest{
		Namespace: params.Namespace,
		Name:      params.SinkName,
	})
	if err != nil {
		return err
	}
	sink := sinkRes.Sink

	rv := sink.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}

	oldSpec := sink.Spec
	newSpec := proto.Clone(oldSpec).(*namespacev1.ExportSinkSpec)
	newSpec.Enabled = true

	if err := params.Prompter.PromptApply(oldSpec, newSpec, false); err != nil {
		return err
	}

	updateSink := wrapUpdateOperation(params.Cloud.UpdateNamespaceExportSink, params.OperationHandler, params.SinkName)
	return updateSink(ctx, &cloudservice.UpdateNamespaceExportSinkRequest{
		Namespace:        params.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func DisableExportSink(ctx context.Context, params DisableExportSinkParams) error {
	sinkRes, err := params.Cloud.GetNamespaceExportSink(ctx, &cloudservice.GetNamespaceExportSinkRequest{
		Namespace: params.Namespace,
		Name:      params.SinkName,
	})
	if err != nil {
		return err
	}
	sink := sinkRes.Sink

	rv := sink.ResourceVersion
	if params.ResourceVersion != "" {
		rv = params.ResourceVersion
	}

	oldSpec := sink.Spec
	newSpec := proto.Clone(oldSpec).(*namespacev1.ExportSinkSpec)
	newSpec.Enabled = false

	if err := params.Prompter.PromptApply(oldSpec, newSpec, false); err != nil {
		return err
	}

	updateSink := wrapUpdateOperation(params.Cloud.UpdateNamespaceExportSink, params.OperationHandler, params.SinkName)
	return updateSink(ctx, &cloudservice.UpdateNamespaceExportSinkRequest{
		Namespace:        params.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func (c *CloudNamespaceExportGetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return GetExportSink(cctx.Context, GetExportSinkParams{
		Namespace: c.Namespace,
		SinkName:  c.SinkName,
		Cloud:     cloudClient.CloudService(),
		Printer:   cctx.Printer,
	})
}

func (c *CloudNamespaceExportListCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return ListExportSinks(cctx.Context, ListExportSinksParams{
		Namespace: c.Namespace,
		Cloud:     cloudClient.CloudService(),
		Printer:   cctx.Printer,
	})
}

func (c *CloudNamespaceExportDeleteCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return DeleteExportSink(cctx.Context, DeleteExportSinkParams{
		Namespace:        c.Namespace,
		SinkName:         c.SinkName,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudNamespaceExportEnableCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return EnableExportSink(cctx.Context, EnableExportSinkParams{
		Namespace:        c.Namespace,
		SinkName:         c.SinkName,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudNamespaceExportDisableCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return DisableExportSink(cctx.Context, DisableExportSinkParams{
		Namespace:        c.Namespace,
		SinkName:         c.SinkName,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudNamespaceExportS3CreateCommand) run(cctx *CommandContext, _ []string) error {
	roleName, accountID, err := ParseRoleARN(c.RoleArn)
	if err != nil {
		return err
	}
	cloudClient, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	spec := &namespacev1.ExportSinkSpec{
		Name:    c.SinkName,
		Enabled: true,
		S3: &sinkv1.S3Spec{
			RoleName:     roleName,
			BucketName:   c.BucketName,
			Region:       c.Region,
			AwsAccountId: accountID,
			KmsArn:       c.KmsArn,
		},
	}
	if c.Command.Flags().Changed("disable-standalone-activity-export") {
		spec.Filter = &namespacev1.ExportSinkSpec_SinkFilter{
			DisableStandaloneActivityExport: c.DisableStandaloneActivityExport,
		}
	}

	yes, err := cctx.GetPrompter().PromptApply(&namespacev1.ExportSinkSpec{}, spec, false)
	if err != nil {
		return err
	}
	if !yes {
		return nil
	}
	resp, err := cloudClient.CreateNamespaceExportSink(cctx, &cloudservice.CreateNamespaceExportSinkRequest{
		Namespace:        c.Namespace,
		Spec:             spec,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(cloudClient, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudNamespaceExportS3UpdateCommand) run(cctx *CommandContext, _ []string) error {
	var roleName, accountID string
	if c.RoleArn != "" {
		var err error
		roleName, accountID, err = ParseRoleARN(c.RoleArn)
		if err != nil {
			return err
		}
	}
	cloudClient, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	sinkRes, err := cloudClient.GetNamespaceExportSink(cctx, &cloudservice.GetNamespaceExportSinkRequest{
		Namespace: c.Namespace,
		Name:      c.SinkName,
	})
	if err != nil {
		return err
	}
	sink := sinkRes.Sink

	rv := sink.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}

	oldSpec := sink.Spec
	newSpec := proto.Clone(oldSpec).(*namespacev1.ExportSinkSpec)
	newS3 := newSpec.GetS3()
	if roleName != "" {
		newS3.RoleName = roleName
	}
	if accountID != "" {
		newS3.AwsAccountId = accountID
	}
	if c.BucketName != "" {
		newS3.BucketName = c.BucketName
	}
	if c.KmsArn != "" {
		newS3.KmsArn = c.KmsArn
	}

	// AIDEV-NOTE: Check flag presence so an omitted bool preserves the existing filter.
	if c.Command.Flags().Changed("disable-standalone-activity-export") {
		if newSpec.Filter == nil {
			newSpec.Filter = &namespacev1.ExportSinkSpec_SinkFilter{}
		}
		newSpec.Filter.DisableStandaloneActivityExport = c.DisableStandaloneActivityExport
	}

	yes, err := cctx.GetPrompter().PromptApply(oldSpec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return nil
	}

	resp, err := cloudClient.UpdateNamespaceExportSink(cctx, &cloudservice.UpdateNamespaceExportSinkRequest{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(cloudClient, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNamespaceExportS3ValidateCommand) run(cctx *CommandContext, _ []string) error {
	roleName, accountID, err := ParseRoleARN(c.RoleArn)
	if err != nil {
		return err
	}
	cloudClient, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	spec := &namespacev1.ExportSinkSpec{
		Name: c.SinkName,
		S3: &sinkv1.S3Spec{
			RoleName:     roleName,
			BucketName:   c.BucketName,
			Region:       c.Region,
			AwsAccountId: accountID,
			KmsArn:       c.KmsArn,
		},
	}
	if c.Command.Flags().Changed("disable-standalone-activity-export") {
		spec.Filter = &namespacev1.ExportSinkSpec_SinkFilter{
			DisableStandaloneActivityExport: c.DisableStandaloneActivityExport,
		}
	}

	_, err = cloudClient.ValidateNamespaceExportSink(cctx, &cloudservice.ValidateNamespaceExportSinkRequest{
		Namespace: c.Namespace,
		Spec:      spec,
	})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintStructured(
		struct{ Status string }{Status: fmt.Sprintf("Export sink %q configuration is valid.", c.SinkName)},
		printer.StructuredOptions{},
	)
}

func (c *CloudNamespaceExportGcsCreateCommand) run(cctx *CommandContext, _ []string) error {
	saID, projectID, err := ParseServiceAccountEmail(c.ServiceAccountEmail)
	if err != nil {
		return err
	}
	cloudClient, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	spec := &namespacev1.ExportSinkSpec{
		Name:    c.SinkName,
		Enabled: true,
		Gcs: &sinkv1.GCSSpec{
			SaId:         saID,
			BucketName:   c.BucketName,
			GcpProjectId: projectID,
			Region:       c.Region,
		},
	}
	if c.Command.Flags().Changed("disable-standalone-activity-export") {
		spec.Filter = &namespacev1.ExportSinkSpec_SinkFilter{
			DisableStandaloneActivityExport: c.DisableStandaloneActivityExport,
		}
	}

	yes, err := cctx.GetPrompter().PromptApply(&namespacev1.ExportSinkSpec{}, spec, false)
	if err != nil {
		return err
	}
	if !yes {
		return nil
	}
	resp, err := cloudClient.CreateNamespaceExportSink(cctx, &cloudservice.CreateNamespaceExportSinkRequest{
		Namespace:        c.Namespace,
		Spec:             spec,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(cloudClient, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudNamespaceExportGcsUpdateCommand) run(cctx *CommandContext, _ []string) error {
	var saID, projectID string
	if c.ServiceAccountEmail != "" {
		var err error
		saID, projectID, err = ParseServiceAccountEmail(c.ServiceAccountEmail)
		if err != nil {
			return err
		}
	}
	cloudClient, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	sinkRes, err := cloudClient.GetNamespaceExportSink(cctx, &cloudservice.GetNamespaceExportSinkRequest{
		Namespace: c.Namespace,
		Name:      c.SinkName,
	})
	if err != nil {
		return err
	}
	sink := sinkRes.Sink

	rv := sink.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}

	oldSpec := sink.Spec
	newSpec := proto.Clone(oldSpec).(*namespacev1.ExportSinkSpec)
	newGcs := newSpec.GetGcs()
	if saID != "" {
		newGcs.SaId = saID
	}
	if projectID != "" {
		newGcs.GcpProjectId = projectID
	}
	if c.BucketName != "" {
		newGcs.BucketName = c.BucketName
	}

	// AIDEV-NOTE: Check flag presence so an omitted bool preserves the existing filter.
	if c.Command.Flags().Changed("disable-standalone-activity-export") {
		if newSpec.Filter == nil {
			newSpec.Filter = &namespacev1.ExportSinkSpec_SinkFilter{}
		}
		newSpec.Filter.DisableStandaloneActivityExport = c.DisableStandaloneActivityExport
	}

	yes, err := cctx.GetPrompter().PromptApply(oldSpec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return nil
	}

	resp, err := cloudClient.UpdateNamespaceExportSink(cctx, &cloudservice.UpdateNamespaceExportSinkRequest{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(cloudClient, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNamespaceExportGcsValidateCommand) run(cctx *CommandContext, _ []string) error {
	saID, projectID, err := ParseServiceAccountEmail(c.ServiceAccountEmail)
	if err != nil {
		return err
	}
	cloudClient, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	spec := &namespacev1.ExportSinkSpec{
		Name: c.SinkName,
		Gcs: &sinkv1.GCSSpec{
			SaId:         saID,
			BucketName:   c.BucketName,
			GcpProjectId: projectID,
			Region:       c.Region,
		},
	}
	if c.Command.Flags().Changed("disable-standalone-activity-export") {
		spec.Filter = &namespacev1.ExportSinkSpec_SinkFilter{
			DisableStandaloneActivityExport: c.DisableStandaloneActivityExport,
		}
	}

	_, err = cloudClient.ValidateNamespaceExportSink(cctx, &cloudservice.ValidateNamespaceExportSinkRequest{
		Namespace: c.Namespace,
		Spec:      spec,
	})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintStructured(
		struct{ Status string }{Status: fmt.Sprintf("Export sink %q configuration is valid.", c.SinkName)},
		printer.StructuredOptions{},
	)
}

func (c *CloudNamespaceExportAzureBlobCreateCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	spec := &namespacev1.ExportSinkSpec{
		Name:    c.SinkName,
		Enabled: true,
		AzureBlob: &sinkv1.AzureBlobSpec{
			TenantId:       c.TenantId,
			SubscriptionId: c.SubscriptionId,
			ResourceGroup:  c.ResourceGroup,
			StorageAccount: c.StorageAccount,
			ContainerName:  c.ContainerName,
			Region:         c.Region,
		},
	}
	if c.Command.Flags().Changed("disable-standalone-activity-export") {
		spec.Filter = &namespacev1.ExportSinkSpec_SinkFilter{
			DisableStandaloneActivityExport: c.DisableStandaloneActivityExport,
		}
	}

	yes, err := cctx.GetPrompter().PromptApply(&namespacev1.ExportSinkSpec{}, spec, false)
	if err != nil {
		return err
	}
	if !yes {
		return nil
	}
	resp, err := cloudClient.CreateNamespaceExportSink(cctx, &cloudservice.CreateNamespaceExportSinkRequest{
		Namespace:        c.Namespace,
		Spec:             spec,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(cloudClient, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudNamespaceExportAzureBlobUpdateCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	sinkRes, err := cloudClient.GetNamespaceExportSink(cctx, &cloudservice.GetNamespaceExportSinkRequest{
		Namespace: c.Namespace,
		Name:      c.SinkName,
	})
	if err != nil {
		return err
	}
	sink := sinkRes.Sink

	rv := sink.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}

	oldSpec := sink.Spec
	newSpec := proto.Clone(oldSpec).(*namespacev1.ExportSinkSpec)
	newAzureBlob := newSpec.GetAzureBlob()
	if c.TenantId != "" {
		newAzureBlob.TenantId = c.TenantId
	}
	if c.SubscriptionId != "" {
		newAzureBlob.SubscriptionId = c.SubscriptionId
	}
	if c.ResourceGroup != "" {
		newAzureBlob.ResourceGroup = c.ResourceGroup
	}
	if c.StorageAccount != "" {
		newAzureBlob.StorageAccount = c.StorageAccount
	}
	if c.ContainerName != "" {
		newAzureBlob.ContainerName = c.ContainerName
	}

	// AIDEV-NOTE: Check flag presence so an omitted bool preserves the existing filter.
	if c.Command.Flags().Changed("disable-standalone-activity-export") {
		if newSpec.Filter == nil {
			newSpec.Filter = &namespacev1.ExportSinkSpec_SinkFilter{}
		}
		newSpec.Filter.DisableStandaloneActivityExport = c.DisableStandaloneActivityExport
	}

	yes, err := cctx.GetPrompter().PromptApply(oldSpec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return nil
	}

	resp, err := cloudClient.UpdateNamespaceExportSink(cctx, &cloudservice.UpdateNamespaceExportSinkRequest{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(cloudClient, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNamespaceExportAzureBlobValidateCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	spec := &namespacev1.ExportSinkSpec{
		Name: c.SinkName,
		AzureBlob: &sinkv1.AzureBlobSpec{
			TenantId:       c.TenantId,
			SubscriptionId: c.SubscriptionId,
			ResourceGroup:  c.ResourceGroup,
			StorageAccount: c.StorageAccount,
			ContainerName:  c.ContainerName,
			Region:         c.Region,
		},
	}
	if c.Command.Flags().Changed("disable-standalone-activity-export") {
		spec.Filter = &namespacev1.ExportSinkSpec_SinkFilter{
			DisableStandaloneActivityExport: c.DisableStandaloneActivityExport,
		}
	}

	_, err = cloudClient.ValidateNamespaceExportSink(cctx, &cloudservice.ValidateNamespaceExportSinkRequest{
		Namespace: c.Namespace,
		Spec:      spec,
	})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintStructured(
		struct{ Status string }{Status: fmt.Sprintf("Export sink %q configuration is valid.", c.SinkName)},
		printer.StructuredOptions{},
	)
}
