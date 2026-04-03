// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"encoding/binary"
	"math"
)

// float32ToBytes serves as a public interface for interacting with float32ToBytes.
//
// Summary: Float the to bytes appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func float32ToBytes(floats []float32) []byte {
	bytes := make([]byte, len(floats)*4)
	for i, f := range floats {
		binary.LittleEndian.PutUint32(bytes[i*4:], math.Float32bits(f))
	}
	return bytes
}

// bytesToFloat32 serves as a public interface for interacting with bytesToFloat32.
//
// Summary: Bytes the to float appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func bytesToFloat32(bytes []byte) []float32 {
	if len(bytes)%4 != 0 {
		return nil // Invalid length
	}
	floats := make([]float32, len(bytes)/4)
	for i := range floats {
		floats[i] = math.Float32frombits(binary.LittleEndian.Uint32(bytes[i*4:]))
	}
	return floats
}
