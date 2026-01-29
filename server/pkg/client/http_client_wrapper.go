package client

import (
	"context"
	"net/http"

	"github.com/alexliesenfeld/health"
	healthChecker "github.com/mcpany/core/server/pkg/health"
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
// It accepts a shared health checker to avoid creating a new one for every client.
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
// ctx is the context for the request.
//
// Returns true if successful.
func (w *HTTPClientWrapper) IsHealthy(ctx context.Context) bool {
	if w.checker == nil {
		return true // No health check configured, assume healthy.
	}
	return w.checker.Check(ctx).Status == health.StatusUp
}

// Close is a no-op for the wrapper as it does not own the http.Client.
// The owner of the http.Client (e.g., the pool manager) is responsible for closing idle connections
// on the shared Transport when the service is shut down.
//
// Previously, this called CloseIdleConnections on the shared transport, which would negatively
// impact other concurrent requests sharing the same Transport.
func (w *HTTPClientWrapper) Close() error {
	return nil
}
