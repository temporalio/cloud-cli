package temporalcloudcli

import (
	"errors"

	accountv1 "go.temporal.io/cloud-sdk/api/account/v1"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/internal/cert"
	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/printer"
)

func (c *CloudAccountMetricsCertCaListCommand) run(cctx *CommandContext, _ []string) error {
	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetAccount(cctx, &cloudservice.GetAccountRequest{})
	if err != nil {
		return err
	}

	certs, err := cert.ParseCACerts(res.Account.GetSpec().GetMetrics().GetAcceptedClientCa())
	if err != nil {
		return err
	}

	return cctx.Printer.PrintStructured(certs, printer.StructuredOptions{})
}

func (c *CloudAccountMetricsCertCaCreateCommand) run(cctx *CommandContext, _ []string) error {
	newCerts, err := readAndParseCACerts(c.CaCertificateOptions)
	if err != nil {
		return err
	}

	client, err := cctx.GetCloudClient(c.ClientOptions)
	if err != nil {
		return err
	}

	res, err := client.GetAccount(cctx, &cloudservice.GetAccountRequest{})
	if err != nil {
		return err
	}

	account := res.Account
	existingData := account.GetSpec().GetMetrics().GetAcceptedClientCa()

	var existingCerts []cert.CACert
	if len(existingData) > 0 {
		existingCerts, err = cert.ParseCACerts(existingData)
		if err != nil {
			return err
		}
	}

	bundleBytes, err := cert.EncodeCACerts(cert.Add(existingCerts, newCerts))
	if err != nil {
		return err
	}

	newSpec := proto.Clone(account.Spec).(*accountv1.AccountSpec)
	if newSpec.Metrics == nil {
		newSpec.Metrics = &accountv1.MetricsSpec{}
	}
	newSpec.Metrics.AcceptedClientCa = bundleBytes

	rv := account.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}

	resp, err := client.UpdateAccount(cctx, &cloudservice.UpdateAccountRequest{
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}

func (c *CloudAccountMetricsCertCaDeleteCommand) run(cctx *CommandContext, _ []string) error {
	certsToRemove, err := readAndParseCACerts(c.CaCertificateOptions)
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

	res, err := client.GetAccount(cctx, &cloudservice.GetAccountRequest{})
	if err != nil {
		return err
	}

	account := res.Account
	existingCerts, err := cert.ParseCACerts(account.GetSpec().GetMetrics().GetAcceptedClientCa())
	if err != nil {
		return err
	}

	bundleBytes, err := cert.EncodeCACerts(cert.Remove(existingCerts, certsToRemove))
	if err != nil {
		return err
	}

	newSpec := proto.Clone(account.Spec).(*accountv1.AccountSpec)
	if newSpec.Metrics == nil {
		newSpec.Metrics = &accountv1.MetricsSpec{}
	}
	newSpec.Metrics.AcceptedClientCa = bundleBytes

	rv := account.ResourceVersion
	if c.ResourceVersion != "" {
		rv = c.ResourceVersion
	}

	resp, err := client.UpdateAccount(cctx, &cloudservice.UpdateAccountRequest{
		Spec:             newSpec,
		ResourceVersion:  rv,
		AsyncOperationId: c.AsyncOperationId,
	})
	return cctx.GetPoller(client, c.AsyncOperationOptions).HandleUpdateOperation(cctx, resp, err)
}
