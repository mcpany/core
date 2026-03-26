// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BundleLocalTransport implements mcp.Transport for running a bundle locally via exec.
//
// Summary: MCP transport for running bundle services as local processes.
type BundleLocalTransport struct {
	Command    string
	Args       []string
	Env        []string
	WorkingDir string
}

// Connect establishes a connection to the local process.
//
// Summary: Executes Connect operation to start and connect to a local process.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//
// Returns:
//   - mcp.Connection: The established connection.
//   - error: An error if connection fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (t *BundleLocalTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	cmd := exec.CommandContext(ctx, t.Command, t.Args...) //nolint:gosec // Trusted configuration
	cmd.Env = t.Env
	cmd.Dir = t.WorkingDir

	stdio := &StdioTransport{Command: cmd}
	return stdio.Connect(ctx)
}
