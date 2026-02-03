// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

// contextKey is a custom type for context keys to prevent collisions.
type contextKey string

// HTTPRequestContextKey is the context key for the HTTP request.
const HTTPRequestContextKey contextKey = "http.request"

// TraceIDKey is the context key for the Trace ID.
const TraceIDKey contextKey = "trace.id"

// TraceIDCaptureKey is the context key for capturing the Trace ID via a pointer.
const TraceIDCaptureKey contextKey = "trace.id.capture"
