package namespace

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	"github.com/temporalio/cloud-cli/internal/cert"
)

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

type AddCACertsParams struct {
	Namespace        string
	Certs            []cert.CACert
	ResourceVersion  string
	AsyncOperationID string
}

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

	fingerprints := map[string]struct{}{}
	for _, cert := range existingCerts {
		fingerprints[cert.Fingerprint] = struct{}{}
	}

	for _, cert := range params.Certs {
		if _, ok := fingerprints[cert.Fingerprint]; ok {
			return nil, fmt.Errorf("certificate with fingerprint %q already exists", cert.Fingerprint)
		}
	}

	newBundle := append(existingCerts, params.Certs...)

	var out [][]byte
	for _, cert := range newBundle {
		data, err := base64.StdEncoding.DecodeString(cert.Base64EncodedData)
		if err != nil {
			return nil, err
		}

		out = append(out, data)
	}

	spec := ns.Spec
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

type DeleteCACertsParams struct {
	Namespace        string
	Certs            []cert.CACert
	ResourceVersion  string
	AsyncOperationID string
}

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
