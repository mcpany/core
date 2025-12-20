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
// certificate and key, the server name for SNI, and skipping verification.
//
// It also configures the client with a SafeDialer to prevent SSRF attacks against
// cloud metadata services (LinkLocal addresses) and optionally private networks.
//
// tlsConfig contains the TLS settings to apply to the HTTP client's transport.
// It returns a configured *http.Client or an error if the TLS configuration
// is invalid or files cannot be read.
func NewHTTPClientWithTLS(tlsConfig *configv1.TLSConfig) (*http.Client, error) {
	var tlsClientConfig *tls.Config

	if tlsConfig != nil {
		tlsClientConfig = &tls.Config{
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
			tlsClientConfig.RootCAs = caCertPool
		}

		if tlsConfig.GetClientCertPath() != "" && tlsConfig.GetClientKeyPath() != "" {
			clientCert, err := tls.LoadX509KeyPair(tlsConfig.GetClientCertPath(), tlsConfig.GetClientKeyPath())
			if err != nil {
				return nil, fmt.Errorf("failed to load client key pair: %w", err)
			}
			tlsClientConfig.Certificates = []tls.Certificate{clientCert}
		}
	}

	// Create a SafeDialer to prevent SSRF.
	// By default, we allow private and loopback connections for upstreams (to support local services),
	// but block link-local addresses (e.g., AWS metadata service).
	dialer := NewSafeDialer()
	dialer.AllowPrivate = true
	dialer.AllowLoopback = true

	// Allow stricter security via environment variable.
	if os.Getenv("MCPANY_DENY_PRIVATE_UPSTREAM") == "true" {
		dialer.AllowPrivate = false
		dialer.AllowLoopback = false
	}

	// Start with DefaultTransport to keep Proxy settings etc.
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsClientConfig
	transport.DialContext = dialer.DialContext

	return &http.Client{
		Transport: transport,
	}, nil
}
