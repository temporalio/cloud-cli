package temporalcloudcli

import (
	"errors"
	"maps"
	"slices"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

// haRegionStatus pairs a region ID with its status state for display purposes.
type haRegionStatus struct {
	Region string
	Status namespacev1.NamespaceRegionStatus_State
}

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
	return cctx.Printer.PrintStructured(struct {
		Namespace              string
		ActiveRegion           string
		ManagedFailoverEnabled bool
	}{
		Namespace:              c.Namespace,
		ActiveRegion:           ns.GetActiveRegion(),
		ManagedFailoverEnabled: !ns.GetSpec().GetHighAvailability().GetDisableManagedFailover(),
	}, printer.StructuredOptions{})
}

func (c *CloudNamespaceHaUpdateCommand) run(cctx *CommandContext, _ []string) error {
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
	newSpec.HighAvailability.DisableManagedFailover = c.DisableAutoFailover

	yes, err := cctx.GetPrompter().PromptApply(ns.Spec, newSpec, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting update.")
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
	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	ns := res.Namespace
	projectedNs := proto.Clone(ns).(*namespacev1.Namespace)
	projectedNs.ActiveRegion = c.Region

	yes, err := cctx.GetPrompter().PromptApply(ns, projectedNs, false)
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
	m := res.Namespace.GetRegionStatus()
	keys := slices.Sorted(maps.Keys(m))
	regions := make([]haRegionStatus, len(keys))
	for i, k := range keys {
		regions[i] = haRegionStatus{Region: k, Status: m[k].GetState()}
	}
	return cctx.Printer.PrintResourceList(
		struct{ Regions []haRegionStatus }{Regions: regions},
		printer.PrintResourceOptions{},
		printer.TableOptions{},
	)
}

func (c *CloudNamespaceHaRegionAddCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	nsRes, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	ns := nsRes.Namespace
	projectedNs := proto.Clone(ns).(*namespacev1.Namespace)
	if projectedNs.RegionStatus == nil {
		projectedNs.RegionStatus = make(map[string]*namespacev1.NamespaceRegionStatus)
	}
	projectedNs.RegionStatus[c.Region] = &namespacev1.NamespaceRegionStatus{
		State: namespacev1.NamespaceRegionStatus_STATE_PASSIVE,
	}

	yes, err := cctx.GetPrompter().PromptApply(ns, projectedNs, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
	}

	rv := ns.ResourceVersion
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
	nsRes, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	ns := nsRes.Namespace
	projectedNs := proto.Clone(ns).(*namespacev1.Namespace)
	delete(projectedNs.RegionStatus, c.Region)

	yes, err := cctx.GetPrompter().PromptApply(ns, projectedNs, false)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting delete.")
	}

	rv := ns.ResourceVersion
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
