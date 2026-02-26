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
// Summary: Returns a new context with trace information.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - _ (traceID): Ignored.
//   - _ (spanID): Ignored.
//   - parentID (string): The parent id.
//
// Returns:
//   - context.Context: The result.
//
// Side Effects:
//   - None.
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
// Summary: Returns the trace ID from the context.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//
// Returns:
//   - string: The result.
//
// Side Effects:
//   - None.
func GetTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}

// GetSpanID returns the span ID from the context.
//
// Summary: Returns the span ID from the context.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//
// Returns:
//   - string: The result.
//
// Side Effects:
//   - None.
func GetSpanID(ctx context.Context) string {
	if v, ok := ctx.Value(spanIDKey).(string); ok {
		return v
	}
	return ""
}

// GetParentID returns the parent span ID from the context.
//
// Summary: Returns the parent span ID from the context.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//
// Returns:
//   - string: The result.
//
// Side Effects:
//   - None.
func GetParentID(ctx context.Context) string {
	if v, ok := ctx.Value(parentIDKey).(string); ok {
		return v
	}
	return ""
}
