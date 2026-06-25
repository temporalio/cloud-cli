package temporalcloudcli

import (
	"errors"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceApiKeyGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}

	enabled := false
	if res.Namespace.Spec.ApiKeyAuth != nil {
		enabled = res.Namespace.Spec.ApiKeyAuth.Enabled
	}

	result := struct {
		Namespace         string `json:"namespace"`
		ApiKeyAuthEnabled bool   `json:"apiKeyAuthEnabled"`
	}{
		Namespace:         res.Namespace.Namespace,
		ApiKeyAuthEnabled: enabled,
	}
	return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
}

func (c *CloudNamespaceApiKeyEnableCommand) run(cctx *CommandContext, _ []string) error {
	return setApiKeyAuthEnabled(cctx, c.ClientOptions, c.NamespaceOptions, c.ResourceVersionOptions, c.AsyncOperationOptions, true)
}

func (c *CloudNamespaceApiKeyDisableCommand) run(cctx *CommandContext, _ []string) error {
	return setApiKeyAuthEnabled(cctx, c.ClientOptions, c.NamespaceOptions, c.ResourceVersionOptions, c.AsyncOperationOptions, false)
}

func setApiKeyAuthEnabled(cctx *CommandContext, clientOpts ClientOptions, nsOpts NamespaceOptions, rvOpts ResourceVersionOptions, asyncOpts AsyncOperationOptions, enabled bool) error {
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
	if newSpec.ApiKeyAuth == nil {
		newSpec.ApiKeyAuth = &namespacev1.ApiKeyAuthSpec{}
	}
	newSpec.ApiKeyAuth.Enabled = enabled

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
