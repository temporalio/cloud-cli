package temporalcloudcli

import (
	"encoding/base64"
	"errors"
	"fmt"
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

	newCerts, err := readAndParseCACerts(c.CaCertificateOptions)
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

	certsToRemove, err := readAndParseCACerts(c.CaCertificateOptions)
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

// readCACertBytes reads raw PEM bytes when cert flags are optional.
// Returns nil (no error) if neither flag is provided.
// Returns an error if both flags are provided.
func readCACertBytes(opts CaCertificateOptions) ([]byte, error) {
	if opts.CaCertificate != "" && opts.CaCertificateFile != "" {
		return nil, errors.New("cannot specify both --ca-certificate and --ca-certificate-file")
	}
	if opts.CaCertificateFile != "" {
		data, err := os.ReadFile(opts.CaCertificateFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate file: %w", err)
		}
		return data, nil
	}
	if opts.CaCertificate != "" {
		data, err := base64.StdEncoding.DecodeString(opts.CaCertificate)
		if err != nil {
			return nil, errors.New("invalid base64 encoded certificate data")
		}
		return data, nil
	}
	return nil, nil
}

// readAndParseCACerts requires exactly one cert flag to be provided.
// Returns an error if:
// - Neither flag is provided
// - Both flags are provided
// - The file cannot be read
// - The base64 data is invalid
// - The certificate data cannot be parsed
// - No valid certificates are found
func readAndParseCACerts(opts CaCertificateOptions) ([]cert.CACert, error) {
	if opts.CaCertificate == "" && opts.CaCertificateFile == "" {
		return nil, errors.New("either --ca-certificate-file or --ca-certificate must be provided")
	}
	certData, err := readCACertBytes(opts)
	if err != nil {
		return nil, err
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
