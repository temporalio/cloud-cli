package namespace

import (
	"bytes"
	"context"
	"encoding/base64"

	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	"github.com/temporalio/cloud-cli/internal/cert"
)

// ListCACerts retrieves all CA certificates configured for mTLS authentication
// on the specified namespace.
func (c *Client) ListCACerts(ctx context.Context, name string) ([]cert.CACert, error) {
	ns, err := c.GetNamespace(ctx, name)
	if err != nil {
		return nil, err
	}

	certs, err := cert.ParseCACerts(ns.GetSpec().GetMtlsAuth().GetAcceptedClientCa())
	if err != nil {
		return nil, err
	}

	return certs, nil
}

// AddCACertsParams contains parameters for adding CA certificates to a namespace.
type AddCACertsParams struct {
	Namespace        string
	Certs            []cert.CACert
	ResourceVersion  string
	AsyncOperationID string
}

// AddCACerts adds CA certificates to the namespace's mTLS authentication configuration.
// Certificates that already exist (matched by fingerprint) are silently filtered out.
// If all certificates already exist, the server will return a "nothing to change" error.
func (c *Client) AddCACerts(ctx context.Context, params AddCACertsParams) (*operation.AsyncOperation, error) {
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	existingData := ns.Spec.GetMtlsAuth().GetAcceptedClientCa()
	existingCerts, err := cert.ParseCACerts(existingData)
	if err != nil {
		return nil, err
	}

	// Build a map of existing certificate fingerprints
	existingFingerprints := map[string]struct{}{}
	for _, cert := range existingCerts {
		existingFingerprints[cert.Fingerprint] = struct{}{}
	}

	// Filter out certificates that already exist (for idempotent behavior)
	var certsToAdd []cert.CACert
	for _, cert := range params.Certs {
		if _, exists := existingFingerprints[cert.Fingerprint]; !exists {
			certsToAdd = append(certsToAdd, cert)
		}
	}

	// Build the new certificate bundle
	newBundle := append(existingCerts, certsToAdd...)

	var out [][]byte
	for _, cert := range newBundle {
		data, err := base64.StdEncoding.DecodeString(cert.Base64EncodedData)
		if err != nil {
			return nil, err
		}

		out = append(out, data)
	}

	spec := ns.Spec
	// Ensure MtlsAuth is initialized before accessing its fields
	if spec.MtlsAuth == nil {
		spec.MtlsAuth = &namespacev1.MtlsAuthSpec{}
	}
	spec.MtlsAuth.AcceptedClientCa = bytes.Join(out, []byte("\n"))

	resourceVersion := ns.ResourceVersion
	if params.ResourceVersion != "" {
		resourceVersion = params.ResourceVersion
	}

	updateParams := UpdateNamespaceParams{
		Namespace:        params.Namespace,
		Spec:             spec,
		ResourceVersion:  resourceVersion,
		AsyncOperationID: params.AsyncOperationID,
	}
	return c.UpdateNamespace(ctx, updateParams)
}

// DeleteCACertsParams contains parameters for deleting CA certificates from a namespace.
type DeleteCACertsParams struct {
	Namespace        string
	Certs            []cert.CACert
	ResourceVersion  string
	AsyncOperationID string
}

// DeleteCACerts removes CA certificates from the namespace's mTLS authentication configuration.
// Certificates are matched by fingerprint. If removing all certificates, the mTLS configuration
// is set to nil.
func (c *Client) DeleteCACerts(ctx context.Context, params DeleteCACertsParams) (*operation.AsyncOperation, error) {
	ns, err := c.GetNamespace(ctx, params.Namespace)
	if err != nil {
		return nil, err
	}

	existingData := ns.Spec.GetMtlsAuth().GetAcceptedClientCa()
	existingCerts, err := cert.ParseCACerts(existingData)
	if err != nil {
		return nil, err
	}

	fingerprintsToRemove := map[string]struct{}{}
	for _, cert := range params.Certs {
		fingerprintsToRemove[cert.Fingerprint] = struct{}{}
	}

	var newBundle []cert.CACert
	for _, existing := range existingCerts {
		if _, ok := fingerprintsToRemove[existing.Fingerprint]; ok {
			continue
		}

		newBundle = append(newBundle, existing)
	}

	var out [][]byte
	for _, cert := range newBundle {
		data, err := base64.StdEncoding.DecodeString(cert.Base64EncodedData)
		if err != nil {
			return nil, err
		}

		out = append(out, data)
	}

	spec := ns.Spec
	if len(out) == 0 {
		spec.MtlsAuth = nil
	} else {
		// Ensure MtlsAuth is initialized before accessing its fields
		if spec.MtlsAuth == nil {
			spec.MtlsAuth = &namespacev1.MtlsAuthSpec{}
		}
		spec.MtlsAuth.AcceptedClientCa = bytes.Join(out, []byte("\n"))
	}

	resourceVersion := ns.ResourceVersion
	if params.ResourceVersion != "" {
		resourceVersion = params.ResourceVersion
	}

	updateParams := UpdateNamespaceParams{
		Namespace:        params.Namespace,
		Spec:             spec,
		ResourceVersion:  resourceVersion,
		AsyncOperationID: params.AsyncOperationID,
	}
	return c.UpdateNamespace(ctx, updateParams)
}
