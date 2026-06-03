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

	CreateS3ExportSinkParams struct {
		Namespace        string
		SinkName         string
		RoleName         string
		BucketName       string
		Region           string
		AwsAccountID     string
		KmsArn           string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	UpdateS3ExportSinkParams struct {
		Namespace        string
		SinkName         string
		RoleName         string
		BucketName       string
		AwsAccountID     string
		KmsArn           string
		ResourceVersion  string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	ValidateS3ExportSinkParams struct {
		Namespace    string
		SinkName     string
		RoleName     string
		BucketName   string
		Region       string
		AwsAccountID string
		KmsArn       string

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
	}

	CreateGCSExportSinkParams struct {
		Namespace        string
		SinkName         string
		SaID             string
		BucketName       string
		GcpProjectID     string
		Region           string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	UpdateGCSExportSinkParams struct {
		Namespace        string
		SinkName         string
		SaID             string
		BucketName       string
		GcpProjectID     string
		ResourceVersion  string
		AsyncOperationID string

		Cloud            cloudservice.CloudServiceClient
		Prompter         Prompter
		OperationHandler AsyncOperationHandler
	}

	ValidateGCSExportSinkParams struct {
		Namespace    string
		SinkName     string
		SaID         string
		BucketName   string
		GcpProjectID string
		Region       string

		Cloud   cloudservice.CloudServiceClient
		Printer *printer.Printer
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

func CreateS3ExportSink(ctx context.Context, params CreateS3ExportSinkParams) error {
	spec := &namespacev1.ExportSinkSpec{
		Name:    params.SinkName,
		Enabled: true,
		S3: &sinkv1.S3Spec{
			RoleName:     params.RoleName,
			BucketName:   params.BucketName,
			Region:       params.Region,
			AwsAccountId: params.AwsAccountID,
			KmsArn:       params.KmsArn,
		},
	}
	if err := params.Prompter.PromptApply(&namespacev1.ExportSinkSpec{}, spec, false); err != nil {
		return err
	}
	createSink := wrapCreateOperation(
		params.Cloud.CreateNamespaceExportSink,
		params.OperationHandler,
		func(_ *cloudservice.CreateNamespaceExportSinkResponse) string { return params.SinkName },
	)
	return createSink(ctx, &cloudservice.CreateNamespaceExportSinkRequest{
		Namespace:        params.Namespace,
		Spec:             spec,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func UpdateS3ExportSink(ctx context.Context, params UpdateS3ExportSinkParams) error {
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
	newS3 := proto.Clone(oldSpec.GetS3()).(*sinkv1.S3Spec)
	if params.RoleName != "" {
		newS3.RoleName = params.RoleName
	}
	if params.AwsAccountID != "" {
		newS3.AwsAccountId = params.AwsAccountID
	}
	if params.BucketName != "" {
		newS3.BucketName = params.BucketName
	}
	if params.KmsArn != "" {
		newS3.KmsArn = params.KmsArn
	}
	newSpec := &namespacev1.ExportSinkSpec{
		Name:    params.SinkName,
		Enabled: oldSpec.GetEnabled(),
		S3:      newS3,
	}

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

func ValidateS3ExportSink(ctx context.Context, params ValidateS3ExportSinkParams) error {
	spec := &namespacev1.ExportSinkSpec{
		Name: params.SinkName,
		S3: &sinkv1.S3Spec{
			RoleName:     params.RoleName,
			BucketName:   params.BucketName,
			Region:       params.Region,
			AwsAccountId: params.AwsAccountID,
			KmsArn:       params.KmsArn,
		},
	}
	_, err := params.Cloud.ValidateNamespaceExportSink(ctx, &cloudservice.ValidateNamespaceExportSinkRequest{
		Namespace: params.Namespace,
		Spec:      spec,
	})
	if err != nil {
		return err
	}
	return params.Printer.PrintStructured(
		struct{ Status string }{Status: fmt.Sprintf("Export sink %q configuration is valid.", params.SinkName)},
		printer.StructuredOptions{},
	)
}

func CreateGCSExportSink(ctx context.Context, params CreateGCSExportSinkParams) error {
	spec := &namespacev1.ExportSinkSpec{
		Name:    params.SinkName,
		Enabled: true,
		Gcs: &sinkv1.GCSSpec{
			SaId:         params.SaID,
			BucketName:   params.BucketName,
			GcpProjectId: params.GcpProjectID,
			Region:       params.Region,
		},
	}
	if err := params.Prompter.PromptApply(&namespacev1.ExportSinkSpec{}, spec, false); err != nil {
		return err
	}
	createSink := wrapCreateOperation(
		params.Cloud.CreateNamespaceExportSink,
		params.OperationHandler,
		func(_ *cloudservice.CreateNamespaceExportSinkResponse) string { return params.SinkName },
	)
	return createSink(ctx, &cloudservice.CreateNamespaceExportSinkRequest{
		Namespace:        params.Namespace,
		Spec:             spec,
		AsyncOperationId: params.AsyncOperationID,
	})
}

func UpdateGCSExportSink(ctx context.Context, params UpdateGCSExportSinkParams) error {
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
	newGcs := proto.Clone(oldSpec.GetGcs()).(*sinkv1.GCSSpec)
	if params.SaID != "" {
		newGcs.SaId = params.SaID
	}
	if params.GcpProjectID != "" {
		newGcs.GcpProjectId = params.GcpProjectID
	}
	if params.BucketName != "" {
		newGcs.BucketName = params.BucketName
	}
	newSpec := &namespacev1.ExportSinkSpec{
		Name:    params.SinkName,
		Enabled: oldSpec.GetEnabled(),
		Gcs:     newGcs,
	}

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

func ValidateGCSExportSink(ctx context.Context, params ValidateGCSExportSinkParams) error {
	spec := &namespacev1.ExportSinkSpec{
		Name: params.SinkName,
		Gcs: &sinkv1.GCSSpec{
			SaId:         params.SaID,
			BucketName:   params.BucketName,
			GcpProjectId: params.GcpProjectID,
			Region:       params.Region,
		},
	}
	_, err := params.Cloud.ValidateNamespaceExportSink(ctx, &cloudservice.ValidateNamespaceExportSinkRequest{
		Namespace: params.Namespace,
		Spec:      spec,
	})
	if err != nil {
		return err
	}
	return params.Printer.PrintStructured(
		struct{ Status string }{Status: fmt.Sprintf("Export sink %q configuration is valid.", params.SinkName)},
		printer.StructuredOptions{},
	)
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
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return CreateS3ExportSink(cctx.Context, CreateS3ExportSinkParams{
		Namespace:        c.Namespace,
		SinkName:         c.SinkName,
		RoleName:         roleName,
		BucketName:       c.BucketName,
		Region:           c.Region,
		AwsAccountID:     accountID,
		KmsArn:           c.KmsArn,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
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
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return UpdateS3ExportSink(cctx.Context, UpdateS3ExportSinkParams{
		Namespace:        c.Namespace,
		SinkName:         c.SinkName,
		RoleName:         roleName,
		BucketName:       c.BucketName,
		AwsAccountID:     accountID,
		KmsArn:           c.KmsArn,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudNamespaceExportS3ValidateCommand) run(cctx *CommandContext, _ []string) error {
	roleName, accountID, err := ParseRoleARN(c.RoleArn)
	if err != nil {
		return err
	}
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return ValidateS3ExportSink(cctx.Context, ValidateS3ExportSinkParams{
		Namespace:    c.Namespace,
		SinkName:     c.SinkName,
		RoleName:     roleName,
		BucketName:   c.BucketName,
		Region:       c.Region,
		AwsAccountID: accountID,
		KmsArn:       c.KmsArn,
		Cloud:        cloudClient.CloudService(),
		Printer:      cctx.Printer,
	})
}

func (c *CloudNamespaceExportGcsCreateCommand) run(cctx *CommandContext, _ []string) error {
	saID, projectID, err := ParseServiceAccountEmail(c.ServiceAccountEmail)
	if err != nil {
		return err
	}
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return CreateGCSExportSink(cctx.Context, CreateGCSExportSinkParams{
		Namespace:        c.Namespace,
		SinkName:         c.SinkName,
		SaID:             saID,
		BucketName:       c.BucketName,
		GcpProjectID:     projectID,
		Region:           c.Region,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
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
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return UpdateGCSExportSink(cctx.Context, UpdateGCSExportSinkParams{
		Namespace:        c.Namespace,
		SinkName:         c.SinkName,
		SaID:             saID,
		BucketName:       c.BucketName,
		GcpProjectID:     projectID,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
		Cloud:            cloudClient.CloudService(),
		Prompter:         newPrompter(cctx),
		OperationHandler: NewOperationHandler(cctx, c.AsyncOperationOptions, c.ClientOptions),
	})
}

func (c *CloudNamespaceExportGcsValidateCommand) run(cctx *CommandContext, _ []string) error {
	saID, projectID, err := ParseServiceAccountEmail(c.ServiceAccountEmail)
	if err != nil {
		return err
	}
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	return ValidateGCSExportSink(cctx.Context, ValidateGCSExportSinkParams{
		Namespace:    c.Namespace,
		SinkName:     c.SinkName,
		SaID:         saID,
		BucketName:   c.BucketName,
		GcpProjectID: projectID,
		Region:       c.Region,
		Cloud:        cloudClient.CloudService(),
		Printer:      cctx.Printer,
	})
}
