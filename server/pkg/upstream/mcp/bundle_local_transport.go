// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BundleLocalTransport represents the public BundleLocalTransport entity.
//
// Summary: Defines the structured data model representing a local transport.
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
type BundleLocalTransport struct {
	Command    string
	Args       []string
	Env        []string
	WorkingDir string
}

// Connect serves as a public interface for interacting with Connect.
//
// Summary: Connect the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (t *BundleLocalTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	cmd := exec.CommandContext(ctx, t.Command, t.Args...) //nolint:gosec // Trusted configuration
	cmd.Env = t.Env
	cmd.Dir = t.WorkingDir

	stdio := &StdioTransport{Command: cmd}
	return stdio.Connect(ctx)
}
