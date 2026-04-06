// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

// contextKey is a custom type for context keys to prevent collisions.
type contextKey string

// HTTPRequestContextKey is the context key for the HTTP request.
//
// Summary: Context key used to store the original HTTP request.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Throws/Errors:
//   - None.
//
// Side Effects:
//   - None.
const HTTPRequestContextKey contextKey = "http.request"
