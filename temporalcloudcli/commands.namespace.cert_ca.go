package temporalcloudcli

import (
	"errors"
	"os"

	"github.com/temporalio/cloud-cli/internal/cert"
	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceCertCaAddCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	certData, err := os.ReadFile(c.CaCertificateFile)
	if err != nil {
		return err
	}

	newCerts, err := cert.ParseCACerts(certData)
	if err != nil {
		return err
	}

	if len(newCerts) == 0 {
		return errors.New("invalid certificate")
	}

	addCACerts := wrapAsyncOperation(cctx, c.ResourceModifyOptions, c.Namespace, c.ClientOptions, namespaceClient.AddCACerts)
	return addCACerts(namespace.AddCACertsParams{
		Namespace:        c.Namespace,
		Certs:            newCerts,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	})
}

func (c *CloudNamespaceCertCaListCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	certs, err := namespaceClient.ListCACerts(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	return cctx.Printer.PrintStructured(certs, printer.StructuredOptions{})
}

func (c *CloudNamespaceCertCaDeleteCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	certData, err := os.ReadFile(c.CaCertificateFile)
	if err != nil {
		return err
	}

	certsToRemove, err := cert.ParseCACerts(certData)
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

	deleteCACerts := wrapAsyncOperation(cctx, c.ResourceModifyOptions, c.Namespace, c.ClientOptions, namespaceClient.DeleteCACerts)
	return deleteCACerts(namespace.DeleteCACertsParams{
		Namespace:        c.Namespace,
		Certs:            certsToRemove,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	})
}
