package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
)

type TlsConfig struct {
	Enable bool   `json:"enable"`
	Verify bool   `json:"verify"`
	CA     string `json:"ca"`
	Cert   string `json:"cert"`
	Key    string `json:"key"`
}

func (t TlsConfig) NewTLSConfig() (*tls.Config, error) {

	cert, err := tls.LoadX509KeyPair(t.Cert, t.Key)
	if err != nil {
		return nil, fmt.Errorf("error parsing X509 certificate/key pair: %v", zap.Error(err))
	}
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("error parsing certificate: %v", zap.Error(err))
	}

	// Create TLSConfig
	// We will determine the cipher suites that we prefer.
	config := tls.Config{
		Certificates:       []tls.Certificate{cert},
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
		ClientAuth:         tls.NoClientCert,
	}

	// Require client certificates as needed
	if t.Verify {
		config.InsecureSkipVerify = false
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}
	// Add in CAs if applicable.
	if t.CA != "" {
		rootPEM, err := ioutil.ReadFile(t.CA)
		if err != nil || rootPEM == nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		ok := pool.AppendCertsFromPEM([]byte(rootPEM))
		if !ok {
			return nil, fmt.Errorf("failed to parse root ca certificate")
		}
		config.ClientCAs = pool
	}

	return &config, nil
}
