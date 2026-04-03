// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

// ValidateConfigRequest represents the public ValidateConfigRequest entity.
//
// Summary: Defines the structured data model representing a config request.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type ValidateConfigRequest struct {
	// Content is the raw YAML/JSON content of the configuration file.
	Content string `json:"content"`
}

// ValidateConfigResponse represents the public ValidateConfigResponse entity.
//
// Summary: Defines the structured data model representing a config response.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type ValidateConfigResponse struct {
	// Valid indicates whether the configuration is valid.
	Valid bool `json:"valid"`
	// Errors is a list of validation errors, if any.
	Errors []string `json:"errors,omitempty"`
}

// ValidationResult represents the public ValidationResult entity.
//
// Summary: Defines the structured data model representing a result.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type ValidationResult struct {
	// Valid is true if the configuration is valid.
	Valid bool
	// Errors contains a list of error messages if validation failed.
	Errors []string
}
