package temporalcloudcli

import (
	"errors"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

const (
	capacityModeOnDemand    = "on_demand"
	capacityModeProvisioned = "provisioned"
)

func (c *CloudNamespaceCapacityGetCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespaceCapacityInfo(cctx, &cloudservice.GetNamespaceCapacityInfoRequest{
		Namespace: c.Namespace,
	})
	if err != nil {
		return err
	}
	return cctx.Printer.PrintResource(res.CapacityInfo, printer.PrintResourceOptions{})
}

func (c *CloudNamespaceCapacityUpdateCommand) run(cctx *CommandContext, _ []string) error {
	var capacitySpec *namespacev1.CapacitySpec
	switch c.CapacityMode.Value {
	case capacityModeOnDemand:
		capacitySpec = &namespacev1.CapacitySpec{
			Spec: &namespacev1.CapacitySpec_OnDemand_{OnDemand: &namespacev1.CapacitySpec_OnDemand{}},
		}
	case capacityModeProvisioned:
		if c.CapacityValue <= 0 {
			return errors.New("--capacity-value must be greater than 0 when --capacity-mode is 'provisioned'")
		}
		capacitySpec = &namespacev1.CapacitySpec{
			Spec: &namespacev1.CapacitySpec_Provisioned_{
				Provisioned: &namespacev1.CapacitySpec_Provisioned{Value: float64(c.CapacityValue)},
			},
		}
	default:
		return errors.New("--capacity-mode must be 'on_demand' or 'provisioned'")
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
	newSpec.CapacitySpec = capacitySpec

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
