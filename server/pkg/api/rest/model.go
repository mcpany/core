// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

// ValidateConfigRequest represents the request body for config validation.
//
// Summary: Request payload for config validation.
type ValidateConfigRequest struct {
	// Content is the raw YAML/JSON content of the configuration file.
	//
	// Type: string.
	// Description: The configuration content to validate.
	Content string `json:"content"`
}

// ValidateConfigResponse represents the response body for config validation.
//
// Summary: Response payload for config validation.
type ValidateConfigResponse struct {
	// Valid indicates whether the configuration is valid.
	//
	// Type: bool.
	// Description: True if validation passed.
	Valid bool `json:"valid"`

	// Errors is a list of validation errors, if any.
	//
	// Type: []string.
	// Description: A list of error messages describing why validation failed.
	Errors []string `json:"errors,omitempty"`
}

// ValidationResult represents the result of the validation logic.
//
// Summary: Internal result of validation logic.
type ValidationResult struct {
	// Valid is true if the configuration is valid.
	//
	// Type: bool.
	// Description: True if validation passed.
	Valid bool

	// Errors contains a list of error messages if validation failed.
	//
	// Type: []string.
	// Description: A list of error messages describing why validation failed.
	Errors []string
}
