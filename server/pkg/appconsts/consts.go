// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package appconsts

const (
	// Name is the name of the MCP Any server. This is used in help messages and
	// other user-facing output.
	//
	// Summary: Defines the official, user-facing name of the MCP Any application.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: The name of the application ("mcpany")
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	Name = "mcpany"
)

// Version is the version of the MCP Any server. This is a variable so it can be
// set at build time using ldflags. The default value is "dev", which is used
// for local development builds.
//
// Summary: Represents a Version.
var Version = "dev"
