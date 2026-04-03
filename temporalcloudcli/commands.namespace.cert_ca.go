package temporalcloudcli

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"

	"github.com/temporalio/cloud-cli/internal/cert"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceCertCaCreateCommand) run(cctx *CommandContext, _ []string) error {
	newCerts, err := readAndParseCACerts(c.CaCertificateOptions)
	if err != nil {
		return err
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	ns := res.Namespace

	spec, err := buildSpecWithAddedCerts(ns.Spec, newCerts)
	if err != nil {
		return err
	}

	yes, err := cctx.GetPrompter().PromptYes("Create")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
	}

	rv := ns.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateNamespace(cctx, &cloudservice.UpdateNamespaceRequest{
		Namespace:        c.Namespace,
		Spec:             spec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNamespaceCertCaListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}

	certs, err := cert.ParseCACerts(res.Namespace.GetSpec().GetMtlsAuth().GetAcceptedClientCa())
	if err != nil {
		return err
	}

	return cctx.Printer.PrintStructured(certs, printer.StructuredOptions{})
}

func (c *CloudNamespaceCertCaDeleteCommand) run(cctx *CommandContext, _ []string) error {
	certsToRemove, err := readAndParseCACerts(c.CaCertificateOptions)
	if err != nil {
		return err
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	ns := res.Namespace

	spec, err := buildSpecWithDeletedCerts(ns.Spec, certsToRemove)
	if err != nil {
		return err
	}

	yes, err := cctx.GetPrompter().PromptYes("Delete")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting delete.")
	}

	rv := ns.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateNamespace(cctx, &cloudservice.UpdateNamespaceRequest{
		Namespace:        c.Namespace,
		Spec:             spec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

// buildSpecWithAddedCerts returns a new NamespaceSpec with the given certs appended to the existing mTLS CA bundle.
// Certs that already exist (matched by fingerprint) are silently filtered out.
func buildSpecWithAddedCerts(spec *namespacev1.NamespaceSpec, newCerts []cert.CACert) (*namespacev1.NamespaceSpec, error) {
	existingData := spec.GetMtlsAuth().GetAcceptedClientCa()
	existingCerts, err := cert.ParseCACerts(existingData)
	if err != nil {
		return nil, err
	}

	existingFingerprints := map[string]struct{}{}
	for _, c := range existingCerts {
		existingFingerprints[c.Fingerprint] = struct{}{}
	}

	var certsToAdd []cert.CACert
	for _, c := range newCerts {
		if _, exists := existingFingerprints[c.Fingerprint]; !exists {
			certsToAdd = append(certsToAdd, c)
		}
	}

	newBundle := append(existingCerts, certsToAdd...)
	bundleBytes, err := encodeCertBundle(newBundle)
	if err != nil {
		return nil, err
	}

	if spec.MtlsAuth == nil {
		spec.MtlsAuth = &namespacev1.MtlsAuthSpec{}
	}
	spec.MtlsAuth.AcceptedClientCa = bundleBytes
	return spec, nil
}

// buildSpecWithDeletedCerts returns a new NamespaceSpec with the given certs removed from the mTLS CA bundle.
// Certs are matched by fingerprint.
func buildSpecWithDeletedCerts(spec *namespacev1.NamespaceSpec, certsToRemove []cert.CACert) (*namespacev1.NamespaceSpec, error) {
	existingData := spec.GetMtlsAuth().GetAcceptedClientCa()
	existingCerts, err := cert.ParseCACerts(existingData)
	if err != nil {
		return nil, err
	}

	fingerprintsToRemove := map[string]struct{}{}
	for _, c := range certsToRemove {
		fingerprintsToRemove[c.Fingerprint] = struct{}{}
	}

	var remaining []cert.CACert
	for _, existing := range existingCerts {
		if _, ok := fingerprintsToRemove[existing.Fingerprint]; !ok {
			remaining = append(remaining, existing)
		}
	}

	if len(remaining) == 0 {
		spec.MtlsAuth = nil
		return spec, nil
	}

	bundleBytes, err := encodeCertBundle(remaining)
	if err != nil {
		return nil, err
	}
	if spec.MtlsAuth == nil {
		spec.MtlsAuth = &namespacev1.MtlsAuthSpec{}
	}
	spec.MtlsAuth.AcceptedClientCa = bundleBytes
	return spec, nil
}

// encodeCertBundle encodes a slice of CACerts as a joined PEM byte slice.
func encodeCertBundle(certs []cert.CACert) ([]byte, error) {
	var out [][]byte
	for _, c := range certs {
		data, err := base64.StdEncoding.DecodeString(c.Base64EncodedData)
		if err != nil {
			return nil, err
		}
		out = append(out, data)
	}
	return bytes.Join(out, []byte("\n")), nil
}

// readCACertBytes reads raw PEM bytes.
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
