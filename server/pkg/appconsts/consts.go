// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package appconsts

const (
// Summary: Name is the name of the MCP Any server. This is used in help messages and other user-facing output. Defines Name.
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
const (
	// Name is the name of the MCP Any server. This is used in help messages and
	// other user-facing output.
	// Summary: Defines Name.
	Name = "mcpany"
)

// Summary: Version is the version of the MCP Any server. This is a variable so it can be set at build time using ldflags. The default value is "dev", which is used for local development builds. Represents a Version.
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
var Version = "dev"
