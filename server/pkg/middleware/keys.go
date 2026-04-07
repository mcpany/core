// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

// contextKey is a custom type for context keys to prevent collisions.
type contextKey string

// Summary: HTTPRequestContextKey is a global constant or variable.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
const HTTPRequestContextKey contextKey = "http.request"
