// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Initial load should return nil
	cfg, err := store.Load(ctx)
	assert.NoError(t, err)
	assert.Nil(t, cfg)

	// Update configuration
	newCfg := &configv1.McpAnyServerConfig{
		GlobalSettings: &configv1.GlobalSettings{
			McpListenAddress: proto.String(":8080"),
		},
	}
	store.Update(newCfg)

	// Load should return the new configuration
	loadedCfg, err := store.Load(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, loadedCfg)
	assert.Equal(t, ":8080", loadedCfg.GetGlobalSettings().GetMcpListenAddress())

	// Verify it's a clone/copy
	newCfg.GlobalSettings.McpListenAddress = proto.String(":9090")
	loadedCfg2, err := store.Load(ctx)
	assert.NoError(t, err)
	assert.Equal(t, ":8080", loadedCfg2.GetGlobalSettings().GetMcpListenAddress())
}
