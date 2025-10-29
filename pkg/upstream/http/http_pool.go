/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package http

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/mcpxy/core/pkg/client"
	"github.com/mcpxy/core/pkg/pool"
	configv1 "github.com/mcpxy/core/proto/config/v1"
)

var (
	// NewHttpPool creates a new connection pool for HTTP clients. It is defined as
	// a variable to allow for easy mocking in tests.
	//
	// minSize is the initial number of clients to create.
	// maxSize is the maximum number of clients the pool can hold.
	// idleTimeout is the duration after which an idle client may be closed (not
	// currently implemented).
	// healthCheck is the configuration for the health check.
	// It returns a new HTTP client pool or an error if the pool cannot be
	// created.
	NewHttpPool = func(
		minSize, maxSize, idleTimeout int,
		healthCheck *configv1.HttpHealthCheck,
	) (pool.Pool[*client.HttpClientWrapper], error) {
		factory := func(ctx context.Context) (*client.HttpClientWrapper, error) {
			return &client.HttpClientWrapper{
				Client: &http.Client{
					Transport: &http.Transport{
						DisableKeepAlives: true,
						TLSClientConfig:   &tls.Config{MinVersion: tls.VersionTLS12},
						DialContext: (&net.Dialer{
							Timeout: 30 * time.Second,
						}).DialContext,
					},
				},
				HealthCheck: healthCheck,
			}, nil
		}
		return pool.New(factory, minSize, maxSize, idleTimeout)
	}
)
