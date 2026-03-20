// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// Summary: ErrToolNotFound is returned when a requested tool cannot be found.
//
// Side Effects:
//   - None.
package tool

import "errors"

var ErrToolNotFound = errors.New("unknown tool")
