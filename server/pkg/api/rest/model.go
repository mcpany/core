// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

// ValidateConfigRequest represents the request body for config validation.
//
// Summary: Request payload for config validation.
// ValidateConfigResponse represents the response body for config validation.
//
// Summary: Response payload for config validation.
//
// Parameters:
//   - Valid (bool): Indicates whether the configuration is valid.
// ValidationResult represents the result of the validation logic.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//
// Summary: Internal result of validation logic.
//
// Parameters:
//   - Valid (bool): True if the configuration is valid.
//   - Errors ([]string): A list of error messages if validation failed.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type ValidationResult struct {
	// Valid is true if the configuration is valid.
	Valid bool
	// Errors contains a list of error messages if validation failed.
	Errors []string
//   - None.
// Side Effects:
//   - None.
// Errors:
//   - None.
// Returns:
//   - None.
// Parameters:
// Summary: Response payload for config validation.
//
// ValidateConfigResponse represents the response body for config validation.
// Summary: Request payload for config validation.
//
// ValidateConfigRequest represents the request body for config validation.
}
