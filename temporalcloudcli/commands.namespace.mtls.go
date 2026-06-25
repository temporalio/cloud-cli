package temporalcloudcli

import (
	"errors"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/internal/cert"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceMtlsGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}

	var (
		enabled    bool
		caCerts    []cert.CACert
		certFilter []*namespacev1.CertificateFilterSpec
	)
	if mtls := res.Namespace.Spec.MtlsAuth; mtls != nil {
		enabled = mtls.Enabled
		certFilter = mtls.CertificateFilters
		if len(mtls.AcceptedClientCa) > 0 {
			var err error
			caCerts, err = cert.ParseCACerts(mtls.AcceptedClientCa)
			if err != nil {
				return err
			}
		}
	}

	result := struct {
		Namespace          string                               `json:"namespace"`
		MtlsAuthEnabled    bool                                 `json:"mtlsAuthEnabled"`
		CACerts            []cert.CACert                        `json:"caCerts,omitempty"`
		CertificateFilters []*namespacev1.CertificateFilterSpec `json:"certificateFilters,omitempty"`
	}{
		Namespace:          res.Namespace.Namespace,
		MtlsAuthEnabled:    enabled,
		CACerts:            caCerts,
		CertificateFilters: certFilter,
	}
	return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
}

func (c *CloudNamespaceMtlsEnableCommand) run(cctx *CommandContext, _ []string) error {
	return setMtlsAuthEnabled(cctx, c.ClientOptions, c.NamespaceOptions, c.ResourceVersionOptions, c.AsyncOperationOptions, true)
}

func (c *CloudNamespaceMtlsDisableCommand) run(cctx *CommandContext, _ []string) error {
	return setMtlsAuthEnabled(cctx, c.ClientOptions, c.NamespaceOptions, c.ResourceVersionOptions, c.AsyncOperationOptions, false)
}

func setMtlsAuthEnabled(cctx *CommandContext, clientOpts ClientOptions, nsOpts NamespaceOptions, rvOpts ResourceVersionOptions, asyncOpts AsyncOperationOptions, enabled bool) error {
	client, err := cctx.GetCloudClient(clientOpts)
	if err != nil {
		return err
	}
	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: nsOpts.Namespace})
	if err != nil {
		return err
	}

	ns := res.Namespace
	newSpec := proto.Clone(ns.Spec).(*namespacev1.NamespaceSpec)
	if newSpec.MtlsAuth == nil {
		newSpec.MtlsAuth = &namespacev1.MtlsAuthSpec{}
	}
	newSpec.MtlsAuth.Enabled = enabled

	yes, err := cctx.GetPrompter().PromptApply(ns.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting.")
	}

	rv := ns.ResourceVersion
	if rvOpts.ResourceVersion != "" {
		rv = rvOpts.ResourceVersion
	}
	resp, err := client.UpdateNamespace(cctx, &cloudservice.UpdateNamespaceRequest{
		Namespace:        nsOpts.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: asyncOpts.AsyncOperationId,
	})
	return cctx.GetPoller(client, asyncOpts).HandleUpdateOperation(cctx, resp, err)
}
