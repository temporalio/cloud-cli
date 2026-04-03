package temporalcloudcli

import (
	"errors"
	"slices"

	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudNamespaceCertFilterListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}
	res, err := client.GetNamespace(cctx, &cloudservice.GetNamespaceRequest{Namespace: c.Namespace})
	if err != nil {
		return err
	}
	filters := res.GetNamespace().GetSpec().GetMtlsAuth().GetCertificateFilters()
	if filters == nil {
		filters = []*namespacev1.CertificateFilterSpec{}
	}
	return cctx.Printer.PrintResourceList(
		struct {
			CertificateFilters []*namespacev1.CertificateFilterSpec
		}{CertificateFilters: filters},
		printer.PrintResourceOptions{},
		printer.TableOptions{},
	)
}

func (c *CloudNamespaceCertFilterCreateCommand) run(cctx *CommandContext, _ []string) error {
	filter, err := buildCertFilterFromFlags(c.CommonName, c.Organization, c.OrganizationalUnit, c.SubjectAlternativeName)
	if err != nil {
		return err
	}

	yes, err := cctx.GetPrompter().PromptYes("Create")
	if err != nil {
		return err
	}
	if !yes {
		return errors.New("Aborting create.")
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
	if newSpec.MtlsAuth == nil {
		newSpec.MtlsAuth = &namespacev1.MtlsAuthSpec{}
	}
	newSpec.MtlsAuth.CertificateFilters = append(newSpec.MtlsAuth.CertificateFilters, filter)

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

func (c *CloudNamespaceCertFilterDeleteCommand) run(cctx *CommandContext, _ []string) error {
	filter, err := buildCertFilterFromFlags(c.CommonName, c.Organization, c.OrganizationalUnit, c.SubjectAlternativeName)
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
	if newSpec.MtlsAuth == nil {
		newSpec.MtlsAuth = &namespacev1.MtlsAuthSpec{}
	}
	existingFilters := newSpec.MtlsAuth.GetCertificateFilters()
	var newFilters []*namespacev1.CertificateFilterSpec
	for _, existing := range existingFilters {
		if !slices.ContainsFunc([]*namespacev1.CertificateFilterSpec{filter}, func(toRemove *namespacev1.CertificateFilterSpec) bool {
			return proto.Equal(existing, toRemove)
		}) {
			newFilters = append(newFilters, existing)
		}
	}
	newSpec.MtlsAuth.CertificateFilters = newFilters

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

// buildCertFilterFromFlags creates a CertificateFilterSpec from command line flags.
// Returns an error if no fields are specified (at least one field is required).
func buildCertFilterFromFlags(commonName, organization, organizationalUnit, subjectAlternativeName string) (*namespacev1.CertificateFilterSpec, error) {
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
