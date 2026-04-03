// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

// contextKey is a custom type for context keys to prevent collisions.
type contextKey string

// HTTPRequestContextKey represents the public HTTPRequestContextKey entity.
//
// Summary: Defines the structured data model representing a request context key.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
const HTTPRequestContextKey contextKey = "http.request"
