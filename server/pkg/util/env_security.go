// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"log/slog"
	"os"
	"strings"
)

// IsEnvVarAllowed checks if an environment variable is allowed to be accessed
// by the configuration system.
//
// Security Policy:
// 1. Block `MCPANY_*` variables by default to prevent exfiltration of server secrets
//    (like MCPANY_API_KEY, MCPANY_DB_DSN) via configuration injection.
// 2. Allow explicitly whitelisted variables via `MCPANY_ALLOWED_ENV` (comma-separated).
// 3. In Strict Mode (`MCPANY_STRICT_ENV_MODE=true`), block ALL variables unless whitelisted.
//
// Parameters:
//   - name: The name of the environment variable to check.
//
// Returns:
//   - bool: True if the environment variable is allowed, false otherwise.
func IsEnvVarAllowed(name string) bool {
	// 1. Check Allowlist
	allowedEnv := os.Getenv("MCPANY_ALLOWED_ENV")
	if allowedEnv != "" {
		parts := strings.Split(allowedEnv, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			// Simple wildcard support at the end (e.g. "AWS_*")
			if strings.HasSuffix(trimmed, "*") {
				prefix := strings.TrimSuffix(trimmed, "*")
				if strings.HasPrefix(name, prefix) {
					return true
				}
			}
			if trimmed == name {
				return true
			}
		}
	}

	// 2. Check Blocklist (MCPANY_*)
	// We block these to protect server internal configuration from being leaked
	// if a user runs a shared malicious configuration.
	if strings.HasPrefix(strings.ToUpper(name), "MCPANY_") {
		// Log a warning for visibility (at debug level to avoid spam, or info)
		// logging.GetLogger().Warn("Blocked access to protected environment variable", "name", name)
		return false
	}

	// 3. Strict Mode Check
	strictMode := os.Getenv("MCPANY_STRICT_ENV_MODE") == "true"
	if strictMode {
		slog.Warn("Blocked access to environment variable in strict mode", "name", name)
		return false
	}

	// Default: Allow
	return true
}
