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
// Summary. Represents a Runtime.
type Runtime interface {
	// LoadPlugin loads a WASM plugin from bytecode.
	//
// Parameters.
	//   - ctx: The context for the request.
	//   - bytecode: The WASM bytecode to load.
	//
// Returns.
	//   - Plugin: The instantiated plugin.
	//   - error: An error if the operation fails.
	LoadPlugin(ctx context.Context, bytecode []byte) (Plugin, error)

	// Close closes the runtime and releases resources.
	//
// Returns.
	//   - error: An error if the operation fails.
	Close() error
}

// Plugin defines an instantiated WASM plugin.
//
// Summary. Represents a Plugin.
type Plugin interface {
	// Execute runs a function exported by the WASM module
	//
// Parameters.
	//   - ctx: The context for the request.
	//   - function: The name of the function to execute.
	//   - args: The arguments to pass to the function.
	//
// Returns.
	//   - []byte: The result of the execution.
	//   - error: An error if the operation fails.
	Execute(ctx context.Context, function string, args ...[]byte) ([]byte, error)

	// Close closes the plugin instance.
	//
// Returns.
	//   - error: An error if the operation fails.
	Close() error
}

// MockRuntime is a placeholder implementation.
//
// Summary. Represents a MockRuntime.
type MockRuntime struct{}

// NewMockRuntime provides newmockruntime functionality.
//
// Summary: NewMockRuntime.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func NewMockRuntime() *MockRuntime {
	return &MockRuntime{}
}

// LoadPlugin provides loadplugin functionality.
//
// Summary: LoadPlugin.
//
// Parameters.
//   - _: The parameter.
//   - bytecode: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (m *MockRuntime) LoadPlugin(_ context.Context, bytecode []byte) (Plugin, error) {
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("btyecode cannot be empty")
	}
	return &MockPlugin{}, nil
}

// Close provides close functionality.
//
// Summary: Close.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (m *MockRuntime) Close() error {
	return nil
}

// MockPlugin is a mock plugin.
//
// Summary. Represents a MockPlugin.
type MockPlugin struct{}

// Execute provides execute functionality.
//
// Summary: Execute.
//
// Parameters.
//   - _: The parameter.
//   - function: The parameter.
//   - _: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (p *MockPlugin) Execute(_ context.Context, function string, _ ...[]byte) ([]byte, error) {
	if function == "error" {
		return nil, fmt.Errorf("simulated error")
	}
	return []byte("success"), nil
}

// Close provides close functionality.
//
// Summary: Close.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (p *MockPlugin) Close() error {
	return nil
}
