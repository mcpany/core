// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"net/http"

	"github.com/alexliesenfeld/health"
	healthChecker "github.com/mcpany/core/server/pkg/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// HTTPClientWrapper wraps an `*http.Client` to adapt it to the.
//
// Summary: wraps an `*http.Client` to adapt it to the.
type HTTPClientWrapper struct {
	*http.Client
	config *configv1.UpstreamServiceConfig
	// checker is cached to avoid recreation overhead on every health check.
	checker health.Checker
}

// NewHTTPClientWrapper creates a new HTTPClientWrapper.
//
// Summary: creates a new HTTPClientWrapper.
//
// Parameters:
//   - client: *http.Client. The client.
//   - config: *configv1.UpstreamServiceConfig. The config.
//   - checker: health.Checker. The checker.
//
// Returns:
//   - *HTTPClientWrapper: The *HTTPClientWrapper.
func NewHTTPClientWrapper(client *http.Client, config *configv1.UpstreamServiceConfig, checker health.Checker) *HTTPClientWrapper {
	// If no checker is provided, create a new one (backward compatibility or standalone usage).
	if checker == nil {
		checker = healthChecker.NewChecker(config)
	}
	return &HTTPClientWrapper{
		Client:  client,
		config:  config,
		checker: checker,
	}
}

// IsHealthy checks the health of the upstream service by making a request to the configured health check endpoint.
//
// Summary: checks the health of the upstream service by making a request to the configured health check endpoint.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//
// Returns:
//   - bool: The bool.
func (w *HTTPClientWrapper) IsHealthy(ctx context.Context) bool {
	if w.checker == nil {
		return true // No health check configured, assume healthy.
	}
	return w.checker.Check(ctx).Status == health.StatusUp
}

// Close is a no-op for the wrapper as it does not own the http.Client.
//
// Summary: is a no-op for the wrapper as it does not own the http.Client.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (w *HTTPClientWrapper) Close() error {
	return nil
}
