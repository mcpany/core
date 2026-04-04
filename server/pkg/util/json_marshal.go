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
// Summary: Is a configured instance of jsoniter that is optimized for performance.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

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

// FastMarshalToString performs a high-performance JSON marshal into a string.
//
// Summary: Performs a high-performance JSON marshal into a string.
//
// Parameters:
//   - v (interface{}): Parameter.
//
// Returns:
//   - string: Return value.
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
//
// Side Effects:
//   - None.

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
// Summary: Performs a high-performance JSON marshal into a byte slice.
//
// Parameters:
//   - v (interface{}): Parameter.
//
// Returns:
//   - []byte: Return value.
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
//
// Side Effects:
//   - None.

func FastMarshal(v interface{}) ([]byte, error) {
	return FastJSON.Marshal(v)
}
