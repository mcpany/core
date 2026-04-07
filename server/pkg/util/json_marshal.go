// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	jsoniter "github.com/json-iterator/go"
	"sync"
)

var (
	// FastJSON is a configured instance of jsoniter that is optimized for performance.
	//
	// Summary: Provides a high-performance JSON marshaller.
	FastJSON = jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            false,
		ValidateJsonRawMessage: true,
	}.Froze()

	bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

// Summary: FastMarshalToString executes the operation.
//
// Parameters:
//   - v interface{}: Input parameter.
//
// Returns:
//   - (string, error): Result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None
func FastMarshalToString(v interface{}) (string, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	stream := FastJSON.BorrowStream(buf)
	defer FastJSON.ReturnStream(stream)

	stream.WriteVal(v)
	if stream.Error != nil {
		return "", stream.Error
	}

	stream.Flush()
	return buf.String(), nil
}

// FastMarshal performs a high-performance JSON marshal into a byte slice.
//
// Summary: Marshals to a byte slice efficiently.
//
// Parameters:
//   - v (interface{}): The value to marshal.
//
// Returns:
//   - []byte: The marshaled byte slice.
// Summary: FastMarshal executes the operation.
//
// Parameters:
//   - v interface{}: Input parameter.
//
// Returns:
//   - ([]byte, error): Result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None
func FastMarshal(v interface{}) ([]byte, error) {
	return FastJSON.Marshal(v)
}
