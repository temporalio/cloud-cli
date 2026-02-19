package cert_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cloud-cli/internal/cert"
)

func TestParseCACerts_SingleCertificate(t *testing.T) {
	notBefore := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	notAfter := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	certData, certDER := generateTestCertificate(t, "test.temporal.io", notBefore, notAfter, "Engineering")

	// Compute expected fingerprint
	x509Cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)
	sum := sha1.Sum(x509Cert.Raw)
	expectedFingerprint := strings.ToLower(hex.EncodeToString(sum[:]))

	// Build expected CACert - Base64EncodedData should match what ParseCACerts stores (trimmed PEM)
	expected := cert.CACert{
		Fingerprint:       expectedFingerprint,
		Issuer:            "CN=test.temporal.io,OU=Engineering,O=Temporal,L=Bellevue,ST=Washington,C=US",
		Subject:           "CN=test.temporal.io,OU=Engineering,O=Temporal,L=Bellevue,ST=Washington,C=US",
		NotBefore:         notBefore,
		NotAfter:          notAfter,
		Base64EncodedData: base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(string(certData)))),
	}

	certs, err := cert.ParseCACerts(certData)
	require.NoError(t, err)
	require.Len(t, certs, 1)

	assert.Equal(t, expected, certs[0])
}

func TestParseCACerts_MultipleCertificates(t *testing.T) {
	cert1, _ := generateTestCertificate(t, "test1.temporal.io", time.Time{}, time.Time{}, "")
	cert2, _ := generateTestCertificate(t, "test2.temporal.io", time.Time{}, time.Time{}, "")
	cert3, _ := generateTestCertificate(t, "test3.temporal.io", time.Time{}, time.Time{}, "")

	// Combine certificates with newlines
	certData := bytes.Join([][]byte{cert1, cert2, cert3}, []byte("\n"))

	certs, err := cert.ParseCACerts(certData)
	require.NoError(t, err)
	require.Len(t, certs, 3)

	// Verify each certificate has unique fingerprints and expected subjects
	fingerprints := make(map[string]bool)
	commonNames := []string{"test1.temporal.io", "test2.temporal.io", "test3.temporal.io"}

	for i, parsed := range certs {
		assert.NotEmpty(t, parsed.Fingerprint)
		assert.Contains(t, parsed.Subject, "CN="+commonNames[i])
		assert.Contains(t, parsed.Subject, "O=Temporal")
		assert.NotEmpty(t, parsed.Base64EncodedData)

		// Ensure fingerprints are unique
		assert.False(t, fingerprints[parsed.Fingerprint], "Duplicate fingerprint found")
		fingerprints[parsed.Fingerprint] = true
	}
}

func TestParseCACerts_WithWhitespace(t *testing.T) {
	cert1, _ := generateTestCertificate(t, "test1.temporal.io", time.Time{}, time.Time{}, "")
	cert2, _ := generateTestCertificate(t, "test2.temporal.io", time.Time{}, time.Time{}, "")

	// Add various whitespace between certificates
	certData := bytes.Join([][]byte{
		cert1,
		[]byte("\n\n\n"),
		cert2,
		[]byte("\n"),
	}, nil)

	certs, err := cert.ParseCACerts(certData)
	require.NoError(t, err)
	require.Len(t, certs, 2)

	assert.Contains(t, certs[0].Subject, "CN=test1.temporal.io")
	assert.Contains(t, certs[1].Subject, "CN=test2.temporal.io")
}

func TestParseCACerts_EmptyInput(t *testing.T) {
	certs, err := cert.ParseCACerts([]byte{})
	require.NoError(t, err)
	assert.Empty(t, certs)
}

func TestParseCACerts_InvalidPEM(t *testing.T) {
	// Generate valid cert first
	validCert, _ := generateTestCertificate(t, "valid.temporal.io", time.Time{}, time.Time{}, "")

	tests := []struct {
		name    string
		data    []byte
		wantErr string
	}{
		{
			name:    "garbage data",
			data:    []byte("this is not a valid PEM certificate"),
			wantErr: "", // No error, just returns empty list
		},
		{
			name: "invalid certificate data",
			data: []byte(`-----BEGIN CERTIFICATE-----
SGVsbG8gV29ybGQ=
-----END CERTIFICATE-----`),
			wantErr: "x509: malformed certificate",
		},
		{
			name: "mixed valid and invalid",
			data: append(validCert, []byte(`-----BEGIN CERTIFICATE-----
SGVsbG8gV29ybGQ=
-----END CERTIFICATE-----`)...),
			wantErr: "x509: malformed certificate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			certs, err := cert.ParseCACerts(tt.data)

			if tt.wantErr == "" {
				require.NoError(t, err)
				assert.Empty(t, certs)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

// generateTestCertificate creates a self-signed certificate for testing and returns the PEM-encoded bytes and DER bytes.
// If notBefore/notAfter are zero values, defaults to time.Now() and time.Now()+1hour.
// If ou is empty, no OrganizationalUnit is added.
func generateTestCertificate(t *testing.T, cn string, notBefore, notAfter time.Time, ou string) ([]byte, []byte) {
	t.Helper()

	// Use default times if not specified
	if notBefore.IsZero() {
		notBefore = time.Now()
	}
	if notAfter.IsZero() {
		notAfter = time.Now().Add(time.Hour)
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Create certificate template
	subject := pkix.Name{
		Country:      []string{"US"},
		Province:     []string{"Washington"},
		Locality:     []string{"Bellevue"},
		Organization: []string{"Temporal"},
		CommonName:   cn,
	}
	if ou != "" {
		subject.OrganizationalUnit = []string{ou}
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               subject,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	// Encode to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	return certPEM, certDER
}
