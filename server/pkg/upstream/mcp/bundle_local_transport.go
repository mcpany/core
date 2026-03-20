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
// Summary: BundleLocalTransport implements mcp.Transport for running a bundle locally via exec.
type BundleLocalTransport struct {
	Command    string
	Args       []string
	Env        []string
	WorkingDir string
}

// Connect establishes a connection to the local process.
//
// Summary: Connect establishes a connection to the local process.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//
// Returns:
//   - mcp.Connection: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (t *BundleLocalTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	cmd := exec.CommandContext(ctx, t.Command, t.Args...) //nolint:gosec // Trusted configuration
	cmd.Env = t.Env
	cmd.Dir = t.WorkingDir

	stdio := &StdioTransport{Command: cmd}
	return stdio.Connect(ctx)
}
