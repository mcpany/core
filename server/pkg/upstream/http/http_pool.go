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

	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

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

	if mtlsConfig := config.GetUpstreamAuthentication().GetMtls(); mtlsConfig != nil {
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

	dialer := util.NewSafeDialer()
	if os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES") == util.TrueStr {
		dialer.AllowLoopback = true
	}
	if os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES") == util.TrueStr {
		dialer.AllowPrivate = true
	}

	sharedClient := &http.Client{
		Transport: otelhttp.NewTransport(&http.Transport{
			TLSClientConfig: tlsConfig,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// Enforce timeout for the connection establishment
				ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
				defer cancel()
				return dialer.DialContext(ctx, network, addr)
			},
			MaxIdleConns:        maxSize,
			MaxIdleConnsPerHost: maxSize,
		}),
	}

	factory := func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return client.NewHTTPClientWrapper(
			sharedClient,
			config,
		), nil
	}
	return pool.New(factory, minSize, maxSize, idleTimeout, false)
}
