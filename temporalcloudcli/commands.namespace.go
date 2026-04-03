package temporalcloudcli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/temporalio/cli/cliext"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	if c.Spec {
		return cctx.Printer.PrintStructured(res.Namespace.Spec, printer.StructuredOptions{})
	}
	return cctx.Printer.PrintResource(res.Namespace, printer.PrintResourceOptions{})
}

func (c *CloudNamespaceEditCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	ns := res.Namespace

	edited, err := cctx.GetEditor().EditProto(ns.Spec)
	if err != nil {
		return err
	}
	newSpec := edited.(*namespacev1.NamespaceSpec)

	yes, err := cctx.GetPrompter().PromptApply(ns.Spec, newSpec, c.VerboseDiff)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting edit.")
	}

	rv := ns.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	asyncOpts := AsyncOperationOptions{
		Idempotent:       c.Idempotent,
		Async:            c.Async,
		AsyncOperationId: c.AsyncOperationId,
		PollInterval:     cliext.FlagDuration(time.Second),
	}
	resp, err := client.UpdateNamespace(cctx, &cloudservice.UpdateNamespaceRequest{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, asyncOpts).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNamespaceApplyCommand) run(cctx *CommandContext, _ []string) error {
	specData, err := loadJSONSpec(c.Spec)
	if err != nil {
		return err
	}

	spec := &namespacev1.NamespaceSpec{}
	if err := cctx.UnmarshalProtoJSON(specData, spec); err != nil {
		return fmt.Errorf("failed to parse JSON spec: %w", err)
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	existing, err := getNamespaceBySpecName(cctx, client, spec.Name)
	if err != nil && !isNotFoundErr(err) {
		return err
	}

	asyncOpts := AsyncOperationOptions{
		Idempotent:       c.Idempotent,
		Async:            c.Async,
		AsyncOperationId: c.AsyncOperationId,
		PollInterval:     cliext.FlagDuration(time.Second),
	}

	var existingSpec *namespacev1.NamespaceSpec
	if existing != nil {
		existingSpec = existing.Spec
	}

	yes, err := cctx.GetPrompter().PromptApply(existingSpec, spec, c.VerboseDiff)
	if err != nil {
		return err
	}
	if !yes {
		return nil
	}

	if existing != nil {
		rv := existing.ResourceVersion
		if c.ResourceVersion != "" {
			rv = c.ResourceVersion
		}
		resp, err := client.UpdateNamespace(cctx, &cloudservice.UpdateNamespaceRequest{
			Namespace:        existing.Namespace,
			Spec:             spec,
			ResourceVersion:  rv,
			AsyncOperationId: c.AsyncOperationId,
		})
		return cctx.GetPoller(client, asyncOpts).HandleUpdateOperation(cctx, resp, err)
	}

	resp, err := client.CreateNamespace(cctx, &cloudservice.CreateNamespaceRequest{
		Spec:             spec,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, asyncOpts).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudNamespaceDeleteCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
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

	asyncOpts := AsyncOperationOptions{
		Idempotent:       c.Idempotent,
		Async:            c.Async,
		AsyncOperationId: c.AsyncOperationId,
		PollInterval:     cliext.FlagDuration(time.Second),
	}

	getRes, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return cctx.GetPoller(client, asyncOpts).HandleDeleteOperation(cctx, nil, err)
	}

	rv := getRes.Namespace.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.DeleteNamespace(cctx, &cloudservice.DeleteNamespaceRequest{
		Namespace:        c.Namespace,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, asyncOpts).HandleDeleteOperation(cctx, resp, err)
}

func (c *CloudNamespaceListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespaces(cctx, &cloudservice.GetNamespacesRequest{
		PageSize:  int32(c.PageSize),
		PageToken: c.PageToken,
		Name:      c.Name,
	})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResourceList(
		struct {
			Namespaces    []*namespacev1.Namespace
			NextPageToken string
		}{
			Namespaces:    res.Namespaces,
			NextPageToken: res.NextPageToken,
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Namespace", "State", "CreatedTime"},
			SpecFields: []string{"Regions"},
		},
		printer.TableOptions{},
	)
}

func (c *CloudNamespaceCreateCommand) run(cctx *CommandContext, _ []string) error {
	certBytes, err := readCACertBytes(c.CaCertificateOptions)
	if err != nil {
		return err
	}

	var certFilters []*namespacev1.CertificateFilterSpec
	for _, raw := range c.CertificateFilterOptions.CertificateFilter {
		filterData, err := loadJSONSpec(raw)
		if err != nil {
			return err
		}
		filter := &namespacev1.CertificateFilterSpec{}
		if err := cctx.UnmarshalProtoJSON(filterData, filter); err != nil {
			return fmt.Errorf("failed to parse certificate filter: %w", err)
		}
		certFilters = append(certFilters, filter)
	}
	if c.CertificateFilterOptions.CertificateFilterFile != "" {
		filterData, err := loadJSONSpec("@" + c.CertificateFilterOptions.CertificateFilterFile)
		if err != nil {
			return err
		}
		filter := &namespacev1.CertificateFilterSpec{}
		if err := cctx.UnmarshalProtoJSON(filterData, filter); err != nil {
			return fmt.Errorf("failed to parse certificate filter file: %w", err)
		}
		certFilters = append(certFilters, filter)
	}

	var searchAttrs map[string]namespacev1.NamespaceSpec_SearchAttributeType
	if len(c.SearchAttribute) > 0 {
		searchAttrs = make(map[string]namespacev1.NamespaceSpec_SearchAttributeType, len(c.SearchAttribute))
		for _, sa := range c.SearchAttribute {
			saName, typStr, ok := strings.Cut(sa, "=")
			if !ok {
				return fmt.Errorf("invalid search attribute format %q: expected 'name=Type'", sa)
			}
			attrType, err := parseSearchAttributeType(typStr)
			if err != nil {
				return err
			}
			searchAttrs[saName] = attrType
		}
	}

	spec := &namespacev1.NamespaceSpec{
		Name:                c.Name,
		Regions:             c.Region,
		RetentionDays:       int32(c.RetentionDays),
		ApiKeyAuth:          &namespacev1.ApiKeyAuthSpec{Enabled: c.ApiKeyAuthEnabled},
		Lifecycle:           &namespacev1.LifecycleSpec{EnableDeleteProtection: c.EnableDeleteProtection},
		ConnectivityRuleIds: c.ConnectionRuleId,
		SearchAttributes:    searchAttrs,
	}

	if len(certBytes) > 0 {
		spec.MtlsAuth = &namespacev1.MtlsAuthSpec{AcceptedClientCa: certBytes}
	}
	if len(certFilters) > 0 {
		if spec.MtlsAuth == nil {
			spec.MtlsAuth = &namespacev1.MtlsAuthSpec{}
		}
		spec.MtlsAuth.CertificateFilters = certFilters
	}
	if c.CodecEndpoint != "" {
		spec.CodecServer = &namespacev1.CodecServerSpec{
			Endpoint:                      c.CodecEndpoint,
			PassAccessToken:               c.CodecPassAccessToken,
			IncludeCrossOriginCredentials: c.CodecIncludeCrossOriginCredentials,
		}
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	yes, err := cctx.GetPrompter().PromptApply(&namespacev1.NamespaceSpec{}, spec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
	}

	resp, err := client.CreateNamespace(cctx, &cloudservice.CreateNamespaceRequest{
		Spec:             spec,
		AsyncOperationId: c.AsyncOperationOptions.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

// getNamespaceBySpecName finds a namespace by its spec name field (not the namespace ID).
// Returns nil, notFoundErr if no namespace with that name exists.
func getNamespaceBySpecName(cctx *CommandContext, client cloudservice.CloudServiceClient, name string) (*namespacev1.Namespace, error) {
	res, err := client.GetNamespaces(cctx, &cloudservice.GetNamespacesRequest{Name: name})
	if err != nil {
		return nil, err
	}
	switch len(res.Namespaces) {
	case 0:
		return nil, status.Errorf(codes.NotFound, "namespace not found")
	case 1:
		return res.Namespaces[0], nil
	default:
		return nil, fmt.Errorf("multiple namespaces match namespace name: %s", name)
	}
}

func getNamespaceClient(cctx *CommandContext, opts ClientOptions) (NamespaceClient, error) {
	if cctx.NamespaceClient != nil {
		return cctx.NamespaceClient, nil
	}

	cloudClient, err := cctx.BuildCloudClient(opts)
	if err != nil {
		return nil, err
	}
	return namespace.NewClient(cloudClient.CloudService()), nil
}
