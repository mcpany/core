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

// Summary: WithTraceContext executes the operation.
//
// Parameters:
//   - ctx context.Context: Input parameter.
//   - traceID: Input parameter.
//   - spanID: Input parameter.
//   - parentID string: Input parameter.
//
// Returns:
//   - context.Context {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
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
// Summary: Retrieves the trace ID from the context.
//
// Parameters:
//   - ctx: context.Context. The context to check.
//
// Returns:
//   - string: The trace ID if present, otherwise an empty string.
// Summary: GetTraceID executes the operation.
//
// Parameters:
//   - ctx context.Context: Input parameter.
//
// Returns:
//   - string {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func GetTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}

// GetSpanID returns the span ID from the context.
//
// Summary: Retrieves the span ID from the context.
//
// Parameters:
//   - ctx: context.Context. The context to check.
//
// Returns:
//   - string: The span ID if present, otherwise an empty string.
// Summary: GetSpanID executes the operation.
//
// Parameters:
//   - ctx context.Context: Input parameter.
//
// Returns:
//   - string {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func GetSpanID(ctx context.Context) string {
	if v, ok := ctx.Value(spanIDKey).(string); ok {
		return v
	}
	return ""
}

// GetParentID returns the parent span ID from the context.
//
// Summary: Retrieves the parent span ID from the context.
//
// Parameters:
//   - ctx: context.Context. The context to check.
//
// Returns:
//   - string: The parent ID if present, otherwise an empty string.
// Summary: GetParentID executes the operation.
//
// Parameters:
//   - ctx context.Context: Input parameter.
//
// Returns:
//   - string {
: Result of the operation.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func GetParentID(ctx context.Context) string {
	if v, ok := ctx.Value(parentIDKey).(string); ok {
		return v
	}
	return ""
}
