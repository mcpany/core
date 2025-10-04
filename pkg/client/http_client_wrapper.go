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

package client

import "net/http"

// HttpClientWrapper wraps an *http.Client to adapt it to the
// pool.ClosableClient interface. This allows HTTP clients to be managed by a
// connection pool.
type HttpClientWrapper struct {
	*http.Client
}

// IsHealthy always returns true for an *http.Client. Unlike gRPC connections,
// HTTP clients are generally long-lived and do not have a persistent connection
// state to check.
func (w *HttpClientWrapper) IsHealthy() bool {
	return true
}

// Close is a no-op for the *http.Client wrapper. The underlying transport manages
// connections, and closing the client is not necessary.
func (w *HttpClientWrapper) Close() error {
	return nil
}
