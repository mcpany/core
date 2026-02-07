// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

// ValidateConfigRequest represents the request body for config validation.
//
// Fields:
//   Content (string): The raw YAML/JSON content of the configuration file. Must be a non-empty string.
type ValidateConfigRequest struct {
	Content string `json:"content"`
}

// ValidateConfigResponse represents the response body for config validation.
//
// Fields:
//   Valid (bool): Indicates whether the configuration is valid.
//   Errors ([]string): A list of validation errors, if any. Empty if Valid is true.
type ValidateConfigResponse struct {
	Valid bool `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

// ValidationResult represents the result of the validation logic.
//
// Fields:
//   Valid (bool): True if the configuration is valid.
//   Errors ([]string): Contains a list of error messages if validation failed.
type ValidationResult struct {
	Valid bool
	Errors []string
}
