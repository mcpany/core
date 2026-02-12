// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

// ValidateConfigRequest represents the request body for config validation.
//
// Summary: Request payload for config validation.
type ValidateConfigRequest struct {
	// Content is the raw YAML/JSON content of the configuration file.
	Content string `json:"content"`
}

// ValidateConfigResponse represents the response body for config validation.
//
// Summary: Response payload for config validation.
type ValidateConfigResponse struct {
	// Valid indicates whether the configuration passed all validation checks.
	Valid bool `json:"valid"`
	// Errors contains details of any validation failures encountered.
	Errors []string `json:"errors,omitempty"`
}

// ValidationResult represents the result of the validation logic.
//
// Summary: Internal result of validation logic.
type ValidationResult struct {
	// Valid indicates whether the configuration passed all validation checks.
	Valid bool
	// Errors contains details of any validation failures encountered.
	Errors []string
}
