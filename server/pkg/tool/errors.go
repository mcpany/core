// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// ErrToolNotFound is returned when a requested tool cannot be found.
//
// Summary: Represents a ErrToolNotFound.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
package tool

var ErrToolNotFound = errors.New("unknown tool")
