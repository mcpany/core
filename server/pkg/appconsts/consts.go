// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// Summary: Name is the name of the MCP Any server. This is used in help messages and
// other user-facing output.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: Version is the version of the MCP Any server. This is a variable so it can be
// set at build time using ldflags. The default value is "dev", which is used
// for local development builds.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
package appconsts

const (
	Name = "mcpany"
)

var Version = "dev"
