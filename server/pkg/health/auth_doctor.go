// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"os"
)

// CheckAuth performs health checks for authentication configuration.
//
// Summary: performs health checks for authentication configuration.
//
// Parameters:
//   None.
//
// Returns:
//   - map[string]CheckResult: The map[string]CheckResult.
func CheckAuth() map[string]CheckResult {
	results := make(map[string]CheckResult)

	// Check common API Keys
	// We can extend this to check for valid formats if needed.
	keys := []string{
		"ANTHROPIC_API_KEY",
		"OPENAI_API_KEY",
		"GEMINI_API_KEY",
	}

	for _, k := range keys {
		val := os.Getenv(k)
		if val != "" {
			// Do not leak the key, just say present
			masked := "Present"
			if len(val) > 4 {
				masked = "Present (..." + val[len(val)-4:] + ")"
			}
			results[k] = CheckResult{
				Status:  "ok",
				Message: masked,
			}
		} else {
			results[k] = CheckResult{
				Status:  "missing",
				Message: "Environment variable not set",
			}
		}
	}

	// Check OAuth configuration (basic check if env vars exist)
	// Example: GOOGLE_CLIENT_ID, GITHUB_CLIENT_ID
	oauthProviders := []string{"GOOGLE", "GITHUB"}
	for _, p := range oauthProviders {
		clientID := os.Getenv(p + "_CLIENT_ID")
		clientSecret := os.Getenv(p + "_CLIENT_SECRET")
		switch {
		case clientID != "" && clientSecret != "":
			results["oauth_"+p] = CheckResult{
				Status:  "ok",
				Message: "Configured",
			}
		case clientID != "" || clientSecret != "":
			results["oauth_"+p] = CheckResult{
				Status:  "warning",
				Message: "Partial configuration (missing ID or Secret)",
			}
		default:
			results["oauth_"+p] = CheckResult{
				Status:  "info",
				Message: "Not configured",
			}
		}
	}

	return results
}
