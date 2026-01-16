// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package wasm provides a WASM plugin runtime.
package wasm

import (
	"context"
	"fmt"
	"math"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Runtime defines the interface for a WASM plugin runtime.
type Runtime interface {
	// LoadPlugin loads a WASM plugin from bytecode.
	LoadPlugin(ctx context.Context, bytecode []byte) (Plugin, error)
	// Close closes the runtime and releases resources.
	Close() error
}

// Plugin defines an instantiated WASM plugin.
type Plugin interface {
	// Execute runs a function exported by the WASM module
	Execute(ctx context.Context, function string, args ...[]byte) ([]byte, error)
	// Close closes the plugin instance.
	Close() error
}

// WazeroRuntime implements Runtime using wazero.
type WazeroRuntime struct {
	runtime wazero.Runtime
}

// NewRuntime creates a new WASM runtime.
func NewRuntime(ctx context.Context) (Runtime, error) {
	r := wazero.NewRuntime(ctx)
	// Instantiate WASI
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		_ = r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}
	return &WazeroRuntime{runtime: r}, nil
}

// LoadPlugin loads a plugin.
func (r *WazeroRuntime) LoadPlugin(ctx context.Context, bytecode []byte) (Plugin, error) {
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("bytecode cannot be empty")
	}

	// Compile and Instantiate
	// Note: In a real system we might separate compilation and instantiation.
	mod, err := r.runtime.Instantiate(ctx, bytecode)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}

	return &WazeroPlugin{mod: mod}, nil
}

// Close closes the runtime.
func (r *WazeroRuntime) Close() error {
	return r.runtime.Close(context.Background())
}

// WazeroPlugin implements Plugin.
type WazeroPlugin struct {
	mod api.Module
}

// Execute executes a function.
// It assumes a simple ABI where arguments are passed as pointers (after allocation)
// and result is a pointer/size.
// For now, to keep it safe, if we don't know the ABI, we just look for the function.
// If the function takes no args, we call it.
// To fully support args, we need 'malloc' and 'free'.
func (p *WazeroPlugin) Execute(ctx context.Context, function string, args ...[]byte) ([]byte, error) {
	f := p.mod.ExportedFunction(function)
	if f == nil {
		return nil, fmt.Errorf("function %s not exported", function)
	}

	// Basic ABI check: if we have args, we need malloc
	var malloc api.Function
	if len(args) > 0 {
		malloc = p.mod.ExportedFunction("malloc")
		if malloc == nil {
			return nil, fmt.Errorf("module must export 'malloc' to accept arguments")
		}
	}

	// Prepare args
	wasmArgs := make([]uint64, 0, len(args)*2)
	for _, arg := range args {
		size := uint64(len(arg))
		results, err := malloc.Call(ctx, size)
		if err != nil {
			return nil, fmt.Errorf("malloc failed: %w", err)
		}
		ptr := results[0]
		if ptr > math.MaxUint32 {
			return nil, fmt.Errorf("malloc returned out of bounds pointer")
		}

		// Write arg to memory
		if !p.mod.Memory().Write(uint32(ptr), arg) {
			return nil, fmt.Errorf("memory write failed")
		}

		wasmArgs = append(wasmArgs, ptr, size)
	}

	// Call function
	results, err := f.Call(ctx, wasmArgs...)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	// Handle result
	// Assumption: Result is (ptr, len) encoded as uint64 or two return values?
	// wazero supports multiple returns.
	// Let's assume the function returns a pointer to a buffer (and maybe we assume null-terminated or we need length).
	// Or returns (ptr, len).
	// If results len is 1, maybe it's just ptr (and we need length from somewhere else or it's a primitive).
	// If results len is 0, empty []byte.

	if len(results) == 0 {
		return nil, nil
	}

	// For this implementation, let's assume it returns a pointer to a uint64 length-prefixed buffer?
	// OR (ptr, len) as two values.

	if len(results) == 1 {
	    // If it returns a single value, assume it's a pointer to a string/bytes that is somehow terminated or self-describing?
	    // Or maybe it's just a simple value we convert to string.
	    val := results[0]
	    return []byte(fmt.Sprintf("%d", val)), nil
	}

	if len(results) == 2 {
		if results[0] > math.MaxUint32 || results[1] > math.MaxUint32 {
			return nil, fmt.Errorf("result pointers out of bounds")
		}
	    ptr := uint32(results[0]) //nolint:gosec // Checked above
	    size := uint32(results[1]) //nolint:gosec // Checked above
	    bytes, ok := p.mod.Memory().Read(ptr, size)
	    if !ok {
	        return nil, fmt.Errorf("memory read failed")
	    }
	    return bytes, nil
	}

	return nil, fmt.Errorf("unexpected return values count: %d", len(results))
}

// Close closes the plugin.
func (p *WazeroPlugin) Close() error {
	return p.mod.Close(context.Background())
}

// MockRuntime is a placeholder implementation.
type MockRuntime struct{}

// NewMockRuntime creates a new MockRuntime.
func NewMockRuntime() *MockRuntime {
	return &MockRuntime{}
}

// LoadPlugin loads a plugin.
func (m *MockRuntime) LoadPlugin(_ context.Context, bytecode []byte) (Plugin, error) {
	if len(bytecode) == 0 {
		return nil, fmt.Errorf("bytecode cannot be empty")
	}
	return &MockPlugin{}, nil
}

// Close closes the runtime.
func (m *MockRuntime) Close() error {
	return nil
}

// MockPlugin is a mock plugin.
type MockPlugin struct{}

// Execute executes a function.
func (p *MockPlugin) Execute(_ context.Context, function string, _ ...[]byte) ([]byte, error) {
	if function == "error" {
		return nil, fmt.Errorf("simulated error")
	}
	return []byte("success"), nil
}

// Close closes the plugin.
func (p *MockPlugin) Close() error {
	return nil
}
