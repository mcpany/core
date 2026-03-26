// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

// ValidateConfigRequest validateConfigRequest represents a validate config request.
//
// Summary: ValidateConfigRequest represents a validate config request.
type ValidateConfigRequest struct {
	// Content is the raw YAML/JSON content of the configuration file.
	Content string `json:"content"`
}

// ValidateConfigResponse validateConfigResponse represents a validate config response.
//
// Summary: ValidateConfigResponse represents a validate config response.
type ValidateConfigResponse struct {
	// Valid indicates whether the configuration is valid.
	Valid bool `json:"valid"`
	// Errors is a list of validation errors, if any.
	Errors []string `json:"errors,omitempty"`
}

// ValidationResult validationResult represents a validation result.
//
// Summary: ValidationResult represents a validation result.
type ValidationResult struct {
	// Valid is true if the configuration is valid.
	Valid bool
	// Errors contains a list of error messages if validation failed.
	Errors []string
}
