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
// Summary: Runtime defines the interface for a WASM plugin runtime.
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
// Summary: Plugin defines an instantiated WASM plugin.
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
// Summary: MockRuntime is a placeholder implementation.
type MockRuntime struct{}

// NewMockRuntime creates a new MockRuntime.
//
// Summary: NewMockRuntime creates a new MockRuntime.
//
// Parameters:
//   - None.
//
// Returns:
//   - *MockRuntime: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func NewMockRuntime() *MockRuntime {
	return &MockRuntime{}
}

// LoadPlugin loads a plugin.
//
// Summary: LoadPlugin loads a plugin.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - bytecode ([]byte): The provided bytecode data.
//
// Returns:
//   - Plugin: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (m *MockRuntime) LoadPlugin(_ context.Context, bytecode []byte) (Plugin, error) {
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("btyecode cannot be empty")
	}
	return &MockPlugin{}, nil
}

// Close closes the runtime.
//
// Summary: Close closes the runtime.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (m *MockRuntime) Close() error {
	return nil
}

// MockPlugin is a mock plugin.
//
// Summary: MockPlugin is a mock plugin.
type MockPlugin struct{}

// Execute executes a function.
//
// Summary: Execute executes a function.
//
// Parameters:
//   - _ (context.Context): The provided _ data.
//   - function (string): The textual representation of function.
//   - _ (...[]byte): The provided _ data.
//
// Returns:
//   - []byte: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (p *MockPlugin) Execute(_ context.Context, function string, _ ...[]byte) ([]byte, error) {
	if function == "error" {
		return nil, fmt.Errorf("simulated error")
	}
	return []byte("success"), nil
}

// Close closes the plugin.
//
// Summary: Close closes the plugin.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (p *MockPlugin) Close() error {
	return nil
}
