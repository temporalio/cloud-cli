package temporalcloudcli_test

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cliext "github.com/temporalio/cli/cliext"

	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

func pathsInDir(t *testing.T) (certPath, keyPath string) {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "ca.pem"), filepath.Join(dir, "ca.key")
}

func validity(d time.Duration) cliext.FlagDuration {
	return cliext.FlagDuration(d)
}

// cleanCaPromptMsg is the exact prompt the command shows when neither output path
// already exists (i.e. no overwrite warning, only the key-handling warning).
func cleanCaPromptMsg(keyPath string) string {
	return "the ca (private) key will be stored at " + keyPath +
		" - do not share this key with anyone. confirm:"
}

func TestCloudCertificatesGenerateCa_Success(t *testing.T) {
	certPath, keyPath := pathsInDir(t)
	cmd := temporalcloudcli.CloudCertificatesGenerateCaCommand{
		CaCertPath:     certPath,
		CaKeyPath:      keyPath,
		Organization:   "Acme",
		ValidityPeriod: validity(30 * 24 * time.Hour),
	}

	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes:        true,
			ExpectPromptYesMessage: cleanCaPromptMsg(keyPath),
			PromptResult:           true,
		},
		ExpectedOutput: "ca generated at: " + certPath + "\nca key generated at: " + keyPath + "\n",
	})

	// Verify the cert on disk decodes as a self-signed CA.
	certPEM, err := os.ReadFile(certPath)
	require.NoError(t, err)
	block, _ := pem.Decode(certPEM)
	require.NotNil(t, block)
	parsed, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)
	assert.True(t, parsed.IsCA)
	assert.Equal(t, []string{"Acme"}, parsed.Subject.Organization)

	// Key file must have 0600 mode.
	info, err := os.Stat(keyPath)
	require.NoError(t, err)
	assert.Equal(t, fs.FileMode(0600), info.Mode().Perm())
}

func TestCloudCertificatesGenerateCa_InvalidValidity(t *testing.T) {
	certPath, keyPath := pathsInDir(t)
	cmd := temporalcloudcli.CloudCertificatesGenerateCaCommand{
		CaCertPath:     certPath,
		CaKeyPath:      keyPath,
		Organization:   "Acme",
		ValidityPeriod: validity(1 * time.Hour),
	}

	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		ExpectedError: "validity period must be between",
	})
	assertFilesMissing(t, certPath, keyPath)
}

func TestCloudCertificatesGenerateCa_EmptyOrganization(t *testing.T) {
	certPath, keyPath := pathsInDir(t)
	cmd := temporalcloudcli.CloudCertificatesGenerateCaCommand{
		CaCertPath:     certPath,
		CaKeyPath:      keyPath,
		ValidityPeriod: validity(30 * 24 * time.Hour),
	}

	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		ExpectedError: "organization must be a non-empty string",
	})
	assertFilesMissing(t, certPath, keyPath)
}

// TestCloudCertificatesGenerateCa_OverwriteExisting asserts the unified prompt names
// both pre-existing files in its overwrite warning and that the files are replaced when
// the user confirms.
func TestCloudCertificatesGenerateCa_OverwriteExisting(t *testing.T) {
	certPath, keyPath := pathsInDir(t)
	require.NoError(t, os.WriteFile(certPath, []byte("old cert"), 0644))
	require.NoError(t, os.WriteFile(keyPath, []byte("old key"), 0600))

	cmd := temporalcloudcli.CloudCertificatesGenerateCaCommand{
		CaCertPath:     certPath,
		CaKeyPath:      keyPath,
		Organization:   "Acme",
		ValidityPeriod: validity(30 * 24 * time.Hour),
	}

	expectedPrompt := "the following files already exist and will be overwritten:\n" +
		"  " + certPath + "\n" +
		"  " + keyPath + "\n" +
		"the ca (private) key will be stored at " + keyPath +
		" - do not share this key with anyone. confirm:"

	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes:        true,
			ExpectPromptYesMessage: expectedPrompt,
			PromptResult:           true,
		},
	})

	// Verify the cert was actually overwritten with valid PEM (not the "old cert" stub).
	got, err := os.ReadFile(certPath)
	require.NoError(t, err)
	assert.NotEqual(t, []byte("old cert"), got)
	block, _ := pem.Decode(got)
	require.NotNil(t, block, "overwritten cert must be valid PEM")
}

// TestCloudCertificatesGenerateCa_OverwriteOnlyKey covers the case where one output
// path exists but not the other — the overwrite list should mention only the existing
// one.
func TestCloudCertificatesGenerateCa_OverwriteOnlyKey(t *testing.T) {
	certPath, keyPath := pathsInDir(t)
	require.NoError(t, os.WriteFile(keyPath, []byte("old key"), 0600))

	cmd := temporalcloudcli.CloudCertificatesGenerateCaCommand{
		CaCertPath:     certPath,
		CaKeyPath:      keyPath,
		Organization:   "Acme",
		ValidityPeriod: validity(30 * 24 * time.Hour),
	}

	expectedPrompt := "the following files already exist and will be overwritten:\n" +
		"  " + keyPath + "\n" +
		"the ca (private) key will be stored at " + keyPath +
		" - do not share this key with anyone. confirm:"

	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes:        true,
			ExpectPromptYesMessage: expectedPrompt,
			PromptResult:           true,
		},
	})
}

func TestCloudCertificatesGenerateCa_DeclinePrompt(t *testing.T) {
	certPath, keyPath := pathsInDir(t)
	cmd := temporalcloudcli.CloudCertificatesGenerateCaCommand{
		CaCertPath:     certPath,
		CaKeyPath:      keyPath,
		Organization:   "Acme",
		ValidityPeriod: validity(30 * 24 * time.Hour),
	}

	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptResult:    false,
		},
	})
	assertFilesMissing(t, certPath, keyPath)
}

func TestCloudCertificatesGenerateCa_PathIsDirectory(t *testing.T) {
	dir := t.TempDir()
	cmd := temporalcloudcli.CloudCertificatesGenerateCaCommand{
		CaCertPath:     dir,
		CaKeyPath:      filepath.Join(dir, "ca.key"),
		Organization:   "Acme",
		ValidityPeriod: validity(30 * 24 * time.Hour),
	}

	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		ExpectedError: "path cannot be a directory",
	})
}

func TestCloudCertificatesGenerateCa_PromptError(t *testing.T) {
	certPath, keyPath := pathsInDir(t)
	cmd := temporalcloudcli.CloudCertificatesGenerateCaCommand{
		CaCertPath:     certPath,
		CaKeyPath:      keyPath,
		Organization:   "Acme",
		ValidityPeriod: validity(30 * 24 * time.Hour),
	}

	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes: true,
			PromptError:     errors.New("stdin closed"),
		},
		ExpectedError: "failed to confirm",
	})
}

func assertFilesMissing(t *testing.T, paths ...string) {
	t.Helper()
	for _, p := range paths {
		_, err := os.Stat(p)
		assert.Truef(t, errors.Is(err, fs.ErrNotExist), "file %s should not exist; stat err=%v", p, err)
	}
}
