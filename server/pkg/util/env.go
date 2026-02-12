// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"os"
	"strconv"
	"strings"
)

// parseBoolEnv parses a boolean environment variable.
// It handles "true", "1", "yes" (case-insensitive) and quoted strings like "'true'".
func parseBoolEnv(key string) bool {
	val := os.Getenv(key)
	if val == "" {
		return false
	}

	// Handle potential quotes from Docker Compose or shell
	val = strings.Trim(val, "'\"")

	b, err := strconv.ParseBool(val)
	if err != nil {
		// Fallback for simple "true" string check if ParseBool fails for some reason
		return strings.ToLower(val) == "true"
	}
	return b
}
