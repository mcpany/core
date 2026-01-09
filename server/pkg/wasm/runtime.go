// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package wasm provides a WASM plugin runtime.
package wasm

import (
	"context"
	"fmt"
)

// Runtime defines the interface for a WASM plugin runtime.
type Runtime interface {
	// LoadPlugin loads a WASM plugin from bytecode.
	LoadPlugin(ctx context.Context, bytecode []byte) (Plugin, error)
	// Close closes the runtime and releases resources.
	Close() error
}

// Plugin defines an instantiated WASM plugin.
type Plugin interface {
	// Execute runs a function exported by the WASM module
	Execute(ctx context.Context, function string, args ...[]byte) ([]byte, error)
	// Close closes the plugin instance.
	Close() error
}

// MockRuntime is a placeholder implementation.
type MockRuntime struct{}

// NewMockRuntime creates a new MockRuntime.
func NewMockRuntime() *MockRuntime {
	return &MockRuntime{}
}

// LoadPlugin loads a plugin.
func (m *MockRuntime) LoadPlugin(_ context.Context, bytecode []byte) (Plugin, error) {
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("btyecode cannot be empty")
	}
	return &MockPlugin{}, nil
}

// Close closes the runtime.
func (m *MockRuntime) Close() error {
	return nil
}

// MockPlugin is a mock plugin.
type MockPlugin struct{}

// Execute executes a function.
func (p *MockPlugin) Execute(_ context.Context, function string, _ ...[]byte) ([]byte, error) {
	if function == "error" {
		return nil, fmt.Errorf("simulated error")
	}
	return []byte("success"), nil
}

// Close closes the plugin.
func (p *MockPlugin) Close() error {
	return nil
}
