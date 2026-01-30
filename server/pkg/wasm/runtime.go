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
	//
	// ctx is the context for the request.
	// bytecode is the bytecode.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	LoadPlugin(ctx context.Context, bytecode []byte) (Plugin, error)
	// Close closes the runtime and releases resources.
	//
	// Returns an error if the operation fails.
	Close() error
}

// Plugin defines an instantiated WASM plugin.
type Plugin interface {
	// Execute runs a function exported by the WASM module
	//
	// ctx is the context for the request.
	// function is the function.
	// args is the args.
	//
	// Returns the result.
	// Returns an error if the operation fails.
	Execute(ctx context.Context, function string, args ...[]byte) ([]byte, error)
	// Close closes the plugin instance.
	//
	// Returns an error if the operation fails.
	Close() error
}

// MockRuntime is a placeholder implementation.
type MockRuntime struct{}

// NewMockRuntime creates a new MockRuntime.
//
// Returns the result.
func NewMockRuntime() *MockRuntime {
	return &MockRuntime{}
}

// LoadPlugin loads a plugin.
//
// _ is an unused parameter.
// bytecode is the bytecode.
//
// Returns the result.
// Returns an error if the operation fails.
func (m *MockRuntime) LoadPlugin(_ context.Context, bytecode []byte) (Plugin, error) {
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("btyecode cannot be empty")
	}
	return &MockPlugin{}, nil
}

// Close closes the runtime.
//
// Returns an error if the operation fails.
func (m *MockRuntime) Close() error {
	return nil
}

// MockPlugin is a mock plugin.
type MockPlugin struct{}

// Execute executes a function.
//
// _ is an unused parameter.
// function is the function.
// _ is an unused parameter.
//
// Returns the result.
// Returns an error if the operation fails.
func (p *MockPlugin) Execute(_ context.Context, function string, _ ...[]byte) ([]byte, error) {
	if function == "error" {
		return nil, fmt.Errorf("simulated error")
	}
	return []byte("success"), nil
}

// Close closes the plugin.
//
// Returns an error if the operation fails.
func (p *MockPlugin) Close() error {
	return nil
}
