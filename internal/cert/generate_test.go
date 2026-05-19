package cert_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/cloud-cli/internal/cert"
)

func TestGenerateCA_Success(t *testing.T) {
	certPEM, keyPEM, err := cert.GenerateCA("Acme", 30*24*time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, certPEM)
	require.NotEmpty(t, keyPEM)

	certBlock, _ := pem.Decode(certPEM)
	require.NotNil(t, certBlock)
	assert.Equal(t, "CERTIFICATE", certBlock.Type)
	parsedCert, err := x509.ParseCertificate(certBlock.Bytes)
	require.NoError(t, err)
	assert.True(t, parsedCert.IsCA)
	assert.True(t, parsedCert.BasicConstraintsValid)
	assert.Equal(t, []string{"Acme"}, parsedCert.Subject.Organization)
	assert.Equal(t, parsedCert.Subject.String(), parsedCert.Issuer.String(),
		"self-signed: issuer must equal subject")

	// Validity window ~30d (plus the 1-minute back-dating grace).
	delta := parsedCert.NotAfter.Sub(parsedCert.NotBefore)
	assert.InDelta(t, (30*24*time.Hour + time.Minute).Seconds(), delta.Seconds(), 5)

	// SAN includes a DNS name embedding the org.
	require.Len(t, parsedCert.DNSNames, 1)
	assert.True(t, strings.HasPrefix(parsedCert.DNSNames[0], "client.root.Acme."),
		"DNS SAN %q should start with client.root.Acme.", parsedCert.DNSNames[0])

	// Key must be PKCS8-encoded ECDSA P-384.
	keyBlock, _ := pem.Decode(keyPEM)
	require.NotNil(t, keyBlock)
	assert.Equal(t, "PRIVATE KEY", keyBlock.Type)
	parsedKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	require.NoError(t, err)
	ecKey, isEC := parsedKey.(*ecdsa.PrivateKey)
	require.True(t, isEC, "default CA key should be ECDSA")
	assert.Equal(t, elliptic.P384(), ecKey.Curve)
}

func TestGenerateCA_ValidityErrors(t *testing.T) {
	tests := []struct {
		name     string
		validity time.Duration
		wantErr  string
	}{
		{"below minimum", 1 * time.Hour, "validity period must be between"},
		{"below minimum (just under 7d)", 7*24*time.Hour - time.Second, "validity period must be between"},
		{"above maximum (just over 365d)", 365*24*time.Hour + time.Second, "validity period must be between"},
		{"above maximum", 400 * 24 * time.Hour, "validity period must be between"},
		{"negative", -1 * time.Hour, "validity period must be between"},
		{"zero", 0, "validity period must be between"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := cert.GenerateCA("Acme", tt.validity)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestGenerateCA_BoundaryValidity(t *testing.T) {
	// Min and max boundaries are inclusive — they must succeed.
	_, _, err := cert.GenerateCA("Acme", 7*24*time.Hour)
	require.NoError(t, err, "min validity (7d) must be accepted")

	_, _, err = cert.GenerateCA("Acme", 365*24*time.Hour)
	require.NoError(t, err, "max validity (365d) must be accepted")
}

func TestGenerateCA_EmptyOrganization(t *testing.T) {
	_, _, err := cert.GenerateCA("", 30*24*time.Hour)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "organization must be a non-empty string")
}

func TestGenerateEndEntity_Success(t *testing.T) {
	caCertPEM, caKeyPEM, err := cert.GenerateCA("Acme", 30*24*time.Hour)
	require.NoError(t, err)

	eeCertPEM, eeKeyPEM, err := cert.GenerateEndEntity(
		"Acme", "Engineering", "client-1",
		7*24*time.Hour,
		caCertPEM, caKeyPEM,
	)
	require.NoError(t, err)

	certBlock, _ := pem.Decode(eeCertPEM)
	require.NotNil(t, certBlock)
	eeCert, err := x509.ParseCertificate(certBlock.Bytes)
	require.NoError(t, err)
	assert.False(t, eeCert.IsCA)
	assert.Equal(t, "client-1", eeCert.Subject.CommonName)
	assert.Equal(t, []string{"Acme"}, eeCert.Subject.Organization)
	assert.Equal(t, []string{"Engineering"}, eeCert.Subject.OrganizationalUnit)
	require.Len(t, eeCert.DNSNames, 1)
	assert.True(t, strings.HasPrefix(eeCert.DNSNames[0], "client.endentity.Acme."),
		"DNS SAN %q should start with client.endentity.Acme.", eeCert.DNSNames[0])

	// ECDSA EE: KeyUsage is DigitalSignature only (KeyEncipherment is RSA-specific).
	assert.Equal(t, x509.KeyUsageDigitalSignature, eeCert.KeyUsage)
	assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, eeCert.ExtKeyUsage)

	// EE must be signed by the CA.
	caBlock, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	require.NoError(t, err)
	roots := x509.NewCertPool()
	roots.AddCert(caCert)
	_, err = eeCert.Verify(x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})
	assert.NoError(t, err, "end-entity cert should chain to the CA")

	// EE key must be PKCS8-encoded ECDSA (mirrors the ECDSA CA).
	keyBlock, _ := pem.Decode(eeKeyPEM)
	require.NotNil(t, keyBlock)
	parsedKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	require.NoError(t, err)
	_, isEC := parsedKey.(*ecdsa.PrivateKey)
	assert.True(t, isEC, "EE key should be ECDSA when CA is ECDSA")
}

func TestGenerateEndEntity_OnlyCommonName(t *testing.T) {
	caCertPEM, caKeyPEM, err := cert.GenerateCA("Acme", 30*24*time.Hour)
	require.NoError(t, err)

	eeCertPEM, _, err := cert.GenerateEndEntity(
		"", "", "client-2",
		7*24*time.Hour,
		caCertPEM, caKeyPEM,
	)
	require.NoError(t, err)

	certBlock, _ := pem.Decode(eeCertPEM)
	eeCert, err := x509.ParseCertificate(certBlock.Bytes)
	require.NoError(t, err)
	assert.Equal(t, "client-2", eeCert.Subject.CommonName)
	assert.Empty(t, eeCert.Subject.Organization)
	assert.Empty(t, eeCert.Subject.OrganizationalUnit)
	require.Len(t, eeCert.DNSNames, 1)
	// Without organization, the SAN omits the org segment.
	assert.True(t, strings.HasPrefix(eeCert.DNSNames[0], "client.endentity."),
		"DNS SAN %q should start with client.endentity.", eeCert.DNSNames[0])
	assert.NotContains(t, eeCert.DNSNames[0], "client.endentity..",
		"no org should not produce an empty label")
}

func TestGenerateEndEntity_ValidityBeyondCA(t *testing.T) {
	caCertPEM, caKeyPEM, err := cert.GenerateCA("Acme", 7*24*time.Hour)
	require.NoError(t, err)

	_, _, err = cert.GenerateEndEntity(
		"Acme", "", "client-1",
		30*24*time.Hour, // greater than the CA's 7d validity
		caCertPEM, caKeyPEM,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "puts certificate's expiry after certificate authority's expiry")
}

func TestGenerateEndEntity_DefaultValidity(t *testing.T) {
	caCertPEM, caKeyPEM, err := cert.GenerateCA("Acme", 30*24*time.Hour)
	require.NoError(t, err)

	caBlock, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	require.NoError(t, err)

	eeCertPEM, _, err := cert.GenerateEndEntity(
		"Acme", "", "client-1",
		0, // default → CA.NotAfter - 24h
		caCertPEM, caKeyPEM,
	)
	require.NoError(t, err)

	eeBlock, _ := pem.Decode(eeCertPEM)
	eeCert, err := x509.ParseCertificate(eeBlock.Bytes)
	require.NoError(t, err)

	expected := caCert.NotAfter.Add(-24 * time.Hour)
	assert.WithinDuration(t, expected, eeCert.NotAfter, time.Second,
		"default EE validity should be CA.NotAfter - 24h")
}

func TestGenerateEndEntity_MalformedInput(t *testing.T) {
	caCertPEM, caKeyPEM, err := cert.GenerateCA("Acme", 30*24*time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name    string
		caCert  []byte
		caKey   []byte
		wantErr string
	}{
		{
			name:    "non-PEM cert",
			caCert:  []byte("not a pem"),
			caKey:   caKeyPEM,
			wantErr: "decoding ca cert failed",
		},
		{
			name: "malformed cert bytes inside valid PEM",
			caCert: []byte(`-----BEGIN CERTIFICATE-----
SGVsbG8gV29ybGQ=
-----END CERTIFICATE-----`),
			caKey:   caKeyPEM,
			wantErr: "decoding ca cert failed",
		},
		{
			name:    "non-PEM key",
			caCert:  caCertPEM,
			caKey:   []byte("not a pem"),
			wantErr: "decoding ca key failed",
		},
		{
			name:   "PKCS1 RSA key (unsupported encoding)",
			caCert: caCertPEM,
			caKey: func() []byte {
				k, err := rsa.GenerateKey(rand.Reader, 2048)
				require.NoError(t, err)
				return pem.EncodeToMemory(&pem.Block{
					Type:  "RSA PRIVATE KEY",
					Bytes: x509.MarshalPKCS1PrivateKey(k),
				})
			}(),
			wantErr: "parsing ca key failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := cert.GenerateEndEntity(
				"Acme", "", "client-1",
				7*24*time.Hour,
				tt.caCert, tt.caKey,
			)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestGenerateEndEntity_RSACA(t *testing.T) {
	caCertPEM, caKeyPEM := buildRSACA(t, 30*24*time.Hour)

	eeCertPEM, eeKeyPEM, err := cert.GenerateEndEntity(
		"Acme", "", "client-1",
		7*24*time.Hour,
		caCertPEM, caKeyPEM,
	)
	require.NoError(t, err)

	// EE key mirrors CA key type — should be RSA.
	keyBlock, _ := pem.Decode(eeKeyPEM)
	require.NotNil(t, keyBlock)
	parsedKey, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	require.NoError(t, err)
	_, isRSA := parsedKey.(*rsa.PrivateKey)
	assert.True(t, isRSA, "EE key should be RSA when CA is RSA")

	// And the cert still verifies against the RSA CA.
	caBlock, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	require.NoError(t, err)
	eeBlock, _ := pem.Decode(eeCertPEM)
	eeCert, err := x509.ParseCertificate(eeBlock.Bytes)
	require.NoError(t, err)
	roots := x509.NewCertPool()
	roots.AddCert(caCert)
	_, err = eeCert.Verify(x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	})
	assert.NoError(t, err)

	// RSA EE: KeyUsage adds KeyEncipherment so RSA key-exchange ciphersuites work.
	assert.Equal(t, x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment, eeCert.KeyUsage)
	assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, eeCert.ExtKeyUsage)
}

func TestGenerateEndEntity_NegativeValidity(t *testing.T) {
	caCertPEM, caKeyPEM, err := cert.GenerateCA("Acme", 30*24*time.Hour)
	require.NoError(t, err)

	_, _, err = cert.GenerateEndEntity(
		"Acme", "", "client-1",
		-1*time.Hour,
		caCertPEM, caKeyPEM,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validity period must be non-negative")
}

func TestGenerateCA_OrgDNSLabelSanitized(t *testing.T) {
	// Org has spaces and a slash — they must not leak into the DNS SAN.
	certPEM, _, err := cert.GenerateCA("Acme  Inc/Eng", 30*24*time.Hour)
	require.NoError(t, err)

	block, _ := pem.Decode(certPEM)
	parsed, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	// Subject keeps the original string; only the DNS label is sanitized.
	assert.Equal(t, []string{"Acme  Inc/Eng"}, parsed.Subject.Organization)
	require.Len(t, parsed.DNSNames, 1)
	assert.True(t, strings.HasPrefix(parsed.DNSNames[0], "client.root.Acme-Inc-Eng."),
		"DNS SAN %q should start with client.root.Acme-Inc-Eng.", parsed.DNSNames[0])
}

func TestGenerateCA_OrgNoValidDNSChars(t *testing.T) {
	_, _, err := cert.GenerateCA("!!!", 30*24*time.Hour)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "contains no valid DNS label characters")
}

func TestGenerateEndEntity_OrgDNSLabelSanitized(t *testing.T) {
	caCertPEM, caKeyPEM, err := cert.GenerateCA("Acme", 30*24*time.Hour)
	require.NoError(t, err)

	eeCertPEM, _, err := cert.GenerateEndEntity(
		"Acme  Inc/Eng", "", "client-1",
		7*24*time.Hour,
		caCertPEM, caKeyPEM,
	)
	require.NoError(t, err)

	block, _ := pem.Decode(eeCertPEM)
	eeCert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)
	assert.Equal(t, []string{"Acme  Inc/Eng"}, eeCert.Subject.Organization)
	require.Len(t, eeCert.DNSNames, 1)
	assert.True(t, strings.HasPrefix(eeCert.DNSNames[0], "client.endentity.Acme-Inc-Eng."),
		"DNS SAN %q should start with client.endentity.Acme-Inc-Eng.", eeCert.DNSNames[0])
}

func TestGenerateEndEntity_OrgNoValidDNSChars(t *testing.T) {
	caCertPEM, caKeyPEM, err := cert.GenerateCA("Acme", 30*24*time.Hour)
	require.NoError(t, err)

	_, _, err = cert.GenerateEndEntity(
		"!!!", "", "client-1",
		7*24*time.Hour,
		caCertPEM, caKeyPEM,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "contains no valid DNS label characters")
}

// buildRSACA builds a self-signed RSA-2048 CA with a PKCS8-encoded private key, matching the
// format GenerateCA would produce — except for the key algorithm. Used to exercise the
// RSA branch of GenerateEndEntity.
func buildRSACA(t *testing.T, validity time.Duration) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	now := time.Now().UTC()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"Acme"}},
		NotBefore:             now.Add(-time.Minute),
		NotAfter:              now.Add(validity),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCRLSign | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8})

	return certPEM, keyPEM
}
