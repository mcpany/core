// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
)

type traceContextKey string

const (
	traceIDKey  traceContextKey = "trace_id"
	spanIDKey   traceContextKey = "span_id"
	parentIDKey traceContextKey = "parent_id"
)

// WithTraceContext returns a new context with trace information.
//
// Parameters:
//   - ctx (context.Context): The parent context.
//   - traceID (string): The trace ID to inject.
//   - spanID (string): The span ID to inject.
//   - parentID (string): The parent span ID to inject (optional).
//
// Returns:
//   - context.Context: The new context with trace values.
func WithTraceContext(ctx context.Context, traceID, spanID, parentID string) context.Context {
	ctx = context.WithValue(ctx, traceIDKey, traceID)
	ctx = context.WithValue(ctx, spanIDKey, spanID)
	if parentID != "" {
		ctx = context.WithValue(ctx, parentIDKey, parentID)
	}
	return ctx
}

// GetTraceID returns the trace ID from the context.
//
// Parameters:
//   - ctx (context.Context): The context to retrieve the trace ID from.
//
// Returns:
//   - string: The trace ID, or empty string if not found.
func GetTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}

// GetSpanID returns the span ID from the context.
//
// Parameters:
//   - ctx (context.Context): The context to retrieve the span ID from.
//
// Returns:
//   - string: The span ID, or empty string if not found.
func GetSpanID(ctx context.Context) string {
	if v, ok := ctx.Value(spanIDKey).(string); ok {
		return v
	}
	return ""
}

// GetParentID returns the parent span ID from the context.
//
// Parameters:
//   - ctx (context.Context): The context to retrieve the parent span ID from.
//
// Returns:
//   - string: The parent span ID, or empty string if not found.
func GetParentID(ctx context.Context) string {
	if v, ok := ctx.Value(parentIDKey).(string); ok {
		return v
	}
	return ""
}
