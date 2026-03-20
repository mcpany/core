// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// ValidateConfigRequest represents the request body for config validation.
//
// Summary: Request payload for config validation.
//
// Parameters:
//   - Content (string): The raw YAML/JSON content of the configuration file.
//
// Side Effects:
//   - None.
package rest

type ValidateConfigRequest struct {
	// Content is the raw YAML/JSON content of the configuration file.
	// ValidateConfigResponse represents the response body for config validation.
	//
	// Summary: Response payload for config validation.
	//
	// Parameters:
	//   - Valid (bool): Indicates whether the configuration is valid.
	//   - Errors ([]string): A list of validation errors, if any.
	//
	// Side Effects:
	//   - None.
	Content string `json:"content"`
}

type ValidateConfigResponse struct {
	// Valid indicates whether the configuration is valid.
	Valid bool `json:"valid"`
	// Errors is a list of validation errors, if any.
	// ValidationResult represents the result of the validation logic.
	//
	// Summary: Internal result of validation logic.
	//
	// Parameters:
	//   - Valid (bool): True if the configuration is valid.
	//   - Errors ([]string): A list of error messages if validation failed.
	//
	// Side Effects:
	//   - None.
	Errors []string `json:"errors,omitempty"`
}

type ValidationResult struct {
	// Valid is true if the configuration is valid.
	Valid bool
	// Errors contains a list of error messages if validation failed.
	Errors []string
}
