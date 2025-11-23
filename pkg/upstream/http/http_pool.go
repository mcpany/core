// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/pool"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

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
var NewHttpPool = func(
	minSize, maxSize, idleTimeout int,
	config *configv1.UpstreamServiceConfig,
) (pool.Pool[*client.HttpClientWrapper], error) {
	factory := func(ctx context.Context) (*client.HttpClientWrapper, error) {
		return client.NewHttpClientWrapper(
			&http.Client{
				Transport: &http.Transport{
					DisableKeepAlives: true,
					TLSClientConfig:   &tls.Config{MinVersion: tls.VersionTLS12},
					DialContext: (&net.Dialer{
						Timeout: 30 * time.Second,
					}).DialContext,
				},
			},
			config,
		), nil
	}
	return pool.New(factory, minSize, maxSize, idleTimeout, false)
}
