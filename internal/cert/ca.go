package cert

import (
	"bytes"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"strings"
	"time"
)

type CACert struct {
	Fingerprint       string
	Issuer            string
	Subject           string
	NotBefore         time.Time
	NotAfter          time.Time
	Base64EncodedData string
}

func ParseCACerts(data []byte) ([]CACert, error) {
	var der []byte
	var blocks [][]byte
	for {
		var block *pem.Block
		var rem []byte
		block, rem = pem.Decode(data)
		if block == nil {
			break
		}

		der = append(der, block.Bytes...)

		blocks = append(blocks, []byte(strings.TrimSpace(string(data[:len(data)-len(rem)]))))
		data = rem
	}

	certs, err := x509.ParseCertificates(der)
	if err != nil {
		return nil, err
	}

	result := make([]CACert, len(certs))
	for i, cert := range certs {
		sum := sha1.Sum(certs[i].Raw)
		result[i] = CACert{
			Fingerprint:       strings.ToLower(hex.EncodeToString(sum[:])),
			Issuer:            cert.Issuer.String(),
			Subject:           cert.Subject.String(),
			NotBefore:         cert.NotBefore,
			NotAfter:          cert.NotAfter,
			Base64EncodedData: base64.StdEncoding.EncodeToString(blocks[i]),
		}
	}

	return result, nil
}

// Add returns existing with any certs from toAdd whose fingerprint is not
// already present appended. Existing order and entries are preserved.
func Add(existing, toAdd []CACert) []CACert {
	seen := make(map[string]struct{}, len(existing))
	for _, c := range existing {
		seen[c.Fingerprint] = struct{}{}
	}
	result := existing
	for _, c := range toAdd {
		if _, ok := seen[c.Fingerprint]; !ok {
			result = append(result, c)
		}
	}
	return result
}

// Remove returns existing with any certs whose fingerprint appears in toRemove
// filtered out.
func Remove(existing, toRemove []CACert) []CACert {
	drop := make(map[string]struct{}, len(toRemove))
	for _, c := range toRemove {
		drop[c.Fingerprint] = struct{}{}
	}
	var result []CACert
	for _, c := range existing {
		if _, ok := drop[c.Fingerprint]; !ok {
			result = append(result, c)
		}
	}
	return result
}

// EncodeCACerts encodes a slice of CACerts as a concatenated PEM bundle.
// It is the inverse of ParseCACerts. Returns nil for an empty slice so
// callers can treat nil as "no certificates configured".
func EncodeCACerts(certs []CACert) ([]byte, error) {
	if len(certs) == 0 {
		return nil, nil
	}
	out := make([][]byte, 0, len(certs))
	for _, c := range certs {
		data, err := base64.StdEncoding.DecodeString(c.Base64EncodedData)
		if err != nil {
			return nil, err
		}
		out = append(out, data)
	}
	return bytes.Join(out, []byte("\n")), nil
}
