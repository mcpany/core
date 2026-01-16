// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http //nolint:revive,nolintlint // Package name 'http' is intentional for this directory structure.

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"os"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/client"
	healthChecker "github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/util"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type httpPool struct {
	pool.Pool[*client.HTTPClientWrapper]
	transport *http.Transport
}

// Close closes the connection pool and the idle connections.
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

	dialer := util.NewSafeDialerFromEnv()
	dialer.Dialer = &net.Dialer{
		Timeout: 30 * time.Second,
	}

	baseTransport := &http.Transport{
		TLSClientConfig:     tlsConfig,
		DialContext:         dialer.DialContext,
		MaxIdleConns:        maxSize,
		MaxIdleConnsPerHost: maxSize,
	}

	sharedClient := &http.Client{
		Transport: otelhttp.NewTransport(baseTransport),
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
