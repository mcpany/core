// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http //nolint:revive,nolintlint // Package name 'http' is intentional for this directory structure.

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/client"
	healthChecker "github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/validation"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type httpPool struct {
	pool.Pool[*client.HTTPClientWrapper]
	transport *http.Transport
}

// Close closes the connection pool and the idle connections.
//
// Returns an error if the operation fails.
func (p *httpPool) Close() error {
	if err := p.Pool.Close(); err != nil {
		return err
	}
	p.transport.CloseIdleConnections()
	return nil
}

// NewHTTPPool creates a new connection pool for HTTP clients. It is defined as
// a variable to allow for easy mocking in tests.
//
// minSize is the initial number of clients to create.
// maxSize is the maximum number of clients the pool can hold.
// idleTimeout is the duration after which an idle client may be closed (not
// currently implemented).
// healthCheck is the configuration for the health check.
// It returns a new HTTP client pool or an error if the pool cannot be
// created.
var NewHTTPPool = func(
	minSize, maxSize int,
	idleTimeout time.Duration,
	config *configv1.UpstreamServiceConfig,
) (pool.Pool[*client.HTTPClientWrapper], error) {
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: config.GetHttpService().GetTlsConfig().GetInsecureSkipVerify(), //nolint:gosec
	}

	if mtlsConfig := config.GetUpstreamAuth().GetMtls(); mtlsConfig != nil {
		cert, err := tls.LoadX509KeyPair(mtlsConfig.GetClientCertPath(), mtlsConfig.GetClientKeyPath())
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}

		caCert, err := os.ReadFile(mtlsConfig.GetCaCertPath())
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	baseTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			if os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS") == "true" {
				return (&net.Dialer{Timeout: 30 * time.Second}).DialContext(ctx, network, addr)
			}

			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil {
				return nil, err
			}

			for _, ip := range ips {
				if err := validation.ValidateIP(ip); err != nil {
					return nil, fmt.Errorf("host %q resolves to unsafe IP %s: %w", host, ip.String(), err)
				}
			}

			dialer := &net.Dialer{Timeout: 30 * time.Second}
			for _, ip := range ips {
				conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
				if err == nil {
					return conn, nil
				}
			}
			return nil, fmt.Errorf("failed to dial any resolved IP for %s", host)
		},
		MaxIdleConns:        maxSize,
		MaxIdleConnsPerHost: maxSize,
	}

	sharedClient := &http.Client{
		Transport: otelhttp.NewTransport(baseTransport),
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			if err := validation.IsSafeURL(req.URL.String()); err != nil {
				return fmt.Errorf("unsafe redirect: %w", err)
			}
			return nil
		},
	}

	// Create a shared health checker for all clients in this pool
	checker := healthChecker.NewChecker(config)

	factory := func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return client.NewHTTPClientWrapper(
			sharedClient,
			config,
			checker,
		), nil
	}
	basePool, err := pool.New(factory, minSize, maxSize, idleTimeout, false)
	if err != nil {
		return nil, err
	}

	return &httpPool{
		Pool:      basePool,
		transport: baseTransport,
	}, nil
}
