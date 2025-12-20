// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// NewHTTPClientWithTLS creates a new *http.Client configured with the specified
// TLS settings. It supports setting a custom CA certificate, a client
// certificate and key, the server name for SNI, and skipping verification. If
// the provided tlsConfig is nil, it returns http.DefaultClient.
//
// tlsConfig contains the TLS settings to apply to the HTTP client's transport.
// It returns a configured *http.Client or an error if the TLS configuration
// is invalid or files cannot be read.
func NewHTTPClientWithTLS(tlsConfig *configv1.TLSConfig) (*http.Client, error) {
	// Clone defaults from http.DefaultTransport to ensure we have timeouts and pool settings.
	var baseTransport *http.Transport
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		baseTransport = t.Clone()
	} else {
		// Fallback if DefaultTransport is modified or not a *Transport (unlikely)
		baseTransport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}
	}

	baseTransport.DialContext = SafeDialContext

	if tlsConfig == nil {
		return &http.Client{
			Transport: baseTransport,
		}, nil
	}

	config := &tls.Config{
		ServerName:         tlsConfig.GetServerName(),
		InsecureSkipVerify: tlsConfig.GetInsecureSkipVerify(), //nolint:gosec
	}

	if tlsConfig.GetCaCertPath() != "" {
		caCert, err := os.ReadFile(tlsConfig.GetCaCertPath())
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("failed to append CA certs from PEM")
		}
		config.RootCAs = caCertPool
	}

	if tlsConfig.GetClientCertPath() != "" && tlsConfig.GetClientKeyPath() != "" {
		clientCert, err := tls.LoadX509KeyPair(tlsConfig.GetClientCertPath(), tlsConfig.GetClientKeyPath())
		if err != nil {
			return nil, fmt.Errorf("failed to load client key pair: %w", err)
		}
		config.Certificates = []tls.Certificate{clientCert}
	}

	baseTransport.TLSClientConfig = config

	return &http.Client{
		Transport: baseTransport,
	}, nil
}
