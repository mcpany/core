// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"testing"
)

// We define a struct that matches SDK's jsonrpc.ID (which uses interface{})
type brokenIDInterface struct {
	value interface{}
}

func BenchmarkFixID_Old_Baseline(b *testing.B) {
	// We use the same struct type as the new test to be fair,
	// assuming fixIDOld works on it (it prints with %+v so it should work).
	id := brokenIDInterface{value: 123456}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fixIDOld(id)
	}
}

func BenchmarkFixID_New_Optimization(b *testing.B) {
	id := brokenIDInterface{value: 123456}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fixID(id)
	}
}

func TestFixIDReflectionCorrectness_Integration(t *testing.T) {
	id := brokenIDInterface{value: 123456}
	res := fixID(id)
	if res != 123456 {
		t.Errorf("Expected 123456, got %v", res)
	}

	idStr := brokenIDInterface{value: "foo"}
	resStr := fixID(idStr)
	if resStr != "foo" {
		t.Errorf("Expected 'foo', got %v", resStr)
	}
}
