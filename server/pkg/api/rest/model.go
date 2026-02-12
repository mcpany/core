// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

// ValidateConfigRequest represents the request body for config validation.
//
// Summary: Encapsulates the input data required to validate a server configuration.
type ValidateConfigRequest struct {
	// Content is the raw YAML/JSON content of the configuration file to be validated.
	Content string `json:"content"`
}

// ValidateConfigResponse represents the response body for config validation.
//
// Summary: Encapsulates the result of a configuration validation request.
type ValidateConfigResponse struct {
	// Valid indicates whether the submitted configuration is syntactically and semantically valid.
	Valid bool `json:"valid"`
	// Errors contains a list of validation error messages if the configuration is invalid.
	Errors []string `json:"errors,omitempty"`
}

// ValidationResult represents the internal result of the validation logic.
//
// Summary: Stores the outcome of the configuration validation process for internal use.
type ValidationResult struct {
	// Valid is true if the configuration passed all checks.
	Valid bool
	// Errors contains a list of error messages describing why validation failed.
	Errors []string
}
