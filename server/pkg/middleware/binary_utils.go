// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"encoding/binary"
	"math"
)

// float32ToBytes converts a slice of float32 to a byte slice.
func float32ToBytes(floats []float32) []byte {
	bytes := make([]byte, len(floats)*4)
	for i, f := range floats {
		binary.LittleEndian.PutUint32(bytes[i*4:], math.Float32bits(f))
	}
	return bytes
}

// bytesToFloat32 converts a byte slice to a slice of float32.
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
