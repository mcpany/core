// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package wasm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockWASMRuntime(t *testing.T) {
	runtime := NewMockRuntime()

	// Test Load
	plugin, err := runtime.LoadPlugin(context.Background(), []byte("fake-wasm"))
	assert.NoError(t, err)
	assert.NotNil(t, plugin)

	// Test Execute
	out, err := plugin.Execute(context.Background(), "run")
	assert.NoError(t, err)
	assert.Equal(t, []byte("success"), out)

	// Test Error
	_, err = plugin.Execute(context.Background(), "error")
	assert.Error(t, err)

	// Test Empty Bytecode
	_, err = runtime.LoadPlugin(context.Background(), nil)
	assert.Error(t, err)
}

func TestWazeroRuntime(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewRuntime(ctx)
	assert.NoError(t, err)
	defer runtime.Close()

	// Minimal valid WASM module header
	minimalWasm := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

	// Test Load
	plugin, err := runtime.LoadPlugin(ctx, minimalWasm)
	assert.NoError(t, err)
	assert.NotNil(t, plugin)
	defer plugin.Close()

	// Test Execute (non-existent function)
	_, err = plugin.Execute(ctx, "non_existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not exported")
}
