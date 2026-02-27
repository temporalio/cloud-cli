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

	newCerts, err := readAndParseCACerts(c.CaCertificateFile, c.CaCertificate)
	if err != nil {
		return err
	}

	addCACerts := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, namespaceClient.AddCACerts)
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

	certsToRemove, err := readAndParseCACerts(c.CaCertificateFile, c.CaCertificate)
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

	deleteCACerts := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, namespaceClient.DeleteCACerts)
	return deleteCACerts(namespace.DeleteCACertsParams{
		Namespace:        c.Namespace,
		Certs:            certsToRemove,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	})
}

// readAndParseCACerts reads certificate data from either a file or base64 string,
// validates the input, and parses the certificates. Returns an error if:
// - Neither flag is provided
// - Both flags are provided
// - The file cannot be read
// - The base64 data is invalid
// - The certificate data cannot be parsed
// - No valid certificates are found
func readAndParseCACerts(certFile, certBase64 string) ([]cert.CACert, error) {
	// Validate that exactly one of the two flags is provided
	if certFile == "" && certBase64 == "" {
		return nil, errors.New("either --ca-certificate-file or --ca-certificate must be provided")
	}
	if certFile != "" && certBase64 != "" {
		return nil, errors.New("cannot specify both --ca-certificate-file and --ca-certificate")
	}

	var certData []byte
	var err error

	if certFile != "" {
		certData, err = os.ReadFile(certFile)
		if err != nil {
			return nil, err
		}
	} else {
		// Decode base64 certificate data
		certData, err = base64.StdEncoding.DecodeString(certBase64)
		if err != nil {
			return nil, errors.New("invalid base64 encoded certificate data")
		}
	}

	certs, err := cert.ParseCACerts(certData)
	if err != nil {
		return nil, err
	}

	if len(certs) == 0 {
		return nil, errors.New("invalid certificate")
	}

	return certs, nil
}
