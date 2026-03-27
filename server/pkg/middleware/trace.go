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

// WithTraceContext provides withtracecontext functionality.
//
// Summary: WithTraceContext.
//
// Parameters.
//   - ctx: The parameter.
//   - traceID: The parameter.
//   - spanID: The parameter.
//   - parentID: The parameter.
//
// Returns.
//   - result: The result.
func WithTraceContext(ctx context.Context, traceID, spanID, parentID string) context.Context {
	ctx = context.WithValue(ctx, traceIDKey, traceID)
	ctx = context.WithValue(ctx, spanIDKey, spanID)
	if parentID != "" {
		ctx = context.WithValue(ctx, parentIDKey, parentID)
	}
	return ctx
}

// GetTraceID provides gettraceid functionality.
//
// Summary: GetTraceID.
//
// Parameters.
//   - ctx: The parameter.
//
// Returns.
//   - result: The result.
func GetTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}

// GetSpanID provides getspanid functionality.
//
// Summary: GetSpanID.
//
// Parameters.
//   - ctx: The parameter.
//
// Returns.
//   - result: The result.
func GetSpanID(ctx context.Context) string {
	if v, ok := ctx.Value(spanIDKey).(string); ok {
		return v
	}
	return ""
}

// GetParentID provides getparentid functionality.
//
// Summary: GetParentID.
//
// Parameters.
//   - ctx: The parameter.
//
// Returns.
//   - result: The result.
func GetParentID(ctx context.Context) string {
	if v, ok := ctx.Value(parentIDKey).(string); ok {
		return v
	}
	return ""
}
