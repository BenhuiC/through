package util

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// LoadTlsConfig load tls config with private_key ,cert and ca
func LoadTlsConfig(priKeyFile, crtFile, caFile string, insecure bool) (cfg *tls.Config, err error) {
	cert, err := tls.LoadX509KeyPair(crtFile, priKeyFile)
	if err != nil {
		return nil, err
	}
	cfg = &tls.Config{
		ClientAuth:         tls.RequireAndVerifyClientCert,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: insecure,
	}

	if len(caFile) != 0 {
		ca := x509.NewCertPool()
		caBytes, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("load ca err: %v", err)
		}
		if ok := ca.AppendCertsFromPEM(caBytes); !ok {
			return nil, fmt.Errorf("failed to parse ca %q", caFile)
		}

		cfg.ClientCAs = ca
	}

	return cfg, nil
}
