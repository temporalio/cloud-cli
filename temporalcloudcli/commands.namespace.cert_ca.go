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

	params := namespace.AddCACertsParams{
		Namespace:        c.Namespace,
		Certs:            newCerts,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	}
	op, err := namespaceClient.AddCACerts(cctx.Context, params)
	if err != nil {
		if isNothingChangedErr(c.Idempotent, err) {
			result := struct {
				Status    string
				Namespace string
			}{
				Status:    "unchanged",
				Namespace: c.Namespace,
			}
			return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
		}

		return err
	}

	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: op,
			ID:      c.Namespace,
		}, printer.StructuredOptions{})
	}

	poller, err := getPoller(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	return poller.Poll(cctx.Context, op.Id, c.Namespace)
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

	params := namespace.DeleteCACertsParams{
		Namespace:        c.Namespace,
		Certs:            certsToRemove,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	}
	op, err := namespaceClient.DeleteCACerts(cctx.Context, params)
	if err != nil {
		if isNothingChangedErr(c.Idempotent, err) {
			result := struct {
				Status    string
				Namespace string
			}{
				Status:    "unchanged",
				Namespace: c.Namespace,
			}
			return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
		}

		return err
	}

	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: op,
			ID:      c.Namespace,
		}, printer.StructuredOptions{})
	}

	poller, err := getPoller(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	return poller.Poll(cctx.Context, op.Id, c.Namespace)
}
