// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	jsoniter "github.com/json-iterator/go"
	"sync"
)

// FastJSON is a configured instance of jsoniter that is optimized for performance.
//
// Summary: Provides a high-performance JSON marshaller.
var (
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
// Summary: Marshals to a string efficiently.
//
// Parameters:
//   - v (interface{}): The value to marshal.
//
// Returns:
//   - string: The marshaled string.
//   - error: An error if marshaling fails.
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
//   - error: An error if marshaling fails.
func FastMarshal(v interface{}) ([]byte, error) {

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	stream := FastJSON.BorrowStream(buf)
	defer FastJSON.ReturnStream(stream)

	stream.WriteVal(v)
	if stream.Error != nil {
		return nil, stream.Error
	}

	stream.Flush()

	// Create a copy of the bytes because the buffer will be reused
	res := make([]byte, buf.Len())
	copy(res, buf.Bytes())
	return res, nil
}
