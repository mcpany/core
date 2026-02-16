// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package wasm provides a WASM plugin runtime.
package wasm

import (
	"context"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
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

// WazeroRuntime implements Runtime using wazero.
type WazeroRuntime struct {
	runtime wazero.Runtime
}

// NewWazeroRuntime creates a new WazeroRuntime.
func NewWazeroRuntime(ctx context.Context) (*WazeroRuntime, error) {
	r := wazero.NewRuntime(ctx)
	// Instantiate WASI to support modules that use it
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}
	return &WazeroRuntime{runtime: r}, nil
}

func (r *WazeroRuntime) LoadPlugin(ctx context.Context, bytecode []byte) (Plugin, error) {
	compiled, err := r.runtime.CompileModule(ctx, bytecode)
	if err != nil {
		return nil, fmt.Errorf("failed to compile module: %w", err)
	}
	return &WazeroPlugin{runtime: r.runtime, compiled: compiled}, nil
}

func (r *WazeroRuntime) Close() error {
	return r.runtime.Close(context.Background())
}

type WazeroPlugin struct {
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
}

func (p *WazeroPlugin) Execute(ctx context.Context, function string, _ ...[]byte) ([]byte, error) {
	// Instantiate the module for this execution (sandboxing)
	modConfig := wazero.NewModuleConfig().WithStdout(nil).WithStderr(nil)
	mod, err := p.runtime.InstantiateModule(ctx, p.compiled, modConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}
	defer mod.Close(ctx)

	fn := mod.ExportedFunction(function)
	if fn == nil {
		return nil, fmt.Errorf("function %s not exported", function)
	}

	// Simple execution: call with no arguments for now (basic support)
	// TODO: Implement complex ABI for passing []byte args and return values.
	results, err := fn.Call(ctx)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	if len(results) > 0 {
		return []byte(fmt.Sprintf("%d", results[0])), nil
	}

	return []byte("success"), nil
}

func (p *WazeroPlugin) Close() error {
	return p.compiled.Close(context.Background())
}

// MockRuntime is a placeholder implementation.
type MockRuntime struct{}

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

// Close closes the plugin.
//
// Returns:
//   - error: Always returns nil.
func (p *MockPlugin) Close() error {
	return nil
}
