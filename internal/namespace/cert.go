package namespace

import (
	"context"

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

	existingData := ns.GetSpec().GetMtlsAuth().GetAcceptedClientCa()
	existingCerts, err := cert.ParseCACerts(existingData)
	if err != nil {
		return nil, err
	}

	bundleBytes, err := cert.EncodeCACerts(cert.Add(existingCerts, params.Certs))
	if err != nil {
		return nil, err
	}

	spec := ns.GetSpec()
	// Ensure MtlsAuth is initialized before accessing its fields
	if spec.MtlsAuth == nil {
		spec.MtlsAuth = &namespacev1.MtlsAuthSpec{Enabled: true}
	}
	spec.MtlsAuth.AcceptedClientCa = bundleBytes

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

	existingData := ns.GetSpec().GetMtlsAuth().GetAcceptedClientCa()
	existingCerts, err := cert.ParseCACerts(existingData)
	if err != nil {
		return nil, err
	}

	newBundle := cert.Remove(existingCerts, params.Certs)

	bundleBytes, err := cert.EncodeCACerts(newBundle)
	if err != nil {
		return nil, err
	}

	spec := ns.GetSpec()
	if len(newBundle) == 0 {
		spec.MtlsAuth = nil
	} else {
		// Ensure MtlsAuth is initialized before accessing its fields
		if spec.MtlsAuth == nil {
			spec.MtlsAuth = &namespacev1.MtlsAuthSpec{Enabled: true}
		}
		spec.MtlsAuth.AcceptedClientCa = bundleBytes
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
