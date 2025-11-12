
package client

import (
	"context"
	"net/http"

	"github.com/alexliesenfeld/health"
	healthChecker "github.com/mcpany/core/pkg/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// HttpClientWrapper wraps an `*http.Client` to adapt it to the
// `pool.ClosableClient` interface. This allows HTTP clients to be managed by a
// connection pool, which can help control the number of concurrent connections
// and reuse them where appropriate.
type HttpClientWrapper struct {
	*http.Client
	config *configv1.UpstreamServiceConfig
}

// NewHttpClientWrapper creates a new HttpClientWrapper.
func NewHttpClientWrapper(client *http.Client, config *configv1.UpstreamServiceConfig) *HttpClientWrapper {
	return &HttpClientWrapper{
		Client: client,
		config: config,
	}
}

// IsHealthy checks the health of the upstream service by making a request to the configured health check endpoint.
func (w *HttpClientWrapper) IsHealthy(ctx context.Context) bool {
	checker := healthChecker.NewChecker(w.config)
	if checker == nil {
		return true // No health check configured, assume healthy.
	}
	return checker.Check(ctx).Status == health.StatusUp
}

// Close is a no-op for the `*http.Client` wrapper. The underlying transport
// manages the connections, and closing the client itself is not typically
// necessary.
func (w *HttpClientWrapper) Close() error {
	return nil
}
