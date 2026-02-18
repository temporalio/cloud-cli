package temporalcloudcli

import (
	"errors"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceCertFilterListCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	filters, err := namespaceClient.ListCertFilters(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	return cctx.Printer.PrintStructured(filters, printer.StructuredOptions{})
}

func (c *CloudNamespaceCertFilterCreateCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	// At least one field must be specified
	if c.CommonName == "" && c.Organization == "" && c.OrganizationalUnit == "" && c.SubjectAlternativeName == "" {
		return errors.New("at least one filter field must be specified")
	}

	yes, err := cctx.promptYes("Create cert filter (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return errors.New("Aborting create.")
	}

	filter := namespace.CertFilter{
		CommonName:             c.CommonName,
		Organization:           c.Organization,
		OrganizationalUnit:     c.OrganizationalUnit,
		SubjectAlternativeName: c.SubjectAlternativeName,
	}

	params := namespace.AddCertFiltersParams{
		Namespace:        c.Namespace,
		Filters:          []namespace.CertFilter{filter},
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	}
	op, err := namespaceClient.AddCertFilters(cctx.Context, params)
	if err != nil {
		if isNothingChangedErr(c.Idempotent, err) {
			result := struct {
				Status    string
				Namespace string
			}{
				Status:    "unchanged",
				Namespace: c.Namespace,
			}
			return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
		}

		return err
	}

	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: op,
			ID:      c.Namespace,
		}, printer.StructuredOptions{})
	}

	poller, err := getPoller(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	return poller.Poll(cctx.Context, op.Id, c.Namespace)
}

func (c *CloudNamespaceCertFilterDeleteCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	// At least one field must be specified
	if c.CommonName == "" && c.Organization == "" && c.OrganizationalUnit == "" && c.SubjectAlternativeName == "" {
		return errors.New("at least one filter field must be specified")
	}

	filter := namespace.CertFilter{
		CommonName:             c.CommonName,
		Organization:           c.Organization,
		OrganizationalUnit:     c.OrganizationalUnit,
		SubjectAlternativeName: c.SubjectAlternativeName,
	}

	yes, err := cctx.promptYes("Delete (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return errors.New("Aborting delete.")
	}

	params := namespace.DeleteCertFiltersParams{
		Namespace:        c.Namespace,
		Filters:          []namespace.CertFilter{filter},
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	}
	op, err := namespaceClient.DeleteCertFilters(cctx.Context, params)
	if err != nil {
		if isNothingChangedErr(c.Idempotent, err) {
			result := struct {
				Status    string
				Namespace string
			}{
				Status:    "unchanged",
				Namespace: c.Namespace,
			}
			return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
		}

		return err
	}

	if c.Async {
		// Return immediately with the async operation
		return cctx.Printer.PrintStructured(MutationResult{
			AsyncOp: op,
			ID:      c.Namespace,
		}, printer.StructuredOptions{})
	}

	poller, err := getPoller(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	return poller.Poll(cctx.Context, op.Id, c.Namespace)
}
