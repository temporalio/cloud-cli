package temporalcloudcli

import (
	"os"

	"github.com/temporalio/cloud-cli/internal/cert"
	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/async"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceCertCaAddCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	namespaceClient := cctx.NamespaceClient
	if namespaceClient == nil {
		namespaceClient = &namespace.Client{
			Cloud: cloudClient.CloudService(),
		}
	}

	certData, err := os.ReadFile(c.CaCertificateFile)
	if err != nil {
		return err
	}

	newCerts, err := cert.ParseCACerts(certData)
	if err != nil {
		return err
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

	poller := cctx.Poller
	if poller == nil {
		poller = &async.Poller{
			Cloud:      cloudClient.CloudService(),
			Printer:    cctx.Printer,
			JSONOutput: cctx.JSONOutput,
		}
	}
	err = poller.Poll(cctx.Context, op.Id, c.Namespace)
	if err != nil {
		return err
	}

	ns, err := namespaceClient.GetNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	return cctx.Printer.PrintStructured(ns.GetSpec().GetMtlsAuth().GetAcceptedClientCa(), printer.StructuredOptions{})
}

func (c *CloudNamespaceCertCaListCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	namespaceClient := namespace.Client{
		Cloud: cloudClient.CloudService(),
	}

	certs, err := namespaceClient.ListCACerts(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	return cctx.Printer.PrintStructured(certs, printer.StructuredOptions{})
}

func (c *CloudNamespaceCertCaRemoveCommand) run(cctx *CommandContext, _ []string) error {
	return nil
}
