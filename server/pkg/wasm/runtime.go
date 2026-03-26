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
// Summary: Interface providing methods to load and manage WASM-based plugins.
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
// Summary: Interface for interacting with a loaded WASM module.
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
// Summary: A mock implementation of the Runtime interface for testing purposes.
type MockRuntime struct{}

// NewMockRuntime creates a new MockRuntime.
//
// Summary: Returns a pointer to a new MockRuntime instance.
//
// Returns:
//   - *MockRuntime: The initialized mock runtime.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewMockRuntime() *MockRuntime {
	return &MockRuntime{}
}

// LoadPlugin loads a plugin.
//
// Summary: Simulates loading a plugin from WASM bytecode for testing.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - bytecode: []byte. The WASM bytecode to "load".
//
// Returns:
//   - Plugin: A new MockPlugin instance.
//   - error: Returns an error if the provided bytecode slice is empty.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (m *MockRuntime) LoadPlugin(_ context.Context, bytecode []byte) (Plugin, error) {
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("btyecode cannot be empty")
	}
	return &MockPlugin{}, nil
}

// Close closes the runtime.
//
// Summary: Implements the Runtime.Close method as a no-op for the mock.
//
// Returns:
//   - error: Always returns nil.
//
// Parameters:
//   - None.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (m *MockRuntime) Close() error {
	return nil
}

// MockPlugin is a mock plugin.
//
// Summary: A mock implementation of the Plugin interface for testing.
type MockPlugin struct{}

// Execute executes a function.
//
// Summary: Simulates the execution of a WASM function with support for simulated errors.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - function: string. The name of the function to "execute".
//   - args: ...[]byte. The byte-slice arguments to pass.
//
// Returns:
//   - []byte: Returns "success" if the function is not "error".
//   - error: Returns a simulated error if the function name is "error".
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (p *MockPlugin) Execute(_ context.Context, function string, _ ...[]byte) ([]byte, error) {
	if function == "error" {
		return nil, fmt.Errorf("simulated error")
	}
	return []byte("success"), nil
}

// Close closes the plugin.
//
// Summary: Implements the Plugin.Close method as a no-op for the mock.
//
// Returns:
//   - error: Always returns nil.
//
// Parameters:
//   - None.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (p *MockPlugin) Close() error {
	return nil
}
