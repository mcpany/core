// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

// ValidateConfigRequest encapsulates the payload for a configuration validation request.
type ValidateConfigRequest struct {
	// Content is the raw YAML or JSON content of the configuration file to be validated.
	// It must be a non-empty string.
	Content string `json:"content"`
}

// ValidateConfigResponse defines the structure of the validation response.
type ValidateConfigResponse struct {
	// Valid is true if the configuration passes all schema and semantic checks.
	Valid bool `json:"valid"`
	// Errors contains a list of descriptive error messages if validation fails.
	// This field is omitted if the configuration is valid.
	Errors []string `json:"errors,omitempty"`
}

// ValidationResult holds the internal result of a validation operation.
// This structure decouples the core validation logic from the REST API response format.
type ValidationResult struct {
	// Valid indicates if the validation was successful.
	Valid bool
	// Errors is a list of issues found during validation.
	Errors []string
}
