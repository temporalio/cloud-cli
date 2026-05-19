package temporalcloudcli_test

import (
	"crypto/x509"
	"encoding/pem"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/temporalio/cloud-cli/internal/cert"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

// writeCA generates a fresh CA in a temp dir and returns the paths. validityPeriod=0
// means use a sensible 30d default.
func writeCA(t *testing.T, validityPeriod time.Duration) (caCertPath, caKeyPath string) {
	t.Helper()
	if validityPeriod == 0 {
		validityPeriod = 30 * 24 * time.Hour
	}
	caCertPEM, caKeyPEM, err := cert.GenerateCA("Acme", validityPeriod)
	require.NoError(t, err)
	dir := t.TempDir()
	caCertPath = filepath.Join(dir, "ca.pem")
	caKeyPath = filepath.Join(dir, "ca.key")
	require.NoError(t, os.WriteFile(caCertPath, caCertPEM, 0600))
	require.NoError(t, os.WriteFile(caKeyPath, caKeyPEM, 0600))
	return caCertPath, caKeyPath
}

func clientPathsInDir(t *testing.T) (certPath, keyPath string) {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "client.pem"), filepath.Join(dir, "client.key")
}

func TestCloudCertificatesGenerateClient_Success(t *testing.T) {
	caCertPath, caKeyPath := writeCA(t, 0)
	clientCertPath, clientKeyPath := clientPathsInDir(t)

	cmd := temporalcloudcli.CloudCertificatesGenerateClientCommand{
		CaCertPath:     caCertPath,
		CaKeyPath:      caKeyPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
		Organization:   "Acme",
		CommonName:     "client-1",
		ValidityPeriod: validity(7 * 24 * time.Hour),
	}

	expectedPrompt := "the client (private) key will be stored at " + clientKeyPath +
		" - do not share this key with anyone. confirm:"

	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes:        true,
			ExpectPromptYesMessage: expectedPrompt,
			PromptResult:           true,
		},
		ExpectedOutput: "client generated at: " + clientCertPath + "\nclient key generated at: " + clientKeyPath + "\n",
	})

	// Verify the client cert chains to the CA.
	caPEM, err := os.ReadFile(caCertPath)
	require.NoError(t, err)
	caBlock, _ := pem.Decode(caPEM)
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	require.NoError(t, err)

	clientPEM, err := os.ReadFile(clientCertPath)
	require.NoError(t, err)
	clientBlock, _ := pem.Decode(clientPEM)
	clientCert, err := x509.ParseCertificate(clientBlock.Bytes)
	require.NoError(t, err)
	assert.Equal(t, "client-1", clientCert.Subject.CommonName)

	roots := x509.NewCertPool()
	roots.AddCert(caCert)
	_, err = clientCert.Verify(x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})
	assert.NoError(t, err, "client cert should chain to the CA")

	// Key file must have 0600 mode.
	info, err := os.Stat(clientKeyPath)
	require.NoError(t, err)
	assert.Equal(t, fs.FileMode(0600), info.Mode().Perm())
}

func TestCloudCertificatesGenerateClient_CACertMissing(t *testing.T) {
	dir := t.TempDir()
	clientCertPath, clientKeyPath := clientPathsInDir(t)

	cmd := temporalcloudcli.CloudCertificatesGenerateClientCommand{
		CaCertPath:     filepath.Join(dir, "missing-ca.pem"),
		CaKeyPath:      filepath.Join(dir, "missing-ca.key"),
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
		CommonName:     "client-1",
		ValidityPeriod: validity(7 * 24 * time.Hour),
	}

	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		ExpectedError: "failed to read",
	})
}

func TestCloudCertificatesGenerateClient_CAKeyMissing(t *testing.T) {
	caCertPath, _ := writeCA(t, 0)
	clientCertPath, clientKeyPath := clientPathsInDir(t)

	cmd := temporalcloudcli.CloudCertificatesGenerateClientCommand{
		CaCertPath:     caCertPath,
		CaKeyPath:      filepath.Join(filepath.Dir(caCertPath), "missing-ca.key"),
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
		CommonName:     "client-1",
		ValidityPeriod: validity(7 * 24 * time.Hour),
	}

	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		ExpectedError: "missing-ca.key",
	})
}

func TestCloudCertificatesGenerateClient_ValidityBeyondCA(t *testing.T) {
	caCertPath, caKeyPath := writeCA(t, 7*24*time.Hour)
	clientCertPath, clientKeyPath := clientPathsInDir(t)

	cmd := temporalcloudcli.CloudCertificatesGenerateClientCommand{
		CaCertPath:     caCertPath,
		CaKeyPath:      caKeyPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
		CommonName:     "client-1",
		ValidityPeriod: validity(30 * 24 * time.Hour),
	}

	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		ExpectedError: "puts certificate's expiry after certificate authority's expiry",
	})
	assertFilesMissing(t, clientCertPath, clientKeyPath)
}

func TestCloudCertificatesGenerateClient_DefaultValidity(t *testing.T) {
	caCertPath, caKeyPath := writeCA(t, 0)
	clientCertPath, clientKeyPath := clientPathsInDir(t)

	cmd := temporalcloudcli.CloudCertificatesGenerateClientCommand{
		CaCertPath:     caCertPath,
		CaKeyPath:      caKeyPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
		CommonName:     "client-1",
		// ValidityPeriod intentionally zero — falls back to CA.NotAfter - 24h.
	}

	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    true,
		},
	})

	clientPEM, err := os.ReadFile(clientCertPath)
	require.NoError(t, err)
	block, _ := pem.Decode(clientPEM)
	require.NotNil(t, block)
	_, err = x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)
}

func TestCloudCertificatesGenerateClient_DeclinePrompt(t *testing.T) {
	caCertPath, caKeyPath := writeCA(t, 0)
	clientCertPath, clientKeyPath := clientPathsInDir(t)

	cmd := temporalcloudcli.CloudCertificatesGenerateClientCommand{
		CaCertPath:     caCertPath,
		CaKeyPath:      caKeyPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
		CommonName:     "client-1",
		ValidityPeriod: validity(7 * 24 * time.Hour),
	}

	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    false,
		},
	})
	assertFilesMissing(t, clientCertPath, clientKeyPath)
}
