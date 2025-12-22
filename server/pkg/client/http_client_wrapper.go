// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"net/http"

	"github.com/alexliesenfeld/health"
	healthChecker "github.com/mcpany/core/pkg/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// HTTPClientWrapper wraps an `*http.Client` to adapt it to the
// `pool.ClosableClient` interface. This allows HTTP clients to be managed by a
// connection pool, which can help control the number of concurrent connections
// and reuse them where appropriate.
type HTTPClientWrapper struct {
	*http.Client
	config *configv1.UpstreamServiceConfig
	// checker is cached to avoid recreation overhead on every health check.
	checker health.Checker
}

// NewHTTPClientWrapper creates a new HTTPClientWrapper.
// client is the client.
// config is the config.
// Returns the result.
func NewHTTPClientWrapper(client *http.Client, config *configv1.UpstreamServiceConfig) *HTTPClientWrapper {
	return &HTTPClientWrapper{
		Client:  client,
		config:  config,
		checker: healthChecker.NewChecker(config),
	}
}

// IsHealthy checks the health of the upstream service by making a request to the configured health check endpoint.
// ctx is the context.
// Returns the result.
func (w *HTTPClientWrapper) IsHealthy(ctx context.Context) bool {
	if w.checker == nil {
		return true // No health check configured, assume healthy.
	}
	return w.checker.Check(ctx).Status == health.StatusUp
}

// Close is a no-op for the `*http.Client` wrapper. The underlying transport
// manages the connections, and closing the client itself is not typically
// necessary.
// Returns an error.
func (w *HTTPClientWrapper) Close() error {
	return nil
}
