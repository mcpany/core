// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

// ValidateConfigRequest represents the request body for config validation.
//
// Payload containing the configuration content to be validated.
//
// Fields:
//   - Content: string. The raw YAML/JSON content of the configuration file.
type ValidateConfigRequest struct {
	Content string `json:"content"`
}

// ValidateConfigResponse represents the response body for config validation.
//
// Response payload indicating the result of the validation.
//
// Fields:
//   - Valid: bool. Indicates whether the configuration is valid.
//   - Errors: []string. A list of validation errors, if any.
type ValidateConfigResponse struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

// ValidationResult represents the result of the validation logic.
//
// Internal representation of the validation outcome. This is used internally to decouple handler from response format if needed.
//
// Fields:
//   - Valid: bool. True if the configuration is valid.
//   - Errors: []string. A list of error messages if validation failed.
type ValidationResult struct {
	Valid  bool
	Errors []string
}
