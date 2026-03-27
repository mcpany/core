// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import "strings"

// normalizeHTTPMethod normalizes the HTTP method.
//
// Summary: Normalizes the HTTP method.
//
// Parameters:
//   - input (string): The HTTP method to normalize.
//
// Returns:
//   - string: The normalized HTTP method.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func normalizeHTTPMethod(input string) string {
	upper := strings.ToUpper(strings.TrimSpace(input))
	switch upper {
	case "GET", "POST", "PUT", "DELETE", "PATCH":
		return "HTTP_METHOD_" + upper
	default:
		return input
	}
}
