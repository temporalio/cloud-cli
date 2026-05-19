package temporalcloudcli

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/temporalio/cloud-cli/internal/cert"
)

func (c *CloudCertificatesGenerateCaCommand) run(cctx *CommandContext, _ []string) error {
	certPEM, keyPEM, err := cert.GenerateCA(c.Organization, time.Duration(c.ValidityPeriod))
	if err != nil {
		return fmt.Errorf("unable to generate CA certificate: %w", err)
	}
	return writeCertificates(cctx, "ca", certPEM, keyPEM, c.CaCertPath, c.CaKeyPath)
}

// checkCertPath returns true if a regular file already exists at path. It returns an
// error if path is a directory or some other non-regular file — those are unsalvageable
// without user intervention beyond a yes/no prompt.
func checkCertPath(path string) (exists bool, err error) {
	fi, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to stat path %s: %w", path, err)
	}
	switch mode := fi.Mode(); {
	case mode.IsRegular():
		return true, nil
	case mode.IsDir():
		return false, fmt.Errorf("path cannot be a directory: %s", path)
	default:
		return false, fmt.Errorf("invalid file path: %s (file mode=%s)", path, mode.String())
	}
}

// writeCertificates asks the user a single yes/no question that bundles any overwrite
// warning together with the private-key handling warning, then writes the files.
func writeCertificates(cctx *CommandContext, typ string, cert, key []byte, certPath, keyPath string) error {
	certExists, err := checkCertPath(certPath)
	if err != nil {
		return err
	}
	keyExists, err := checkCertPath(keyPath)
	if err != nil {
		return err
	}

	var b strings.Builder
	if certExists || keyExists {
		b.WriteString("the following files already exist and will be overwritten:\n")
		if certExists {
			fmt.Fprintf(&b, "  %s\n", certPath)
		}
		if keyExists {
			fmt.Fprintf(&b, "  %s\n", keyPath)
		}
	}
	fmt.Fprintf(&b, "the %s (private) key will be stored at %s - do not share this key with anyone. confirm:", typ, keyPath)

	yes, err := cctx.GetPrompter().PromptYes(b.String())
	if err != nil {
		return fmt.Errorf("failed to confirm: %w", err)
	}
	if !yes {
		return nil
	}
	if err := os.WriteFile(certPath, cert, 0644); err != nil {
		return fmt.Errorf("failed to write %s certificate: %w", typ, err)
	}
	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return fmt.Errorf("failed to write %s key: %w", typ, err)
	}
	cctx.Printer.Printlnf("%s generated at: %s", typ, certPath)
	cctx.Printer.Printlnf("%s key generated at: %s", typ, keyPath)
	return nil
}
