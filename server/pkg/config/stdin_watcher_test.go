// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"sync"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestStdinWatcher(t *testing.T) {
	input := `{"global_settings": {"mcp_listen_address": ":8081"}}
{"global_settings": {"mcp_listen_address": ":8082"}}
`
	reader := bytes.NewBufferString(input)
	watcher := NewStdinWatcher(reader)

	var updates []*configv1.McpAnyServerConfig
	var mu sync.Mutex
	done := make(chan bool)

	go watcher.Watch(func(cfg *configv1.McpAnyServerConfig) {
		mu.Lock()
		defer mu.Unlock()
		updates = append(updates, cfg)
		if len(updates) == 2 {
			done <- true
		}
	})

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for updates")
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, updates, 2)
	assert.Equal(t, ":8081", updates[0].GetGlobalSettings().GetMcpListenAddress())
	assert.Equal(t, ":8082", updates[1].GetGlobalSettings().GetMcpListenAddress())
}

func TestStdinWatcher_InvalidJSON(t *testing.T) {
	// Should skip invalid JSON lines
	input := `{"global_settings": {"mcp_listen_address": ":8081"}}
INVALID_JSON
{"global_settings": {"mcp_listen_address": ":8082"}}
`
	reader := bytes.NewBufferString(input)
	watcher := NewStdinWatcher(reader)

	var updates []*configv1.McpAnyServerConfig
	var mu sync.Mutex
	done := make(chan bool)

	go watcher.Watch(func(cfg *configv1.McpAnyServerConfig) {
		mu.Lock()
		defer mu.Unlock()
		updates = append(updates, cfg)
		if len(updates) == 2 {
			done <- true
		}
	})

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for updates")
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, updates, 2)
	assert.Equal(t, ":8081", updates[0].GetGlobalSettings().GetMcpListenAddress())
	assert.Equal(t, ":8082", updates[1].GetGlobalSettings().GetMcpListenAddress())
}
