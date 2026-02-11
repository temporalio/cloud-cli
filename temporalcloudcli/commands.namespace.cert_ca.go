package temporalcloudcli

import (
	"bytes"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceCertCaAddCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	certData, err := os.ReadFile(c.CaCertificateFile)
	if err != nil {
		return err
	}

	certs, err := parseCACerts(certData)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))
	ns, err := client.getNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	existingBundle := ns.Spec.GetMtlsAuth().GetAcceptedClientCa()
	existingData, err := base64.StdEncoding.DecodeString(string(existingBundle))
	if err != nil {
		return err
	}

	existingCerts, err := parseCACerts(existingData)
	if err != nil {
		return err
	}

	fingerprints := map[string]struct{}{}
	for _, cert := range existingCerts {
		fingerprints[cert.Fingerprint] = struct{}{}
	}

	for _, cert := range certs {
		if _, ok := fingerprints[cert.Fingerprint]; ok {
			return fmt.Errorf("certificate with fingerprint %q already exists", cert.Fingerprint)
		}
	}

	newBundle := append(existingCerts, certs...)

	var out [][]byte
	for _, cert := range newBundle {
		data, err := base64.StdEncoding.DecodeString(cert.Base64EncodedData)
		if err != nil {
			return err
		}

		out = append(out, data)
	}

	spec := ns.Spec
	spec.MtlsAuth.AcceptedClientCa = bytes.Join(out, []byte("\n"))

	resourceVersion := ns.ResourceVersion
	if c.ResourceVersion != "" {
		resourceVersion = c.ResourceVersion
	}

	updateParams := updateNamespaceParams{
		namespace:        c.Namespace,
		spec:             spec,
		resourceVersion:  resourceVersion,
		idempotent:       c.Idempotent,
		asyncOperationID: c.AsyncOperationId,
	}

	res, err := client.updateNamespace(cctx.Context, updateParams)
	if err != nil {
		return err
	}

	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: res,
			ID:      c.Namespace,
		}, printer.StructuredOptions{})
	}
	err = PollAsyncOperation(cctx, cloudClient, res.Id, c.Namespace)
	if err != nil {
		return err
	}

	return cctx.Printer.PrintStructured(certs, printer.StructuredOptions{})
}

func parseCACerts(data []byte) ([]CACert, error) {
	var der []byte
	var blocks [][]byte
	for {
		var block *pem.Block
		var rem []byte
		block, rem = pem.Decode(data)
		if block == nil {
			break
		}

		der = append(der, block.Bytes...)

		blocks = append(blocks, []byte(strings.TrimSpace(string(data[:len(data)-len(rem)]))))
		data = rem
	}

	certs, err := x509.ParseCertificates(der)
	if err != nil {
		return nil, err
	}

	result := make([]CACert, len(certs))
	for i, cert := range certs {
		sum := sha1.Sum(certs[i].Raw)
		result[i] = CACert{
			Fingerprint:       strings.ToLower(hex.EncodeToString(sum[:])),
			Issuer:            cert.Issuer.String(),
			Subject:           cert.Subject.String(),
			NotBefore:         cert.NotBefore,
			NotAfter:          cert.NotAfter,
			Base64EncodedData: base64.StdEncoding.EncodeToString(blocks[i]),
		}
	}

	return result, nil
}

type CACert struct {
	Fingerprint       string
	Issuer            string
	Subject           string
	NotBefore         time.Time
	NotAfter          time.Time
	Base64EncodedData string
}

func (c *CloudNamespaceCertCaListCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))
	ns, err := client.getNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	bundle := ns.Spec.GetMtlsAuth().GetAcceptedClientCa()

	certs, err := parseCACerts(bundle)
	if err != nil {
		return err
	}

	return cctx.Printer.PrintStructured(certs, printer.StructuredOptions{})
}

func (c *CloudNamespaceCertCaRemoveCommand) run(cctx *CommandContext, _ []string) error {
	return nil
}
