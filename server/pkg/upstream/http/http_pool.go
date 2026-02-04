// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http //nolint:revive,nolintlint // Package name 'http' is intentional for this directory structure.

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/client"
	healthChecker "github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/util"
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
		if err := validation.IsSecurePath(mtlsConfig.GetClientCertPath()); err != nil {
			return nil, fmt.Errorf("invalid client certificate path: %w", err)
		}
		if err := validation.IsSecurePath(mtlsConfig.GetClientKeyPath()); err != nil {
			return nil, fmt.Errorf("invalid client key path: %w", err)
		}
		cert, err := tls.LoadX509KeyPair(mtlsConfig.GetClientCertPath(), mtlsConfig.GetClientKeyPath())
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}

		if err := validation.IsSecurePath(mtlsConfig.GetCaCertPath()); err != nil {
			return nil, fmt.Errorf("invalid CA certificate path: %w", err)
		}
		caCert, err := os.ReadFile(mtlsConfig.GetCaCertPath())
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	dialer := util.NewSafeDialer()
	// Allow overriding safety checks via environment variables (consistent with validation package)
	if os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS") == util.TrueStr {
		dialer.AllowLoopback = true
		dialer.AllowPrivate = true
	}

	if os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES") == util.TrueStr {
		dialer.AllowLoopback = true
	}

	if os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES") == util.TrueStr {
		dialer.AllowPrivate = true
	}

	baseTransport := &http.Transport{
		TLSClientConfig:     tlsConfig,
		DialContext:         dialer.DialContext,
		MaxIdleConns:        maxSize,
		MaxIdleConnsPerHost: maxSize,
		// Bolt: Optimize connection reuse and timeouts
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}

	clientTimeout := 30 * time.Second
	if config.GetResilience() != nil && config.GetResilience().GetTimeout() != nil {
		clientTimeout = config.GetResilience().GetTimeout().AsDuration()
	}

	sharedClient := &http.Client{
		Transport: otelhttp.NewTransport(baseTransport),
		Timeout:   clientTimeout,
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
	// Use minSize as both initialSize and maxIdleSize to preserve existing behavior where minSize was pre-filled.
	basePool, err := pool.New(factory, minSize, minSize, maxSize, idleTimeout, false)
	if err != nil {
		return nil, err
	}

	return &httpPool{
		Pool:      basePool,
		transport: baseTransport,
	}, nil
}
