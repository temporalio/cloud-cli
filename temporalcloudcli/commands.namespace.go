package temporalcloudcli

import (
	"fmt"
	"strings"

	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	operation "go.temporal.io/cloud-sdk/api/operation/v1"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceGetCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))

	n, err := client.getNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	if c.Spec {
		return cctx.Printer.PrintStructured(n.Spec, printer.StructuredOptions{})
	}
	return cctx.Printer.PrintResource(n, printer.PrintResourceOptions{})
}

func (c *CloudNamespaceEditCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))

	ns, err := client.getNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	newSpec := &namespacev1.NamespaceSpec{}
	err = runEditorForJSONEditForProtos(ns.Spec, newSpec)
	if err != nil {
		return err
	}

	err = promptApplyResource(cctx, ns.Spec, newSpec, c.VerboseDiff)
	if err != nil {
		return err
	}

	// Use provided resource version, or fall back to the fetched namespace's resource version.
	resourceVersion := c.ResourceVersion
	if resourceVersion == "" {
		resourceVersion = ns.ResourceVersion
	}

	asyncOp, err := client.updateNamespace(cctx.Context, updateNamespaceParams{
		namespace:        c.Namespace,
		spec:             newSpec,
		asyncOperationID: c.AsyncOperationId,
		resourceVersion:  resourceVersion,
		idempotent:       c.Idempotent,
	})
	if err != nil {
		return err
	}

	// TODO: (gmankes) remove this -- clean up and make shareable
	if asyncOp == nil {
		// Nothing changed (idempotent case)
		result := struct {
			Status    string
			Namespace string
		}{
			Status:    "unchanged",
			Namespace: c.Namespace,
		}
		return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
	}

	// Handle async flag
	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: asyncOp,
			ID:      c.Namespace,
		}, printer.StructuredOptions{})
	}

	// Poll for completion
	poller, err := getPoller(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	return poller.PollAsyncOperation(cctx, asyncOp.Id, c.Namespace)
}

func (c *CloudNamespaceApplyCommand) run(cctx *CommandContext, _ []string) error {
	// Step 1: Load spec from file or inline
	specData, err := loadJSONSpec(c.Spec)
	if err != nil {
		return err
	}

	// Step 2: Parse JSON into NamespaceSpec using protobuf JSON unmarshaling
	spec := &namespacev1.NamespaceSpec{}
	if err := cctx.UnmarshalProtoJSON(specData, spec); err != nil {
		return fmt.Errorf("failed to parse JSON spec: %w", err)
	}

	// Step 3: Create cloud and namespace clients
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	client := newNamespaceClient(withCloudClient(cloudClient))

	// Step 4: Retrieve existing namespace
	var found bool
	existing, err := client.getNamespaceByName(cctx.Context, spec.Name)
	if err != nil && !isNotFoundErr(err) {
		return err
	} else if err == nil {
		found = true
	}

	existingResourceVersion := ""
	var existingSpec *namespacev1.NamespaceSpec
	existingNamespaceIdentifier := ""

	if found {
		existingResourceVersion = existing.ResourceVersion
		existingSpec = existing.Spec
		existingNamespaceIdentifier = existing.Namespace
	}

	// Step 5: Confirm apply if not forced
	err = promptApplyResource(cctx, existingSpec, spec, c.VerboseDiff)
	if err != nil {
		return err
	}

	// Step 6: Apply the namespace (create or update)
	// Use provided resource version, or use fetched version
	resourceVersion := c.ResourceVersion
	if resourceVersion == "" {
		resourceVersion = existingResourceVersion
	}

	var asyncOp *operation.AsyncOperation
	var namespaceID string

	if found {
		// Update existing namespace
		asyncOp, err = client.updateNamespace(cctx.Context, updateNamespaceParams{
			namespace:        existingNamespaceIdentifier,
			spec:             spec,
			asyncOperationID: c.AsyncOperationId,
			idempotent:       c.Idempotent,
			resourceVersion:  resourceVersion,
		})
		if err != nil {
			return fmt.Errorf("failed to update namespace: %w", err)
		}
		namespaceID = existingNamespaceIdentifier
	} else {
		// Create new namespace
		res, err := client.createNamespace(cctx.Context, createNamespaceParams{
			spec:             spec,
			asyncOperationID: c.AsyncOperationId,
		})
		if err != nil {
			return fmt.Errorf("failed to create namespace: %w", err)
		}
		asyncOp = res.asyncOp
		namespaceID = res.Namespace
	}

	// Step 7: Handle result
	if asyncOp == nil {
		// Nothing changed (idempotent case)
		result := struct {
			Status    string
			Namespace string
		}{
			Status:    "unchanged",
			Namespace: namespaceID,
		}
		return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
	}

	// Step 8: Handle async flag
	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: asyncOp,
			ID:      namespaceID,
		}, printer.StructuredOptions{})
	}

	// Step 7: Poll for completion
	poller, err := getPoller(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	return poller.PollAsyncOperation(cctx, asyncOp.Id, namespaceID)
}

func (c *CloudNamespaceDeleteCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))

	yes, err := cctx.promptYes("Delete (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return fmt.Errorf("Aborting delete.")
	}

	asyncOp, err := client.deleteNamespace(cctx.Context, deleteNamespaceParams{
		namespace:        c.Namespace,
		idempotent:       c.Idempotent,
		asyncOperationID: c.AsyncOperationId,
		resourceVersion:  c.ResourceVersion,
	})
	if err != nil {
		return err
	}

	if asyncOp == nil {
		// deleted already (idempotent case)
		result := struct {
			Status    string
			Namespace string
		}{
			Status:    "deleted",
			Namespace: c.Namespace,
		}
		return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
	}

	// Handle async flag
	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: asyncOp,
			ID:      c.Namespace,
		}, printer.StructuredOptions{})
	}

	// Poll for completion
	poller, err := getPoller(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	return poller.PollAsyncOperation(cctx, asyncOp.Id, c.Namespace)
}

func (c *CloudNamespaceListCommand) run(cctx *CommandContext, _ []string) error {
	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	client := newNamespaceClient(withCloudClient(cloudClient))

	namespaces, nextPageToken, err := client.getNamespaces(cctx.Context, getNamespacesParams{
		pageSize:  int32(c.PageSize),
		pageToken: c.PageToken,
		name:      c.Name,
	})
	if err != nil {
		return err
	}

	return cctx.Printer.PrintResourceList(
		struct {
			Namespaces    []*namespacev1.Namespace
			NextPageToken string
		}{
			Namespaces:    namespaces,
			NextPageToken: nextPageToken,
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Namespace", "State", "CreatedTime"},
			SpecFields: []string{"Regions"},
		},
		printer.TableOptions{},
	)
}

func (c *CloudNamespaceCreateCommand) run(cctx *CommandContext, _ []string) error {
	if c.CertificateFilter != "" && c.CertificateFilterFile != "" {
		return fmt.Errorf("cannot specify both --certificate-filter and --certificate-filter-file")
	}

	spec := &namespacev1.NamespaceSpec{
		Name:          c.Name,
		Regions:       c.Region,
		RetentionDays: int32(c.RetentionDays),
		ApiKeyAuth:    &namespacev1.ApiKeyAuthSpec{Enabled: c.ApiKeyAuthEnabled},
		Lifecycle:     &namespacev1.LifecycleSpec{EnableDeleteProtection: c.EnableDeleteProtection},
	}

	// Handle CA certificate
	certBytes, err := readOptionalRawCACertBytes(c.CaCertificateOptions)
	if err != nil {
		return err
	}
	if certBytes != nil {
		if spec.MtlsAuth == nil {
			spec.MtlsAuth = &namespacev1.MtlsAuthSpec{}
		}
		spec.MtlsAuth.AcceptedClientCa = certBytes
	}

	// Handle certificate filter
	if c.CertificateFilter != "" || c.CertificateFilterFile != "" {
		filterSrc := c.CertificateFilter
		if c.CertificateFilterFile != "" {
			filterSrc = "@" + c.CertificateFilterFile
		}
		filterData, err := loadJSONSpec(filterSrc)
		if err != nil {
			return err
		}
		filter := &namespacev1.CertificateFilterSpec{}
		if err := cctx.UnmarshalProtoJSON(filterData, filter); err != nil {
			return fmt.Errorf("failed to parse certificate filter: %w", err)
		}
		if spec.MtlsAuth == nil {
			spec.MtlsAuth = &namespacev1.MtlsAuthSpec{}
		}
		spec.MtlsAuth.CertificateFilters = append(spec.MtlsAuth.CertificateFilters, filter)
	}

	// Handle codec server
	if c.CodecEndpoint != "" {
		spec.CodecServer = &namespacev1.CodecServerSpec{
			Endpoint:                      c.CodecEndpoint,
			PassAccessToken:               c.CodecPassAccessToken,
			IncludeCrossOriginCredentials: c.CodecIncludeCrossOriginCredentials,
		}
	}

	// Handle search attributes
	if len(c.SearchAttribute) > 0 {
		spec.SearchAttributes = make(map[string]namespacev1.NamespaceSpec_SearchAttributeType, len(c.SearchAttribute))
		for _, sa := range c.SearchAttribute {
			name, typStr, ok := strings.Cut(sa, "=")
			if !ok {
				return fmt.Errorf("invalid search attribute format %q: expected 'name=Type'", sa)
			}
			attrType, err := parseSearchAttributeType(typStr)
			if err != nil {
				return err
			}
			spec.SearchAttributes[name] = attrType
		}
	}

	// Handle connectivity rules
	spec.ConnectivityRuleIds = c.ConnectionRuleId

	cloudClient, err := cctx.BuildCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	client := newNamespaceClient(withCloudClient(cloudClient))

	yes, err := cctx.promptYes("Create (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		return fmt.Errorf("Aborting create.")
	}

	res, err := client.createNamespace(cctx.Context, createNamespaceParams{
		spec:             spec,
		asyncOperationID: c.AsyncOperationId,
	})
	if err != nil {
		return err
	}

	if c.Async {
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: res.asyncOp,
			ID:      res.Namespace,
		}, printer.StructuredOptions{})
	}

	poller, err := getPoller(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	return poller.PollAsyncOperation(cctx, res.asyncOp.Id, res.Namespace)
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
