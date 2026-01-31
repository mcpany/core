// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

// ValidateConfigRequest is the Data Transfer Object (DTO) for the configuration validation endpoint.
//
// It encapsulates the raw configuration data that the client wishes to validate against
// the server's schema and logic.
type ValidateConfigRequest struct {
	// Content contains the raw configuration string in YAML or JSON format.
	// This field is required and must contain a valid configuration structure.
	Content string `json:"content"`
}

// ValidateConfigResponse is the Data Transfer Object (DTO) for the validation response.
//
// It communicates the result of the validation operation, including a boolean success flag
// and a list of specific error messages if the validation failed.
type ValidateConfigResponse struct {
	// Valid indicates whether the provided configuration passed all validation checks.
	Valid bool `json:"valid"`
	// Errors is a slice of descriptive error messages returned if validation failed.
	// This field is omitted from the JSON output if the configuration is valid.
	Errors []string `json:"errors,omitempty"`
}

// ValidationResult represents the result of the validation logic.
// This is used internally to decouple handler from response format if needed.
type ValidationResult struct {
	Valid  bool
	Errors []string
}
