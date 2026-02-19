package temporalcloudcli

import (
	"encoding/base64"
	"errors"
	"os"

	"github.com/temporalio/cloud-cli/internal/cert"
	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceCertCaCreateCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	// Validate that exactly one of the two flags is provided
	if c.CaCertificateFile == "" && c.CaCertificate == "" {
		return errors.New("either --ca-certificate-file or --ca-certificate must be provided")
	}
	if c.CaCertificateFile != "" && c.CaCertificate != "" {
		return errors.New("cannot specify both --ca-certificate-file and --ca-certificate")
	}

	var certData []byte
	if c.CaCertificateFile != "" {
		certData, err = os.ReadFile(c.CaCertificateFile)
		if err != nil {
			return err
		}
	} else {
		// Decode base64 certificate data
		certData, err = base64.StdEncoding.DecodeString(c.CaCertificate)
		if err != nil {
			return errors.New("invalid base64 encoded certificate data")
		}
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

	// Validate that exactly one of the two flags is provided
	if c.CaCertificateFile == "" && c.CaCertificate == "" {
		return errors.New("either --ca-certificate-file or --ca-certificate must be provided")
	}
	if c.CaCertificateFile != "" && c.CaCertificate != "" {
		return errors.New("cannot specify both --ca-certificate-file and --ca-certificate")
	}

	var certData []byte
	if c.CaCertificateFile != "" {
		certData, err = os.ReadFile(c.CaCertificateFile)
		if err != nil {
			return err
		}
	} else {
		// Decode base64 certificate data
		certData, err = base64.StdEncoding.DecodeString(c.CaCertificate)
		if err != nil {
			return errors.New("invalid base64 encoded certificate data")
		}
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
