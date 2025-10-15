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
	"net/http"

	"github.com/mcpxy/core/pkg/client"
	"github.com/mcpxy/core/pkg/pool"
)

// NewHttpPool creates a new connection pool for HTTP clients. It configures the
// pool with a factory function that provides new instances of http.Client,
// wrapped in an HttpClientWrapper.
//
// minSize is the initial number of clients to create.
// maxSize is the maximum number of clients the pool can hold.
// idleTimeout is the duration after which an idle client may be closed (not currently implemented).
// It returns a new HTTP client pool or an error if the pool cannot be created.
func NewHttpPool(
	minSize, maxSize, idleTimeout int,
) (pool.Pool[*client.HttpClientWrapper], error) {
	factory := func(ctx context.Context) (*client.HttpClientWrapper, error) {
		return &client.HttpClientWrapper{Client: &http.Client{}}, nil
	}
	return pool.New(factory, minSize, maxSize, idleTimeout)
}
