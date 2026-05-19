package temporalcloudcli

import (
	"fmt"
	"os"
	"time"

	"github.com/temporalio/cloud-cli/internal/cert"
)

func (c *CloudCertificatesGenerateClientCommand) run(cctx *CommandContext, _ []string) error {
	caCertPEM, err := os.ReadFile(c.CaCertPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", c.CaCertPath, err)
	}
	caKeyPEM, err := os.ReadFile(c.CaKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", c.CaKeyPath, err)
	}
	certPEM, keyPEM, err := cert.GenerateEndEntity(
		c.Organization,
		c.OrganizationalUnit,
		c.CommonName,
		time.Duration(c.ValidityPeriod),
		caCertPEM,
		caKeyPEM,
	)
	if err != nil {
		return fmt.Errorf("unable to generate client certificate: %w", err)
	}
	return writeCertificates(cctx, "client", certPEM, keyPEM, c.ClientCertPath, c.ClientKeyPath)
}
