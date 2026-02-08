// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

// ValidateConfigRequest represents the request body for config validation.
//
// It contains the raw configuration content to be validated.
type ValidateConfigRequest struct {
	// Content is the raw YAML/JSON content of the configuration file.
	Content string `json:"content"`
}

// ValidateConfigResponse represents the response body for config validation.
//
// It indicates whether the configuration is valid and lists any errors found.
type ValidateConfigResponse struct {
	// Valid indicates whether the configuration is valid.
	Valid bool `json:"valid"`

	// Errors is a list of validation errors, if any.
	Errors []string `json:"errors,omitempty"`
}

// ValidationResult represents the result of the validation logic.
//
// It is used internally to decouple handler from response format if needed.
type ValidationResult struct {
	// Valid is true if the configuration is valid.
	Valid bool

	// Errors contains a list of error messages if validation failed.
	Errors []string
}
