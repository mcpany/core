// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"fmt"
	"reflect"
	"strconv"
	"unsafe"
)

// transportError implements the error interface for JSON-RPC errors.
type transportError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error returns the error message.
//
// Returns:
//   - string: The result.
//
// Side Effects:
//   - None.
func (e *transportError) Error() string {
	return e.Message
}

func setUnexportedID(idPtr interface{}, val interface{}) error {
	if val == nil {
		return nil // ID{value: nil} is default
	}
	// jsonrpc2.ID struct has 'value' field.
	// Check if val is number (float64 from json) -> convert to int if possible?
	// jsonrpc2.ID value field is interface{}.

	// Ensure val is int64 if it looks like int (for consistency with SDK which uses int64)
	// JSON unmarshals integer as float64.
	if f, ok := val.(float64); ok {
		if float64(int64(f)) == f {
			val = int64(f)
		}
	}

	v := reflect.ValueOf(idPtr).Elem()
	f := v.FieldByName("value")
	if !f.IsValid() {
		// This suggests the SDK internal structure has changed.
		return fmt.Errorf("field 'value' not found in jsonrpc.ID struct")
	}

	// Safety check: ensure the field is addressable before unsafe operation
	if !f.CanAddr() {
		return fmt.Errorf("field 'value' is not addressable")
	}

	f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	f.Set(reflect.ValueOf(val))
	return nil
}

func fixID(id interface{}) interface{} {
	// ⚡ BOLT: Replaced inefficient regex-based parsing with direct reflection access
	// Randomized Selection from Top 5 High-Impact Targets
	if id == nil {
		return nil
	}

	// Fast path for simple types
	switch v := id.(type) {
	case int, int64, float64, string:
		return v
	case map[string]interface{}:
		if val, ok := v["value"]; ok {
			return fixIDExtracted(val)
		}
	}

	val := reflect.ValueOf(id)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Handle maps via reflection if not covered by type switch (e.g. custom map types)
	if val.Kind() == reflect.Map {
		// Check for "value" key
		key := reflect.ValueOf("value")
		v := val.MapIndex(key)
		if v.IsValid() {
			return fixIDExtracted(v.Interface())
		}
		return id
	}

	if val.Kind() != reflect.Struct {
		return id
	}

	f := val.FieldByName("value")
	if !f.IsValid() {
		return id
	}

	// Make a copy if not addressable to use unsafe on the copy
	if !val.CanAddr() {
		copyVal := reflect.New(val.Type()).Elem()
		copyVal.Set(val)
		val = copyVal
		f = val.FieldByName("value")
	}

	if f.CanAddr() {
		// Use unsafe to read unexported field
		rf := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
		// Recursively fix the extracted value (to handle int conversion)
		return fixIDExtracted(rf.Interface())
	}

	return id
}

func fixIDExtracted(val interface{}) interface{} {
	// If string looks like int, convert. This matches legacy behavior where
	// extracting from a struct would convert string "123" to int 123.
	if s, ok := val.(string); ok {
		if i, err := strconv.Atoi(s); err == nil {
			return i
		}
	}
	// Otherwise recurse
	return fixID(val)
}
