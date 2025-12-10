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
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const httpRequestKey contextKey = "http.request"

// HTTPRequestToContext adds the HTTP request to the context.
func HTTPRequestToContext(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, httpRequestKey, r)
}

// HTTPRequestFromContext extracts the HTTP request from the context.
func HTTPRequestFromContext(ctx context.Context) (*http.Request, bool) {
	r, ok := ctx.Value(httpRequestKey).(*http.Request)
	return r, ok
}
