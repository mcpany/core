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
// Summary: Represents a Runtime.
type Runtime interface {
	// LoadPlugin loads a WASM plugin from bytecode.
	//
	// Parameters:
	//   - ctx: The context for the request.
	//   - bytecode: The WASM bytecode to load.
	//
	// Returns:
	//   - Plugin: The instantiated plugin.
	//   - error: An error if the operation fails.
	LoadPlugin(ctx context.Context, bytecode []byte) (Plugin, error)

	// Close closes the runtime and releases resources.
	//
	// Returns:
	//   - error: An error if the operation fails.
	Close() error
}

// Plugin defines an instantiated WASM plugin.
//
// Summary: Represents a Plugin.
type Plugin interface {
	// Execute runs a function exported by the WASM module
	//
	// Parameters:
	//   - ctx: The context for the request.
	//   - function: The name of the function to execute.
	//   - args: The arguments to pass to the function.
	//
	// Returns:
	//   - []byte: The result of the execution.
	//   - error: An error if the operation fails.
	Execute(ctx context.Context, function string, args ...[]byte) ([]byte, error)

	// Close closes the plugin instance.
	//
	// Returns:
	//   - error: An error if the operation fails.
	Close() error
}

// MockRuntime is a placeholder implementation.
//
// Summary: Represents a MockRuntime.
type MockRuntime struct{}

// NewMockRuntime creates a new MockRuntime.
//
// Summary: Initializes NewMockRuntime operation.
//
// Returns:
//   - *MockRuntime: A new mock runtime instance.
func NewMockRuntime() *MockRuntime {
	return &MockRuntime{}
}

// LoadPlugin loads a plugin.
//
// Summary: Executes LoadPlugin operation.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - bytecode: []byte. The bytecode to load.
//
// Returns:
//   - Plugin: A mock plugin.
//   - error: An error if the bytecode is empty.
func (m *MockRuntime) LoadPlugin(_ context.Context, bytecode []byte) (Plugin, error) {
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("btyecode cannot be empty")
	}
	return &MockPlugin{}, nil
}

// Close closes the runtime.
//
// Summary: Executes Close operation.
//
// Returns:
//   - error: An error if closing fails.
func (m *MockRuntime) Close() error {
	return nil
}

// MockPlugin is a mock plugin.
//
// Summary: Represents a MockPlugin.
type MockPlugin struct{}

// Execute executes a function.
//
// Summary: Executes Execute operation.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - function: string. The function name to execute.
//   - args: ...[]byte. The arguments.
//
// Returns:
//   - []byte: The result.
//   - error: An error if the execution fails.
func (p *MockPlugin) Execute(_ context.Context, function string, _ ...[]byte) ([]byte, error) {
	if function == "error" {
		return nil, fmt.Errorf("simulated error")
	}
	return []byte("success"), nil
}

// Close closes the plugin.
//
// Summary: Executes Close operation.
//
// Returns:
//   - error: An error if closing fails.
func (p *MockPlugin) Close() error {
	return nil
}
