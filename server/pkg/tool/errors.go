// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import "errors"

// ErrToolNotFound is returned when a requested tool cannot be found.
//
// Summary: Is returned when a requested tool cannot be found.
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
