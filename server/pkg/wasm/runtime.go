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
	// Parameters: - None.
	//   - ctx: The context for the request.
	//   - bytecode: The WASM bytecode to load.
	//
	// Returns: - None.
	//   - Plugin: The instantiated plugin.
	//   - error: An error if the operation fails.
	LoadPlugin(ctx context.Context, bytecode []byte) (Plugin, error)

	// Close closes the runtime and releases resources.
	//
	// Returns: - None.
	//   - error: An error if the operation fails.
	Close() error
}

// Plugin defines an instantiated WASM plugin.
//
// Summary: Represents a Plugin.
type Plugin interface {
	// Execute runs a function exported by the WASM module
	//
	// Parameters: - None.
	//   - ctx: The context for the request.
	//   - function: The name of the function to execute.
	//   - args: The arguments to pass to the function.
	//
	// Returns: - None.
	//   - []byte: The result of the execution.
	//   - error: An error if the operation fails.
	Execute(ctx context.Context, function string, args ...[]byte) ([]byte, error)

	// Close closes the plugin instance.
	//
	// Returns: - None.
	//   - error: An error if the operation fails.
	Close() error
}

// MockRuntime is a placeholder implementation.
//
// Summary: Represents a MockRuntime.
type MockRuntime struct{}

// NewMockRuntime creates a new MockRuntime.
//
// Returns: - None.
//   - *MockRuntime: A new mock runtime instance.
//
// Summary: Initializes NewMockRuntime operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func NewMockRuntime() *MockRuntime {
	return &MockRuntime{}
}

// LoadPlugin loads a plugin.
//
// Parameters: - None.
//   - _ : The context (unused).
//   - bytecode: The bytecode to load.
//
// Returns: - None.
//   - Plugin: A mock plugin.
//   - error: An error if the bytecode is empty.
//
// Summary: Executes LoadPlugin operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (m *MockRuntime) LoadPlugin(_ context.Context, bytecode []byte) (Plugin, error) {
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("btyecode cannot be empty")
	}
	return &MockPlugin{}, nil
}

// Close closes the runtime.
//
// Returns: - None.
//   - error: Always returns nil.
//
// Summary: Executes Close operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (m *MockRuntime) Close() error {
	return nil
}

// MockPlugin is a mock plugin.
//
// Summary: Represents a MockPlugin.
type MockPlugin struct{}

// Execute executes a function.
//
// Parameters: - None.
//   - _ : The context (unused).
//   - function: The function name to execute.
//   - _ : The arguments (unused).
//
// Returns: - None.
//   - []byte: The result ("success").
//   - error: An error if the function name is "error".
//
// Summary: Executes Execute operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (p *MockPlugin) Execute(_ context.Context, function string, _ ...[]byte) ([]byte, error) {
	if function == "error" {
		return nil, fmt.Errorf("simulated error")
	}
	return []byte("success"), nil
}

// Close closes the plugin.
//
// Returns: - None.
//   - error: Always returns nil.
//
// Summary: Executes Close operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (p *MockPlugin) Close() error {
	return nil
}
