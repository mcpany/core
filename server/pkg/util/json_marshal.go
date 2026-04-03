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

// FastMarshalToString serves as a public interface for interacting with FastMarshalToString.
//
// Summary: Fast the marshal to string appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// FastMarshal serves as a public interface for interacting with FastMarshal.
//
// Summary: Fast the marshal appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func FastMarshal(v interface{}) ([]byte, error) {
	return FastJSON.Marshal(v)
}
