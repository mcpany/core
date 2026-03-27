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

// HTTPClientWrapper wraps an `*http.Client` to adapt it to the
// `pool.ClosableClient` interface. This allows HTTP clients to be managed by a
// connection pool, which can help control the number of concurrent connections
// and reuse them where appropriate.
//
// Summary. Represents a HTTPClientWrapper.
type HTTPClientWrapper struct {
	*http.Client
	config *configv1.UpstreamServiceConfig
	// checker is cached to avoid recreation overhead on every health check.
	checker health.Checker
}

// NewHTTPClientWrapper provides newhttpclientwrapper functionality.
//
// Summary: NewHTTPClientWrapper.
//
// Parameters.
//   - client: The parameter.
//   - config: The parameter.
//   - checker: The parameter.
//
// Returns.
//   - result: The result.
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

// IsHealthy provides ishealthy functionality.
//
// Summary: IsHealthy.
//
// Parameters.
//   - ctx: The parameter.
//
// Returns.
//   - result: The result.
func (w *HTTPClientWrapper) IsHealthy(ctx context.Context) bool {
	if w.checker == nil {
		return true // No health check configured, assume healthy.
	}
	return w.checker.Check(ctx).Status == health.StatusUp
}

// Close provides close functionality.
//
// Summary: Close.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (w *HTTPClientWrapper) Close() error {
	return nil
}
