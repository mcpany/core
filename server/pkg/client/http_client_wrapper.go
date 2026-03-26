// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"net/http"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	healthChecker "github.com/mcpany/core/server/pkg/health"
)

// HTTPClientWrapper hTTPClientWrapper represents a http client wrapper.
//
// Summary: HTTPClientWrapper represents a http client wrapper.
type HTTPClientWrapper struct {
	*http.Client
	config *configv1.UpstreamServiceConfig
	// checker is cached to avoid recreation overhead on every health check.
	checker health.Checker
}

// NewHTTPClientWrapper creates a new HTTPClientWrapper. It accepts a shared health checker to avoid creating a new one for every client.
//
// Parameters: - None.
//   - client (*http.Client): The client parameter.
//   - config (*configv1.UpstreamServiceConfig): The config parameter.
//   - checker (health.Checker): The checker parameter.
//
// Returns: - None.
//   - *HTTPClientWrapper: The resulting *HTTPClientWrapper.
//
// Errors: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Initializes NewHTTPClientWrapper operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
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

// IsHealthy checks the health of the upstream service by making a request to the configured health check endpoint. ctx is the context for the request. Returns true if successful.
//
// Parameters: - None.
//   - ctx (context.Context): The context for the request.
//
// Returns: - None.
//   - bool: True if successful, false otherwise.
//
// Errors: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Checks IsHealthy operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (w *HTTPClientWrapper) IsHealthy(ctx context.Context) bool {
	if w.checker == nil {
		return true // No health check configured, assume healthy.
	}
	return w.checker.Check(ctx).Status == health.StatusUp
}

// Close is a no-op for the wrapper as it does not own the http.Client. The owner of the http.Client (e.g., the pool manager) is responsible for closing idle connections on the shared Transport when the service is shut down. Previously, this called CloseIdleConnections on the shared transport, which would negatively impact other concurrent requests sharing the same Transport.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - error: An error if the operation fails.
//
// Errors: - None.
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes Close operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (w *HTTPClientWrapper) Close() error {
	return nil
}
