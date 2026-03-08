// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/mcpany/core/server/pkg/validation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BundleLocalTransport implements mcp.Transport for running a bundle locally via exec.
type BundleLocalTransport struct {
	Command    string
	Args       []string
	Env        []string
	WorkingDir string
}

// Connect establishes a connection to the local process.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//
// Returns:
//   - mcp.Connection: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the command fails to start.
//
// Side Effects:
//   - Starts a local process.
func (t *BundleLocalTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	if t.WorkingDir != "" {
		if err := validation.IsAllowedPath(t.WorkingDir); err != nil {
			return nil, fmt.Errorf("invalid working directory %q: %w", t.WorkingDir, err)
		}
	}

	cmd := exec.CommandContext(ctx, t.Command, t.Args...) //nolint:gosec // Trusted configuration
	cmd.Env = t.Env
	cmd.Dir = t.WorkingDir

	stdio := &StdioTransport{Command: cmd}
	return stdio.Connect(ctx)
}
