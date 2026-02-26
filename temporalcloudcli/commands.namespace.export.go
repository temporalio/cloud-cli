package temporalcloudcli

import (
	"errors"
	"fmt"

	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceExportGetCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	sink, err := namespaceClient.GetExportSink(cctx.Context, c.Namespace, c.SinkName)
	if err != nil {
		return err
	}

	result := struct {
		Namespace string
		Name      string
		Spec      *namespacev1.ExportSinkSpec
	}{
		Namespace: c.Namespace,
		Name:      sink.Name,
		Spec:      sink.Spec,
	}
	return cctx.Printer.PrintResource(result, printer.PrintResourceOptions{})
}

func (c *CloudNamespaceExportListCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	sinks, err := namespaceClient.ListExportSinks(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	return cctx.Printer.PrintResourceList(
		struct {
			Sinks []*namespacev1.ExportSink
		}{
			Sinks: sinks,
		},
		printer.PrintResourceOptions{
			Fields: []string{"Name", "State"},
		},
		printer.TableOptions{},
	)
}

func (c *CloudNamespaceExportDeleteCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	yes, err := cctx.promptYes("Delete (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting delete.")
	}

	deleteExportSink := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.SinkName, c.ClientOptions, namespaceClient.DeleteExportSink)
	return deleteExportSink(namespace.DeleteExportSinkParams{
		Namespace:        c.Namespace,
		SinkName:         c.SinkName,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	})
}

func (c *CloudNamespaceExportEnableCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	yes, err := cctx.promptYes("Enable (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting enable.")
	}

	enableExportSink := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.SinkName, c.ClientOptions, namespaceClient.EnableExportSink)
	return enableExportSink(namespace.EnableExportSinkParams{
		Namespace:        c.Namespace,
		SinkName:         c.SinkName,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	})
}

func (c *CloudNamespaceExportDisableCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	yes, err := cctx.promptYes("Disable (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting disable.")
	}

	disableExportSink := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.SinkName, c.ClientOptions, namespaceClient.DisableExportSink)
	return disableExportSink(namespace.DisableExportSinkParams{
		Namespace:        c.Namespace,
		SinkName:         c.SinkName,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	})
}

func (c *CloudNamespaceExportS3CreateCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	createS3ExportSink := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.SinkName, c.ClientOptions, namespaceClient.CreateS3ExportSink)
	return createS3ExportSink(namespace.CreateS3ExportSinkParams{
		Namespace:        c.Namespace,
		SinkName:         c.SinkName,
		RoleName:         c.RoleName,
		BucketName:       c.BucketName,
		Region:           c.Region,
		AwsAccountID:     c.AwsAccountId,
		KmsArn:           c.KmsArn,
		AsyncOperationID: c.AsyncOperationId,
	})
}

func (c *CloudNamespaceExportS3UpdateCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	updateS3ExportSink := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.SinkName, c.ClientOptions, namespaceClient.UpdateS3ExportSink)
	return updateS3ExportSink(namespace.UpdateS3ExportSinkParams{
		Namespace:        c.Namespace,
		SinkName:         c.SinkName,
		RoleName:         c.RoleName,
		BucketName:       c.BucketName,
		Region:           c.Region,
		AwsAccountID:     c.AwsAccountId,
		KmsArn:           c.KmsArn,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	})
}

func (c *CloudNamespaceExportS3ValidateCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	if err := namespaceClient.ValidateS3ExportSink(cctx.Context, namespace.ValidateS3ExportSinkParams{
		Namespace:    c.Namespace,
		SinkName:     c.SinkName,
		RoleName:     c.RoleName,
		BucketName:   c.BucketName,
		Region:       c.Region,
		AwsAccountID: c.AwsAccountId,
		KmsArn:       c.KmsArn,
	}); err != nil {
		return err
	}

	return cctx.Printer.PrintStructured(
		struct{ Status string }{Status: fmt.Sprintf("Export sink %q configuration is valid.", c.SinkName)},
		printer.StructuredOptions{},
	)
}

func (c *CloudNamespaceExportGcsCreateCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	createGCSExportSink := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.SinkName, c.ClientOptions, namespaceClient.CreateGCSExportSink)
	return createGCSExportSink(namespace.CreateGCSExportSinkParams{
		Namespace:        c.Namespace,
		SinkName:         c.SinkName,
		SaID:             c.SaId,
		BucketName:       c.BucketName,
		GcpProjectID:     c.GcpProjectId,
		Region:           c.Region,
		AsyncOperationID: c.AsyncOperationId,
	})
}

func (c *CloudNamespaceExportGcsUpdateCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	updateGCSExportSink := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.SinkName, c.ClientOptions, namespaceClient.UpdateGCSExportSink)
	return updateGCSExportSink(namespace.UpdateGCSExportSinkParams{
		Namespace:        c.Namespace,
		SinkName:         c.SinkName,
		SaID:             c.SaId,
		BucketName:       c.BucketName,
		GcpProjectID:     c.GcpProjectId,
		Region:           c.Region,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	})
}

func (c *CloudNamespaceExportGcsValidateCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	if err := namespaceClient.ValidateGCSExportSink(cctx.Context, namespace.ValidateGCSExportSinkParams{
		Namespace:    c.Namespace,
		SinkName:     c.SinkName,
		SaID:         c.SaId,
		BucketName:   c.BucketName,
		GcpProjectID: c.GcpProjectId,
		Region:       c.Region,
	}); err != nil {
		return err
	}

	return cctx.Printer.PrintStructured(
		struct{ Status string }{Status: fmt.Sprintf("Export sink %q configuration is valid.", c.SinkName)},
		printer.StructuredOptions{},
	)
}
