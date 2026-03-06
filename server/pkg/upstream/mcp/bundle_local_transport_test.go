// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBundleLocalTransport_Connect(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid command triggers error", func(t *testing.T) {
		transport := &BundleLocalTransport{
			Command: "non_existent_command_to_trigger_error",
			Args:    []string{},
			Env:     []string{},
		}

		conn, err := transport.Connect(ctx)

		assert.Error(t, err)
		assert.Nil(t, conn)
		assert.Contains(t, err.Error(), "executable file not found")
	})

	t.Run("successful connection to simple command", func(t *testing.T) {
		transport := &BundleLocalTransport{
			Command: "echo",
			Args:    []string{"hello"},
			Env:     []string{},
		}

		conn, err := transport.Connect(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, conn)
		if conn != nil {
			_ = conn.Close()
		}
	})
}
