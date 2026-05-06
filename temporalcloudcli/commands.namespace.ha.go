package temporalcloudcli

import (
	"errors"
	"fmt"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceHaGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	ns := res.Namespace

	result := struct {
		Namespace                      string
		ActiveRegion                   string
		ManagedFailoverEnabled         bool
		PassivePollerForwardingEnabled bool
	}{
		Namespace:                      c.Namespace,
		ActiveRegion:                   ns.GetActiveRegion(),
		ManagedFailoverEnabled:         !ns.GetSpec().GetHighAvailability().GetDisableManagedFailover(),
		PassivePollerForwardingEnabled: !ns.GetSpec().GetHighAvailability().GetDisablePassivePollerForwarding(),
	}
	return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
}

func (c *CloudNamespaceHaUpdateCommand) run(cctx *CommandContext, _ []string) error {
	disableAutoFailoverChanged := c.Command.Flags().Changed("disable-auto-failover")
	disablePassivePollerForwardingChanged := c.Command.Flags().Changed("disable-passive-poller-forwarding")
	if !disableAutoFailoverChanged && !disablePassivePollerForwardingChanged {
		return errors.New("no changes specified")
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	ns := res.Namespace
	newSpec := proto.Clone(ns.Spec).(*namespacev1.NamespaceSpec)
	if newSpec.HighAvailability == nil {
		newSpec.HighAvailability = &namespacev1.HighAvailabilitySpec{}
	}
	if disableAutoFailoverChanged {
		newSpec.HighAvailability.DisableManagedFailover = c.DisableAutoFailover
	}
	if disablePassivePollerForwardingChanged {
		newSpec.HighAvailability.DisablePassivePollerForwarding = c.DisablePassivePollerForwarding
	}

	rv := ns.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.UpdateNamespace(cctx, &cloudservice.UpdateNamespaceRequest{
		Namespace:        c.Namespace,
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNamespaceHaFailoverCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	yes, err := cctx.GetPrompter().PromptYes("Failover")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting failover.")
	}

	resp, err := client.FailoverNamespaceRegion(cctx, &cloudservice.FailoverNamespaceRegionRequest{
		Namespace:        c.Namespace,
		Region:           c.Region,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudNamespaceHaRegionListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}

	type regionStatus struct {
		Region string
		Status namespacev1.NamespaceRegionStatus_State
	}
	replicas := res.Namespace.GetReplicas()

	return cctx.Printer.PrintResourceList(
		struct{ Regions []*namespacev1.Replica }{Regions: replicas},
		printer.PrintResourceOptions{},
		printer.TableOptions{},
	)
}

func (c *CloudNamespaceHaRegionAddCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	yes, err := cctx.GetPrompter().PromptYes(fmt.Sprintf("Add region %s", c.Region))
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
	}

	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}

	rv := res.Namespace.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.AddNamespaceRegion(cctx, &cloudservice.AddNamespaceRegionRequest{
		Namespace:        c.Namespace,
		Region:           c.Region,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleCreateAsyncOperationResponse(cctx, resp, err)
}

func (c *CloudNamespaceHaRegionDeleteCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	yes, err := cctx.GetPrompter().PromptYes(fmt.Sprintf("Delete region %s", c.Region))
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting delete.")
	}

	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}

	rv := res.Namespace.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}
	resp, err := client.DeleteNamespaceRegion(cctx, &cloudservice.DeleteNamespaceRegionRequest{
		Namespace:        c.Namespace,
		Region:           c.Region,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleDeleteOperation(cctx, resp, err)
}
