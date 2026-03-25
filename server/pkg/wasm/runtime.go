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

// NewMockRuntime creates a new mock runtime.
//
// Summary: Creates a new mock runtime.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - *MockRuntime: The result.
func NewMockRuntime() *MockRuntime {
	return &MockRuntime{}
}

// LoadPlugin loadPlugin load plugin.
//
// Summary: LoadPlugin load plugin.
//
// Parameters: - None.
//   - _ (context.Context): Unused parameter.
//   - bytecode ([]byte): The bytecode.
//
// Returns: - None.
//   - Plugin: The result.
//   - error: An error if the operation fails.
func (m *MockRuntime) LoadPlugin(_ context.Context, bytecode []byte) (Plugin, error) {
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("btyecode cannot be empty")
	}
	return &MockPlugin{}, nil
}

// Close close close.
//
// Summary: Close close.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (m *MockRuntime) Close() error {
	return nil
}

// MockPlugin is a mock plugin.
//
// Summary: Represents a MockPlugin.
type MockPlugin struct{}

// Execute executes the operation.
//
// Summary: Executes the operation.
//
// Parameters: - None.
//   - _ (context.Context): Unused parameter.
//   - function (string): The function.
//   - _ (...[]byte): Unused parameter.
//
// Returns: - None.
//   - []byte: The result.
//   - error: An error if the operation fails.
func (p *MockPlugin) Execute(_ context.Context, function string, _ ...[]byte) ([]byte, error) {
	if function == "error" {
		return nil, fmt.Errorf("simulated error")
	}
	return []byte("success"), nil
}

// Close close close.
//
// Summary: Close close.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - error: An error if the operation fails.
func (p *MockPlugin) Close() error {
	return nil
}
