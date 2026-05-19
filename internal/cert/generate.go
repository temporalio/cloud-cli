package cert

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// AIDEV-NOTE: CA validity is clamped to [7d, 365d] — short bound prevents trivially-expired CAs;
// long bound discourages multi-year self-signed authorities that can't be rotated easily.
const (
	maxCAValidity = 365 * 24 * time.Hour
	minCAValidity = 7 * 24 * time.Hour

	pemTypeCertificate = "CERTIFICATE"
	pemTypePrivateKey  = "PRIVATE KEY"
)

// GenerateCA creates a self-signed CA certificate and a PKCS8-encoded private key,
// both PEM-encoded. The organization must be non-empty and the validity period must
// fall in [7d, 365d]. The CA uses an ECDSA P-384 key.
func GenerateCA(organization string, validityPeriod time.Duration) (certPEM, keyPEM []byte, err error) {
	if validityPeriod > maxCAValidity || validityPeriod < minCAValidity {
		return nil, nil, fmt.Errorf("validity period must be between %s and %s", minCAValidity, maxCAValidity)
	}
	if len(organization) == 0 {
		return nil, nil, fmt.Errorf("organization must be a non-empty string")
	}

	orgLabel := sanitizeDNSLabel(organization)
	if orgLabel == "" {
		return nil, nil, fmt.Errorf("organization %q contains no valid DNS label characters", organization)
	}

	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate a random serial number: %w", err)
	}

	randomLetters, err := generateRandomString(4)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate random string for dns name")
	}
	dnsRoot := fmt.Sprintf("client.root.%s.%s", orgLabel, randomLetters)

	keyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageCRLSign | x509.KeyUsageCertSign
	now := time.Now().UTC()
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{organization},
		},
		NotBefore:             now.Add(-time.Minute), // grace of 1 min
		NotAfter:              now.Add(validityPeriod),
		IsCA:                  true,
		KeyUsage:              keyUsage,
		BasicConstraintsValid: true,
		DNSNames:              []string{dnsRoot},
		MaxPathLen:            0,
	}

	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate key: %w", err)
	}

	return encodeCertAndKey(template, template, &key.PublicKey, key, key)
}

// GenerateEndEntity creates an end-entity certificate signed by the given CA. The CA cert
// and key must be PEM-encoded; the CA key must be PKCS8-encoded. The end-entity key type
// mirrors the CA key type (RSA-4096 if the CA is RSA, otherwise ECDSA P-384).
// validityPeriod=0 falls back to "1 day before CA expiry".
func GenerateEndEntity(
	organization, organizationalUnit, commonName string,
	validityPeriod time.Duration,
	caCertPEM, caKeyPEM []byte,
) (certPEM, keyPEM []byte, err error) {
	if validityPeriod < 0 {
		return nil, nil, fmt.Errorf("validity period must be non-negative")
	}

	caCert, caKey, isRSA, err := parseCAForSigning(caCertPEM, caKeyPEM)
	if err != nil {
		return nil, nil, err
	}

	var orgLabel string
	if organization != "" {
		orgLabel = sanitizeDNSLabel(organization)
		if orgLabel == "" {
			return nil, nil, fmt.Errorf("organization %q contains no valid DNS label characters", organization)
		}
	}

	randomLetters, err := generateRandomString(4)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate random string for dns name")
	}
	// AIDEV-NOTE: dnsRoot embeds organization (when set) plus a random suffix so each
	// end-entity gets a distinct SAN even when the rest of the subject matches.
	dnsRoot := "client.endentity." + randomLetters
	if orgLabel != "" {
		dnsRoot = fmt.Sprintf("client.endentity.%s.%s", orgLabel, randomLetters)
	}

	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate a random serial number: %w", err)
	}

	subject := pkix.Name{CommonName: commonName}
	if organization != "" {
		subject.Organization = []string{organization}
	}
	if organizationalUnit != "" {
		subject.OrganizationalUnit = []string{organizationalUnit}
	}

	// AIDEV-NOTE: when validityPeriod is 0 the caller asked for the default —
	// fall back to "1 day before CA expiry" — a safe default that won't outlive the CA.
	now := time.Now().UTC()
	var notAfter time.Time
	if validityPeriod != 0 {
		notAfter = now.Add(validityPeriod).UTC()
		if notAfter.After(caCert.NotAfter.UTC()) {
			return nil, nil, fmt.Errorf("validity period of %s puts certificate's expiry after certificate authority's expiry %s by %s",
				validityPeriod, caCert.NotAfter.UTC().String(), notAfter.Sub(caCert.NotAfter.UTC()))
		}
	} else {
		notAfter = caCert.NotAfter.UTC().Add(-24 * time.Hour)
	}

	// AIDEV-NOTE: KeyEncipherment is only meaningful for RSA (used during the TLS_RSA
	// key exchange). For ECDSA we leave it off — ECDHE key exchange uses
	// DigitalSignature instead.
	keyUsage := x509.KeyUsageDigitalSignature
	if isRSA {
		keyUsage |= x509.KeyUsageKeyEncipherment
	}
	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subject,
		NotBefore:             now.Add(-time.Minute), // grace of 1 min
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		DNSNames:              []string{dnsRoot},
	}

	var key any
	if isRSA {
		key, err = rsa.GenerateKey(rand.Reader, 4096)
	} else {
		key, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate key: %w", err)
	}

	var publicKey any
	switch k := key.(type) {
	case *rsa.PrivateKey:
		publicKey = &k.PublicKey
	case *ecdsa.PrivateKey:
		publicKey = &k.PublicKey
	}

	return encodeCertAndKey(template, caCert, publicKey, caKey, key)
}

// encodeCertAndKey signs template with parent/signer, then PEM-encodes both the resulting
// certificate (with the given public key embedded) and priv as a PKCS8 private key.
func encodeCertAndKey(template, parent *x509.Certificate, pub any, signer any, priv any) (certPEM, keyPEM []byte, err error) {
	der, err := x509.CreateCertificate(rand.Reader, template, parent, pub, signer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate certificate: %w", err)
	}
	certBuf := new(bytes.Buffer)
	if err := pem.Encode(certBuf, &pem.Block{Type: pemTypeCertificate, Bytes: der}); err != nil {
		return nil, nil, err
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to marshal key: %w", err)
	}
	keyBuf := new(bytes.Buffer)
	if err := pem.Encode(keyBuf, &pem.Block{Type: pemTypePrivateKey, Bytes: privBytes}); err != nil {
		return nil, nil, err
	}
	return certBuf.Bytes(), keyBuf.Bytes(), nil
}

func generateRandomString(n int) (string, error) {
	// AIDEV-NOTE: used as a trailing DNS label, so the alphabet stays alphanumeric —
	// including '-' here would risk producing a label that starts or ends with '-',
	// which is forbidden by RFC 1035.
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, n)
	for i := range n {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret), nil
}

// sanitizeDNSLabel returns a DNS-label-safe rendering of s: chars outside
// [A-Za-z0-9] become '-', runs of '-' collapse, edge '-' is trimmed, and the
// result is truncated to 63 bytes. Returns "" if nothing valid remains.
func sanitizeDNSLabel(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevHyphen := true // suppress leading '-' the same way we suppress runs
	for _, r := range s {
		alnum := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
		if alnum {
			b.WriteRune(r)
			prevHyphen = false
			continue
		}
		if !prevHyphen {
			b.WriteByte('-')
			prevHyphen = true
		}
	}
	out := strings.TrimRight(b.String(), "-")
	if len(out) > 63 {
		out = strings.TrimRight(out[:63], "-")
	}
	return out
}

func generateSerialNumber() (*big.Int, error) {
	// 2^130 - 1; cryptographically strong pseudo-random in [0, limit)
	limit := new(big.Int)
	limit.Exp(big.NewInt(2), big.NewInt(130), nil).Sub(limit, big.NewInt(1))
	return rand.Int(rand.Reader, limit)
}

// AIDEV-NOTE: parseCAForSigning only handles PKCS8-encoded private keys (the format
// produced by GenerateCA). PKCS1 ("RSA PRIVATE KEY") and SEC1 ("EC PRIVATE KEY") PEM
// blocks will fail to parse here. Extend with x509.ParsePKCS1PrivateKey /
// x509.ParseECPrivateKey fallbacks if we need to consume externally-generated CA keys.
// Distinct from ParseCACerts (which extracts metadata from a bundle of public certs);
// this one loads a single CA cert + its signing key for issuing new certificates.
func parseCAForSigning(caCertPEM, caKeyPEM []byte) (*x509.Certificate, any, bool, error) {
	block, _ := pem.Decode(caCertPEM)
	if block == nil {
		return nil, nil, false, fmt.Errorf("decoding ca cert failed")
	}
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, false, fmt.Errorf("decoding ca cert failed: %w", err)
	}
	block, _ = pem.Decode(caKeyPEM)
	if block == nil {
		return nil, nil, false, fmt.Errorf("decoding ca key failed")
	}
	caKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, false, fmt.Errorf("parsing ca key failed: %w", err)
	}
	_, isRSA := caKey.(*rsa.PrivateKey)
	return caCert, caKey, isRSA, nil
}
