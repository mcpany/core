// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import "errors"

// Summary: ErrToolNotFound is returned when a requested tool cannot be found. Represents a ErrToolNotFound.
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
