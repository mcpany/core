// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package wasm provides a WASM plugin runtime.
package wasm

import (
	"context"
	"fmt"
)

// Runtime defines the interface for a WASM plugin runtime.
//
// Summary: defines the interface for a WASM plugin runtime.
type Runtime interface {
	// LoadPlugin loads a WASM plugin from bytecode.
	//
	// Summary: loads a WASM plugin from bytecode.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - bytecode: []byte. The []byte.
	//
	// Returns:
	//   - Plugin: The Plugin.
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	LoadPlugin(ctx context.Context, bytecode []byte) (Plugin, error)

	// Close closes the runtime and releases resources.
	//
	// Summary: closes the runtime and releases resources.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	Close() error
}

// Plugin defines an instantiated WASM plugin.
//
// Summary: defines an instantiated WASM plugin.
type Plugin interface {
	// Execute runs a function exported by the WASM module.
	//
	// Summary: runs a function exported by the WASM module.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - function: string. The string.
	//   - args: ...[]byte. The []byte.
	//
	// Returns:
	//   - []byte: The []byte.
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	Execute(ctx context.Context, function string, args ...[]byte) ([]byte, error)

	// Close closes the plugin instance.
	//
	// Summary: closes the plugin instance.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   - error: An error if the operation fails.
	//
	// Throws/Errors:
	//   Returns an error if the operation fails.
	Close() error
}

// MockRuntime is a placeholder implementation.
//
// Summary: is a placeholder implementation.
type MockRuntime struct{}

// NewMockRuntime creates a new MockRuntime.
//
// Summary: creates a new MockRuntime.
//
// Parameters:
//   None.
//
// Returns:
//   - *MockRuntime: The *MockRuntime.
func NewMockRuntime() *MockRuntime {
	return &MockRuntime{}
}

// LoadPlugin loads a plugin.
//
// Summary: loads a plugin.
//
// Parameters:
//   - _: context.Context. The _.
//   - bytecode: []byte. The bytecode.
//
// Returns:
//   - Plugin: The Plugin.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (m *MockRuntime) LoadPlugin(_ context.Context, bytecode []byte) (Plugin, error) {
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("btyecode cannot be empty")
	}
	return &MockPlugin{}, nil
}

// Close closes the runtime.
//
// Summary: closes the runtime.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (m *MockRuntime) Close() error {
	return nil
}

// MockPlugin is a mock plugin.
//
// Summary: is a mock plugin.
type MockPlugin struct{}

// Execute executes a function.
//
// Summary: executes a function.
//
// Parameters:
//   - _: context.Context. The _.
//   - function: string. The function.
//   - _: ...[]byte. The _.
//
// Returns:
//   - []byte: The []byte.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (p *MockPlugin) Execute(_ context.Context, function string, _ ...[]byte) ([]byte, error) {
	if function == "error" {
		return nil, fmt.Errorf("simulated error")
	}
	return []byte("success"), nil
}

// Close closes the plugin.
//
// Summary: closes the plugin.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (p *MockPlugin) Close() error {
	return nil
}
