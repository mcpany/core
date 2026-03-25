// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware
// WithTraceContext returns a new context with trace information.
//
// Summary: Injects trace, span, and parent IDs into the context.
//
// Parameters:
//   - ctx: context.Context. The parent context.
//   - traceID: string. The unique identifier for the trace.
//   - spanID: string. The unique identifier for the current span.
//   - parentID: string. The unique identifier for the parent span (optional).
//
// Returns:
// GetTraceID returns the trace ID from the context.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Summary: Retrieves the trace ID from the context.
//
// Parameters:
//   - ctx: context.Context. The context to check.
// GetSpanID returns the span ID from the context.
//
// Summary: Retrieves the span ID from the context.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Parameters:
//   - ctx: context.Context. The context to check.
// GetParentID returns the parent span ID from the context.
//
// Summary: Retrieves the parent span ID from the context.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Parameters:
//   - ctx: context.Context. The context to check.
//
// Returns:
//   - string: The parent ID if present, otherwise an empty string.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func GetParentID(ctx context.Context) string {
	if v, ok := ctx.Value(parentIDKey).(string); ok {
		return v
	}
	return ""
}
