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
type MockRuntime struct{}

// NewMockRuntime initializes and returns a new mock runtime instance.
//
// Parameters:
//   - None
//
// Returns:
//   - *MockRuntime: The generated or retrieved entity.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// NewMockRuntime creates a new MockRuntime.
//
// Returns:
//   - *MockRuntime: A new mock runtime instance.
func NewMockRuntime() *MockRuntime {
	return &MockRuntime{}
}

// LoadPlugin loads a plugin.
//
// Parameters:
//   - _ : The context (unused).
//   - bytecode: The bytecode to load.
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

// Close handles close.
//
// Parameters:
//   - None
//
// Returns:
//   - error: Returns an error if the execution fails or validation does not pass.
//
// Errors:
//   - Returns an error if the input is malformed, dependencies are unreachable, or state validation fails.
//
// Side Effects:
//   - None.
// Close closes the runtime.
//
// Returns:
//   - error: Always returns nil.
func (m *MockRuntime) Close() error {
	return nil
}

// MockPlugin is a mock plugin.
type MockPlugin struct{}

// Execute executes a function.
//
// Parameters:
//   - _ : The context (unused).
//   - function: The function name to execute.
//   - _ : The arguments (unused).
//
// Returns:
//   - []byte: The result ("success").
//   - error: An error if the function name is "error".
func (p *MockPlugin) Execute(_ context.Context, function string, _ ...[]byte) ([]byte, error) {
	if function == "error" {
		return nil, fmt.Errorf("simulated error")
	}
	return []byte("success"), nil
}

// Close handles close.
//
// Parameters:
//   - None
//
// Returns:
//   - error: Returns an error if the execution fails or validation does not pass.
//
// Errors:
//   - Returns an error if the input is malformed, dependencies are unreachable, or state validation fails.
//
// Side Effects:
//   - None.
// Close closes the plugin.
//
// Returns:
//   - error: Always returns nil.
func (p *MockPlugin) Close() error {
	return nil
}
