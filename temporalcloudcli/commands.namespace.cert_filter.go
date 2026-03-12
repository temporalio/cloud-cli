package temporalcloudcli

import (
	"errors"
	"fmt"

	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"

	"github.com/temporalio/cloud-cli/internal/namespace"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceCertFilterCreateCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	filter, err := buildCertFilterFromFlags(c.CommonName, c.Organization, c.OrganizationalUnit, c.SubjectAlternativeName)
	if err != nil {
		return err
	}

	yes, err := cctx.promptYes("Create (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return errors.New("Aborting create.")
	}

	addCertFilters := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, namespaceClient.AddCertFilters)
	return addCertFilters(namespace.AddCertFiltersParams{
		Namespace:        c.Namespace,
		Filters:          []*namespacev1.CertificateFilterSpec{filter},
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	})
}

func (c *CloudNamespaceCertFilterListCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	filters, err := namespaceClient.ListCertFilters(cctx.Context, c.Namespace)
	if err != nil {
		return err
	}

	return cctx.Printer.PrintResourceList(
		struct {
			CertificateFilters []*namespacev1.CertificateFilterSpec
		}{
			CertificateFilters: filters,
		},
		printer.PrintResourceOptions{},
		printer.TableOptions{},
	)
}

func (c *CloudNamespaceCertFilterDeleteCommand) run(cctx *CommandContext, _ []string) error {
	namespaceClient, err := getNamespaceClient(cctx, c.ClientOptions)
	if err != nil {
		return err
	}

	filter, err := buildCertFilterFromFlags(c.CommonName, c.Organization, c.OrganizationalUnit, c.SubjectAlternativeName)
	if err != nil {
		return err
	}

	yes, err := cctx.promptYes("Delete (y/yes)?", cctx.RootCommand.AutoConfirm)
	if err != nil {
		return err
	}

	if !yes {
		return errors.New("Aborting delete.")
	}

	deleteCertFilters := wrapAsyncOperation(cctx, c.AsyncOperationOptions, c.Namespace, c.ClientOptions, namespaceClient.DeleteCertFilters)
	return deleteCertFilters(namespace.DeleteCertFiltersParams{
		Namespace:        c.Namespace,
		Filters:          []*namespacev1.CertificateFilterSpec{filter},
		ResourceVersion:  c.ResourceVersion,
		AsyncOperationID: c.AsyncOperationId,
	})
}

// buildCertFilterFromFlags creates a CertificateFilterSpec from command line flags.
// Returns an error if no fields are specified (at least one field is required).
func buildCertFilterFromFlags(commonName, organization, organizationalUnit, subjectAlternativeName string) (*namespacev1.CertificateFilterSpec, error) {
	// Validate that at least one field is provided
	if commonName == "" && organization == "" && organizationalUnit == "" && subjectAlternativeName == "" {
		return nil, errors.New("at least one certificate filter field must be specified (--common-name, --organization, --organizational-unit, or --subject-alternative-name)")
	}

	return &namespacev1.CertificateFilterSpec{
		CommonName:             commonName,
		Organization:           organization,
		OrganizationalUnit:     organizationalUnit,
		SubjectAlternativeName: subjectAlternativeName,
	}, nil
}
