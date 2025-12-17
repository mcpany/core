// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package appconsts

const (
	// Name is the name of the MCP Any server. This is used in help messages and
	// other user-facing output.
	Name = "mcpany"
)

// Version is the version of the MCP Any server. This is a variable so it can be
// set at build time using ldflags. The default value is "dev", which is used
// for local development builds.
var Version = "dev"
