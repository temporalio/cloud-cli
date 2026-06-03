package temporalcloudcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	commonv1 "go.temporal.io/api/common/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	nexusv1 "go.temporal.io/cloud-sdk/api/nexus/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

// resolveNexusDescription resolves the endpoint description from --description or --description-file.
// The two flags are mutually exclusive. Returns nil if neither is set.
func resolveNexusDescription(description, descriptionFile string) (*commonv1.Payload, error) {
	if description != "" && descriptionFile != "" {
		return nil, errors.New("--description and --description-file are mutually exclusive")
	}
	if descriptionFile != "" {
		data, err := os.ReadFile(descriptionFile)
		if err != nil {
			return nil, fmt.Errorf("failed reading description file %q: %w", descriptionFile, err)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("empty description file: %q", descriptionFile)
		}
		description = string(data)
	}
	if description == "" {
		return nil, nil
	}
	jsonData, err := json.Marshal(description)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal description: %w", err)
	}
	return &commonv1.Payload{
		Metadata: map[string][]byte{"encoding": []byte("json/plain")},
		Data:     jsonData,
	}, nil
}

func (c *CloudNexusEndpointListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetNexusEndpoints(cctx, &cloudservice.GetNexusEndpointsRequest{
		PageSize:  int32(c.PageSize),
		PageToken: c.PageToken,
	})
	if err != nil {
		return err
	}

	return cctx.Printer.PrintResourceList(
		struct {
			Endpoints     []*nexusv1.Endpoint
			NextPageToken string
		}{
			Endpoints:     res.Endpoints,
			NextPageToken: res.NextPageToken,
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Id", "State"},
			SpecFields: []string{"Name", "Description"},
		},
		printer.TableOptions{},
	)
}

func (c *CloudNexusEndpointGetCommand) run(cctx *CommandContext, _ []string) error {
	if c.Name == "" && c.Id == "" {
		return errors.New("either --name or --id is required")
	}
	if c.Name != "" && c.Id != "" {
		return errors.New("--name and --id are mutually exclusive")
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	var endpoint *nexusv1.Endpoint
	if c.Id != "" {
		res, err := client.GetNexusEndpoint(cctx, &cloudservice.GetNexusEndpointRequest{
			EndpointId: c.Id,
		})
		if err != nil {
			return err
		}
		endpoint = res.Endpoint
	} else {
		endpoint, err = getNexusEndpointByName(cctx, client, c.Name)
		if err != nil {
			return err
		}
	}

	return cctx.Printer.PrintResource(endpoint, printer.PrintResourceOptions{})
}

func (c *CloudNexusEndpointCreateCommand) run(cctx *CommandContext, _ []string) error {
	desc, err := resolveNexusDescription(c.Description, c.DescriptionFile)
	if err != nil {
		return err
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	policySpecs := make([]*nexusv1.EndpointPolicySpec, len(c.AllowNamespace))
	for i, ns := range c.AllowNamespace {
		policySpecs[i] = &nexusv1.EndpointPolicySpec{
			Variant: &nexusv1.EndpointPolicySpec_AllowedCloudNamespacePolicySpec{
				AllowedCloudNamespacePolicySpec: &nexusv1.AllowedCloudNamespacePolicySpec{
					NamespaceId: ns,
				},
			},
		}
	}

	yes, err := cctx.GetPrompter().PromptYes("Create")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
	}

	resp, err := client.CreateNexusEndpoint(cctx, &cloudservice.CreateNexusEndpointRequest{
		Spec: &nexusv1.EndpointSpec{
			Name:        c.Name,
			Description: desc,
			TargetSpec: &nexusv1.EndpointTargetSpec{
				Variant: &nexusv1.EndpointTargetSpec_WorkerTargetSpec{
					WorkerTargetSpec: &nexusv1.WorkerTargetSpec{
						NamespaceId: c.TargetNamespace,
						TaskQueue:   c.TargetTaskQueue,
					},
				},
			},
			PolicySpecs: policySpecs,
		},
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudNexusEndpointDeleteCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	endpoint, err := getNexusEndpointByName(cctx, client, c.Name)
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

	rv := endpoint.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.DeleteNexusEndpoint(cctx, &cloudservice.DeleteNexusEndpointRequest{
		EndpointId:       endpoint.Id,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleDeleteOperation(cctx, resp, err)
}

func (c *CloudNexusEndpointUpdateCommand) run(cctx *CommandContext, _ []string) error {
	// Validate flag combinations.
	if c.UnsetDescription && (c.Description != "" || c.DescriptionFile != "") {
		return errors.New("--unset-description cannot be used with --description or --description-file")
	}
	desc, err := resolveNexusDescription(c.Description, c.DescriptionFile)
	if err != nil {
		return err
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	endpoint, err := getNexusEndpointByName(cctx, client, c.Name)
	if err != nil {
		return err
	}
	newSpec := proto.Clone(endpoint.Spec).(*nexusv1.EndpointSpec)

	// Apply only explicitly provided fields.
	if c.UnsetDescription {
		newSpec.Description = nil
	} else if desc != nil {
		newSpec.Description = desc
	}
	if c.Command.Flags().Changed("target-namespace") {
		newSpec.TargetSpec.GetWorkerTargetSpec().NamespaceId = c.TargetNamespace
	}
	if c.Command.Flags().Changed("target-task-queue") {
		newSpec.TargetSpec.GetWorkerTargetSpec().TaskQueue = c.TargetTaskQueue
	}

	yes, err := cctx.GetPrompter().PromptApply(endpoint.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting update.")
	}

	rv := endpoint.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateNexusEndpoint(cctx, &cloudservice.UpdateNexusEndpointRequest{
		EndpointId:       endpoint.Id,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

// getNexusEndpointByName looks up a Nexus Endpoint by name using the list RPC with a name filter.
func getNexusEndpointByName(
	cctx *CommandContext,
	client cloudservice.CloudServiceClient,
	name string,
) (*nexusv1.Endpoint, error) {
	res, err := client.GetNexusEndpoints(cctx, &cloudservice.GetNexusEndpointsRequest{
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	if len(res.Endpoints) == 0 {
		return nil, fmt.Errorf("endpoint %q not found", name)
	}
	return res.Endpoints[0], nil
}

func (c *CloudNexusEndpointAllowedNamespaceListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	endpoint, err := getNexusEndpointByName(cctx, client, c.Name)
	if err != nil {
		return err
	}

	namespaceIDs := make([]string, 0, len(endpoint.Spec.PolicySpecs))
	for _, ps := range endpoint.Spec.PolicySpecs {
		if ns := ps.GetAllowedCloudNamespacePolicySpec(); ns != nil {
			namespaceIDs = append(namespaceIDs, ns.NamespaceId)
		}
	}

	return cctx.Printer.PrintStructured(struct {
		Namespaces []string
	}{
		Namespaces: namespaceIDs,
	}, printer.StructuredOptions{})
}

func (c *CloudNexusEndpointAllowedNamespaceAddCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	endpoint, err := getNexusEndpointByName(cctx, client, c.Name)
	if err != nil {
		return err
	}

	newSpec := proto.Clone(endpoint.Spec).(*nexusv1.EndpointSpec)

	existingNSMap := make(map[string]struct{}, len(newSpec.PolicySpecs))
	for _, ps := range newSpec.PolicySpecs {
		if ns := ps.GetAllowedCloudNamespacePolicySpec(); ns != nil {
			existingNSMap[ns.NamespaceId] = struct{}{}
		}
	}

	for _, ns := range c.Namespace {
		if _, ok := existingNSMap[ns]; !ok {
			newSpec.PolicySpecs = append(newSpec.PolicySpecs, &nexusv1.EndpointPolicySpec{
				Variant: &nexusv1.EndpointPolicySpec_AllowedCloudNamespacePolicySpec{
					AllowedCloudNamespacePolicySpec: &nexusv1.AllowedCloudNamespacePolicySpec{
						NamespaceId: ns,
					},
				},
			})
		}
	}

	if proto.Equal(endpoint.Spec, newSpec) {
		cctx.Printer.Println("No changes to apply: all specified namespaces are already allowed.")
		return nil
	}

	yes, err := cctx.GetPrompter().PromptApply(endpoint.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting add.")
	}

	rv := endpoint.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateNexusEndpoint(cctx, &cloudservice.UpdateNexusEndpointRequest{
		EndpointId:       endpoint.Id,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNexusEndpointAllowedNamespaceSetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	endpoint, err := getNexusEndpointByName(cctx, client, c.Name)
	if err != nil {
		return err
	}

	newSpec := proto.Clone(endpoint.Spec).(*nexusv1.EndpointSpec)
	newSpec.PolicySpecs = make([]*nexusv1.EndpointPolicySpec, len(c.Namespace))
	for i, ns := range c.Namespace {
		newSpec.PolicySpecs[i] = &nexusv1.EndpointPolicySpec{
			Variant: &nexusv1.EndpointPolicySpec_AllowedCloudNamespacePolicySpec{
				AllowedCloudNamespacePolicySpec: &nexusv1.AllowedCloudNamespacePolicySpec{
					NamespaceId: ns,
				},
			},
		}
	}

	yes, err := cctx.GetPrompter().PromptApply(endpoint.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting set.")
	}

	rv := endpoint.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateNexusEndpoint(cctx, &cloudservice.UpdateNexusEndpointRequest{
		EndpointId:       endpoint.Id,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNexusEndpointAllowedNamespaceRemoveCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	endpoint, err := getNexusEndpointByName(cctx, client, c.Name)
	if err != nil {
		return err
	}

	newSpec := proto.Clone(endpoint.Spec).(*nexusv1.EndpointSpec)

	toRemove := make(map[string]struct{}, len(c.Namespace))
	for _, ns := range c.Namespace {
		toRemove[ns] = struct{}{}
	}

	var updatedPolicySpecs []*nexusv1.EndpointPolicySpec
	for _, ps := range newSpec.PolicySpecs {
		ns := ps.GetAllowedCloudNamespacePolicySpec()
		_, remove := toRemove[ns.NamespaceId]
		if ns == nil || !remove {
			updatedPolicySpecs = append(updatedPolicySpecs, ps)
		}
	}
	newSpec.PolicySpecs = updatedPolicySpecs

	if proto.Equal(endpoint.Spec, newSpec) {
		cctx.Printer.Println("No changes to apply: none of the specified namespaces are currently allowed.")
		return nil
	}

	yes, err := cctx.GetPrompter().PromptApply(endpoint.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting remove.")
	}

	rv := endpoint.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateNexusEndpoint(cctx, &cloudservice.UpdateNexusEndpointRequest{
		EndpointId:       endpoint.Id,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}
