package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
)

// NewHTTPClientWithTLS creates a new *http.Client configured with the specified
// TLS settings. It supports setting a custom CA certificate, a client
// certificate and key, the server name for SNI, and skipping verification.
//
// It also configures the client with a SafeDialer to prevent SSRF attacks against
// cloud metadata services (LinkLocal addresses) and optionally private networks.
//
// Parameters:
//   - tlsConfig: The TLS settings to apply to the HTTP client's transport.
//
// Returns:
//   - *http.Client: A configured *http.Client.
//   - error: An error if the TLS configuration is invalid or files cannot be read.
func NewHTTPClientWithTLS(tlsConfig *configv1.TLSConfig) (*http.Client, error) {
	var tlsClientConfig *tls.Config

	if tlsConfig != nil {
		tlsClientConfig = &tls.Config{
			ServerName:         tlsConfig.GetServerName(),
			InsecureSkipVerify: tlsConfig.GetInsecureSkipVerify(), //nolint:gosec
		}

		if tlsConfig.GetCaCertPath() != "" {
			if err := validation.IsSecurePath(tlsConfig.GetCaCertPath()); err != nil {
				return nil, fmt.Errorf("invalid CA certificate path: %w", err)
			}
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
			if err := validation.IsSecurePath(tlsConfig.GetClientCertPath()); err != nil {
				return nil, fmt.Errorf("invalid client certificate path: %w", err)
			}
			if err := validation.IsSecurePath(tlsConfig.GetClientKeyPath()); err != nil {
				return nil, fmt.Errorf("invalid client key path: %w", err)
			}
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
	if os.Getenv("MCPANY_DENY_PRIVATE_UPSTREAM") == TrueStr {
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
