// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBundleLocalTransport_InvalidWorkingDir(t *testing.T) {
	transport := &BundleLocalTransport{
		Command:    "echo",
		Args:       []string{"hello"},
		WorkingDir: "../../../../../../../../../../../../../../../../../../../../etc",
	}

	_, err := transport.Connect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid working directory")
}
