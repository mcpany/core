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
//
// Methods:
//   - Defines the required contract and behavior for implementations of Runtime.
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
//
// Methods:
//   - Defines the required contract and behavior for implementations of Plugin.
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
//
// Fields:
//   - Contains the configuration and state properties required for MockRuntime functionality.
type MockRuntime struct{}

// NewMockRuntime creates a new MockRuntime. Returns: - *MockRuntime: A new mock runtime instance.
//
// Summary: NewMockRuntime creates a new MockRuntime. Returns: - *MockRuntime: A new mock runtime instance.
//
// Parameters:
//   - None.
//
// Returns:
//   - (*MockRuntime): The resulting MockRuntime object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewMockRuntime() *MockRuntime {
	return &MockRuntime{}
}

// LoadPlugin loads a plugin. Parameters: - _ : The context (unused). - bytecode: The bytecode to load. Returns: - Plugin: A mock plugin. - error: An error if the bytecode is empty.
//
// Summary: LoadPlugin loads a plugin. Parameters: - _ : The context (unused). - bytecode: The bytecode to load. Returns: - Plugin: A mock plugin. - error: An error if the bytecode is empty.
//
// Parameters:
//   - _ (context.Context): The _ parameter used in the operation.
//   - bytecode ([]byte): The bytecode parameter used in the operation.
//
// Returns:
//   - (Plugin): The resulting Plugin object containing the requested data.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - Modifies global state, writes to the database, or establishes network connections.
func (m *MockRuntime) LoadPlugin(_ context.Context, bytecode []byte) (Plugin, error) {
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("btyecode cannot be empty")
	}
	return &MockPlugin{}, nil
}

// Close closes the runtime. Returns: - error: Always returns nil.
//
// Summary: Close closes the runtime. Returns: - error: Always returns nil.
//
// Parameters:
//   - None.
//
// Returns:
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
func (m *MockRuntime) Close() error {
	return nil
}

// MockPlugin is a mock plugin.
//
// Summary: MockPlugin is a mock plugin.
//
// Fields:
//   - Contains the configuration and state properties required for MockPlugin functionality.
type MockPlugin struct{}

// Execute executes a function. Parameters: - _ : The context (unused). - function: The function name to execute. - _ : The arguments (unused). Returns: - []byte: The result ("success"). - error: An error if the function name is "error".
//
// Summary: Execute executes a function. Parameters: - _ : The context (unused). - function: The function name to execute. - _ : The arguments (unused). Returns: - []byte: The result ("success"). - error: An error if the function name is "error".
//
// Parameters:
//   - _ (context.Context): The _ parameter used in the operation.
//   - function (string): The function parameter used in the operation.
//   - _ (...[]byte): The _ parameter used in the operation.
//
// Returns:
//   - ([]byte): The resulting []byte object containing the requested data.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
func (p *MockPlugin) Execute(_ context.Context, function string, _ ...[]byte) ([]byte, error) {
	if function == "error" {
		return nil, fmt.Errorf("simulated error")
	}
	return []byte("success"), nil
}

// Close closes the plugin. Returns: - error: Always returns nil.
//
// Summary: Close closes the plugin. Returns: - error: Always returns nil.
//
// Parameters:
//   - None.
//
// Returns:
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
func (p *MockPlugin) Close() error {
	return nil
}
