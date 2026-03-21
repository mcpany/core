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

// Summary: WithTraceContext returns a new context with trace information. Injects trace, span, and parent IDs into the context.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - traceID (string): The traceID parameter.
//   - spanID (string): The spanID parameter.
//   - parentID (string): The parentID parameter.
//
// Returns:
//   - context.Context: The resulting context.Context.
//
// Errors:
//   - None.
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

// Summary: GetTraceID returns the trace ID from the context. Retrieves the trace ID from the context.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func GetTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}

// Summary: GetSpanID returns the span ID from the context. Retrieves the span ID from the context.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func GetSpanID(ctx context.Context) string {
	if v, ok := ctx.Value(spanIDKey).(string); ok {
		return v
	}
	return ""
}

// Summary: GetParentID returns the parent span ID from the context. Retrieves the parent span ID from the context.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func GetParentID(ctx context.Context) string {
	if v, ok := ctx.Value(parentIDKey).(string); ok {
		return v
	}
	return ""
}
