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

	var endpoints []*nexusv1.Endpoint
	pageToken := ""
	for {
		res, err := client.GetNexusEndpoints(cctx, &cloudservice.GetNexusEndpointsRequest{
			PageToken: pageToken,
		})
		if err != nil {
			return err
		}
		endpoints = append(endpoints, res.Endpoints...)
		pageToken = res.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return cctx.Printer.PrintResourceList(
		struct {
			Endpoints []*nexusv1.Endpoint
		}{
			Endpoints: endpoints,
		},
		printer.PrintResourceOptions{
			Fields:     []string{"Id", "State"},
			SpecFields: []string{"Name"},
		},
		printer.TableOptions{},
	)
}

func (c *CloudNexusEndpointGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	// The singular GetNexusEndpoint RPC requires an endpoint ID, but this command takes a name.
	// Use the list RPC with a name filter instead.
	res, err := client.GetNexusEndpoints(cctx, &cloudservice.GetNexusEndpointsRequest{
		Name: c.Name,
	})
	if err != nil {
		return err
	}

	if len(res.Endpoints) == 0 {
		return fmt.Errorf("endpoint %q not found", c.Name)
	}

	return cctx.Printer.PrintResource(res.Endpoints[0], printer.PrintResourceOptions{})
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

	res, err := client.GetNexusEndpoints(cctx, &cloudservice.GetNexusEndpointsRequest{
		Name: c.Name,
	})
	if err != nil {
		return err
	}
	if len(res.Endpoints) == 0 {
		return fmt.Errorf("endpoint %q not found", c.Name)
	}
	endpoint := res.Endpoints[0]

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

	res, err := client.GetNexusEndpoints(cctx, &cloudservice.GetNexusEndpointsRequest{
		Name: c.Name,
	})
	if err != nil {
		return err
	}
	if len(res.Endpoints) == 0 {
		return fmt.Errorf("endpoint %q not found", c.Name)
	}
	endpoint := res.Endpoints[0]
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
