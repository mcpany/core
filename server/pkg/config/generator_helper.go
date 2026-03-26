// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import "strings"

<<<<<<< HEAD
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
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
func normalizeHTTPMethod(input string) string {
	upper := strings.ToUpper(strings.TrimSpace(input))
	switch upper {
	case "GET", "POST", "PUT", "DELETE", "PATCH":
		return "HTTP_METHOD_" + upper
	default:
		return input
	}
}
