// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import "errors"

// ErrToolNotFound represents the public ErrToolNotFound entity.
//
// Summary: Defines the structured data model representing a tool not found.
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
var ErrToolNotFound = errors.New("unknown tool")
