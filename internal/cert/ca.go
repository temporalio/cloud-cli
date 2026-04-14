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

// EncodeCACerts encodes a slice of CACerts as a concatenated PEM bundle.
// It is the inverse of ParseCACerts.
func EncodeCACerts(certs []CACert) ([]byte, error) {
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
