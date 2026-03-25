// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
// FastMarshalToString performs a high-performance JSON marshal into a string.
//
// Summary: Marshals to a string efficiently.
//
// Parameters:
// Returns:
// Errors:
//   - None.
// Side Effects:
//   - None.
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//   - v (interface{}): The value to marshal.
//
// Returns:
//   - string: The marshaled string.
//   - error: An error if marshaling fails.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func FastMarshalToString(v interface{}) (string, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	stream := FastJSON.BorrowStream(buf)
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
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func FastMarshal(v interface{}) ([]byte, error) {
	return FastJSON.Marshal(v)
}
