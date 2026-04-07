// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

// Summary: ValidateConfigRequest represents a data structure.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
type ValidateConfigRequest struct {
	// Content is the raw YAML/JSON content of the configuration file.
	Content string `json:"content"`
}

// ValidateConfigResponse represents the response body for config validation.
//
// Summary: Response payload for config validation.
//
// Parameters:
// Summary: ValidateConfigResponse represents a data structure.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
type ValidateConfigResponse struct {
	// Valid indicates whether the configuration is valid.
	Valid bool `json:"valid"`
	// Errors is a list of validation errors, if any.
	Errors []string `json:"errors,omitempty"`
}

// ValidationResult represents the result of the validation logic.
//
// Summary: Internal result of validation logic.
//
// Parameters:
//   - Valid (bool): True if the configuration is valid.
// Summary: ValidationResult represents a data structure.
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
type ValidationResult struct {
	// Valid is true if the configuration is valid.
	Valid bool
	// Errors contains a list of error messages if validation failed.
	Errors []string
}
