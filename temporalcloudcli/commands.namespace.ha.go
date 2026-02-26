package temporalcloudcli

import (
	"errors"
	"fmt"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceHaGetCommand) run(cctx *CommandContext, _ []string) error {
	nsClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	ns, err := nsClient.GetNamespace(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	result := struct {
		Namespace              string
		ActiveRegion           string
		ManagedFailoverEnabled bool
	}{
		Namespace:              c.Namespace,
		ActiveRegion:           ns.GetActiveRegion(),
		ManagedFailoverEnabled: !ns.GetSpec().GetHighAvailability().GetDisableManagedFailover(),
	}
	return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
}

func (c *CloudNamespaceHaUpdateCommand) run(cctx *CommandContext, _ []string) error {
	haClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	update := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, haClient.UpdateHA)
	return update(namespace.UpdateHAParams{
		Namespace:           c.Namespace,
		DisableAutoFailover: c.DisableAutoFailover,
		ResourceVersion:     c.ResourceVersion,
		AsyncOperationID:    c.AsyncOperationId,
	})
}

func (c *CloudNamespaceHaFailoverCommand) run(cctx *CommandContext, _ []string) error {
	haClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	yes, err := cctx.promptYes("Failover (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting failover.")
	}

	failover := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, haClient.Failover)
	return failover(namespace.FailoverParams{
		Namespace:        c.Namespace,
		Region:           c.Region,
		AsyncOperationID: c.AsyncOperationId,
	})
}

func (c *CloudNamespaceHaRegionListCommand) run(cctx *CommandContext, _ []string) error {
	haClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	regions, err := haClient.ListRegions(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	return cctx.Printer.PrintResourceList(
		struct{ Regions []namespace.RegionStatus }{Regions: regions},
		printer.PrintResourceOptions{},
		printer.TableOptions{},
	)
}

func (c *CloudNamespaceHaRegionAddCommand) run(cctx *CommandContext, _ []string) error {
	haClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	prompt := fmt.Sprintf("Add region %s (y/yes)?", c.Region)
	yes, err := cctx.promptYes(prompt, cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
	}

	addRegion := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, haClient.AddRegion)
	return addRegion(namespace.AddRegionParams{
		Namespace:        c.Namespace,
		Region:           c.Region,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	})
}

func (c *CloudNamespaceHaRegionDeleteCommand) run(cctx *CommandContext, _ []string) error {
	haClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	prompt := fmt.Sprintf("Delete region %s (y/yes)?", c.Region)
	yes, err := cctx.promptYes(prompt, cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting delete.")
	}

	removeRegion := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, haClient.RemoveRegion)
	return removeRegion(namespace.RemoveRegionParams{
		Namespace:        c.Namespace,
		Region:           c.Region,
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	})
}
