/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
//
// Parameters:
//   - client: The HTTP client to be wrapped.
//   - config: The upstream service configuration, used for health checks.
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
