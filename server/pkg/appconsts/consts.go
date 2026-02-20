// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package appconsts

const (
	// Name is the name of the MCP Any server.
	//
	// Summary: Identifies the application name used in user-facing output.
	Name = "mcpany"
)

// Version is the version of the MCP Any server.
//
// Summary: Holds the current version of the application.
//
// Side Effects:
//   - Can be modified at build time via ldflags.
var Version = "dev"
