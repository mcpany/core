// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

// ValidateConfigRequest represents the request payload for the configuration validation endpoint.
type ValidateConfigRequest struct {
	// Content is the raw string content of the configuration file (YAML or JSON).
	// This field is required.
	Content string `json:"content"`
}

// ValidateConfigResponse represents the JSON response returned by the validation endpoint.
type ValidateConfigResponse struct {
	// Valid indicates whether the submitted configuration is syntactically and semantically correct.
	Valid bool `json:"valid"`
	// Errors contains a list of validation error messages if the configuration is invalid.
	// If Valid is true, this list will be empty or omitted.
	Errors []string `json:"errors,omitempty"`
}

// ValidationResult represents the internal result of the validation logic.
// This is used internally to decouple handler logic from the response format if needed.
type ValidationResult struct {
	// Valid indicates if validation passed.
	Valid bool
	// Errors contains any validation error messages.
	Errors []string
}
