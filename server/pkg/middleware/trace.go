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

// WithTraceContext returns a new context with trace information. Summary: Injects trace, span, and parent IDs into the context. Parameters: - ctx: context.Context. The parent context. - traceID: string. The unique identifier for the trace. - spanID: string. The unique identifier for the current span. - parentID: string. The unique identifier for the parent span (optional). Returns: - context.Context: The new context with trace information attached.
//
// Summary: WithTraceContext returns a new context with trace information. Summary: Injects trace, span, and parent IDs into the context. Parameters: - ctx: context.Context. The parent context. - traceID: string. The unique identifier for the trace. - spanID: string. The unique identifier for the current span. - parentID: string. The unique identifier for the parent span (optional). Returns: - context.Context: The new context with trace information attached.
//
// Parameters:
//   - ctx (context.Context): The context for managing request lifecycle and cancellation.
//   - _ (traceID): An unnamed parameter of type traceID.
//   - _ (spanID): An unnamed parameter of type spanID.
//   - parentID (string): The unique identifier used to reference the parent resource.
//
// Returns:
//   - (context.Context): The resulting context.Context object containing the requested data.
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

// GetTraceID returns the trace ID from the context. Summary: Retrieves the trace ID from the context. Parameters: - ctx: context.Context. The context to check. Returns: - string: The trace ID if present, otherwise an empty string.
//
// Summary: GetTraceID returns the trace ID from the context. Summary: Retrieves the trace ID from the context. Parameters: - ctx: context.Context. The context to check. Returns: - string: The trace ID if present, otherwise an empty string.
//
// Parameters:
//   - ctx (context.Context): The context for managing request lifecycle and cancellation.
//
// Returns:
//   - (string): A string value representing the operation's result.
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

// GetSpanID returns the span ID from the context. Summary: Retrieves the span ID from the context. Parameters: - ctx: context.Context. The context to check. Returns: - string: The span ID if present, otherwise an empty string.
//
// Summary: GetSpanID returns the span ID from the context. Summary: Retrieves the span ID from the context. Parameters: - ctx: context.Context. The context to check. Returns: - string: The span ID if present, otherwise an empty string.
//
// Parameters:
//   - ctx (context.Context): The context for managing request lifecycle and cancellation.
//
// Returns:
//   - (string): A string value representing the operation's result.
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

// GetParentID returns the parent span ID from the context. Summary: Retrieves the parent span ID from the context. Parameters: - ctx: context.Context. The context to check. Returns: - string: The parent ID if present, otherwise an empty string.
//
// Summary: GetParentID returns the parent span ID from the context. Summary: Retrieves the parent span ID from the context. Parameters: - ctx: context.Context. The context to check. Returns: - string: The parent ID if present, otherwise an empty string.
//
// Parameters:
//   - ctx (context.Context): The context for managing request lifecycle and cancellation.
//
// Returns:
//   - (string): A string value representing the operation's result.
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
