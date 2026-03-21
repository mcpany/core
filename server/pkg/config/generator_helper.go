// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import "strings"

func normalizeHTTPMethod(input string) string {
	upper := strings.ToUpper(strings.TrimSpace(input))
	switch upper {
	case "GET", "POST", "PUT", "DELETE", "PATCH":
		return "HTTP_METHOD_" + upper
	default:
		return input
	}
}
