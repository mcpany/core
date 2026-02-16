// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package wasm

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWazeroRuntime(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewWazeroRuntime(ctx)
	require.NoError(t, err)
	defer runtime.Close()

	// Simple WASM module that exports a function "hello" returning 42
	// (module
	//   (func $hello (result i32)
	//     i32.const 42)
	//   (export "hello" (func $hello))
	// )
	wasmHex := "0061736d010000000105016000017f030201000709010568656c6c6f00000a06010400412a0b"
	wasmBytes, err := hex.DecodeString(wasmHex)
	require.NoError(t, err)

	plugin, err := runtime.LoadPlugin(ctx, wasmBytes)
	require.NoError(t, err)
	defer plugin.Close()

	// Execute "hello"
	result, err := plugin.Execute(ctx, "hello")
	require.NoError(t, err)
	assert.Equal(t, "42", string(result))
}

func TestWazeroRuntime_InvalidBytecode(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewWazeroRuntime(ctx)
	require.NoError(t, err)
	defer runtime.Close()

	_, err = runtime.LoadPlugin(ctx, []byte("invalid wasm"))
	assert.Error(t, err)
}

func TestWazeroRuntime_MissingFunction(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewWazeroRuntime(ctx)
	require.NoError(t, err)
	defer runtime.Close()

	// Use same valid WASM
	wasmHex := "0061736d010000000105016000017f030201000709010568656c6c6f00000a06010400412a0b"
	wasmBytes, err := hex.DecodeString(wasmHex)
	require.NoError(t, err)

	plugin, err := runtime.LoadPlugin(ctx, wasmBytes)
	require.NoError(t, err)
	defer plugin.Close()

	_, err = plugin.Execute(ctx, "missing_function")
	assert.Error(t, err)
}
