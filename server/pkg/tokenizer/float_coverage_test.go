// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"testing"
)

// TestFloatOptimizationCoverage specifically targets the optimization branches
// for integer-like floats in countTokensInValueSimpleFast to ensure coverage.
func TestFloatOptimizationCoverage(t *testing.T) {
	st := NewSimpleTokenizer()

	// 1. Single float64 optimization
	// Value must be:
	// - > -1e6
	// - < 1e6
	// - integral (val == float64(int64(val)))
	// Result should be 1 token.

	// Case 1: Hit the optimization
	hitVal := 12345.0
	c, _, err := countTokensInValueSimpleFast(st, hitVal)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if c != 1 {
		t.Errorf("Expected 1 token for optimization hit, got %d", c)
	}

	// Case 2: Miss optimization (not integral)
	_ = 123.45
	// "123.45" -> 6 chars -> 6/4 = 1. Wait, 123.45 is 'g' fmt.
	// We want to make sure we didn't take the int path.
	// If we took int path, it would be count(123) = 1.
	// If we take float path, "123.45" -> 6 chars -> 1 token.
	// This doesn't distinguish.
	// Let's use a value where token counts differ?
	// 12345.6789 -> 10 chars -> 2 tokens.
	// If treated as int (12345) -> 1 token.
	missValDiff := 12345.6789
	cMiss, _, err := countTokensInValueSimpleFast(st, missValDiff)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if cMiss != 2 {
		t.Errorf("Expected 2 tokens for non-integral float, got %d", cMiss)
	}

	// Case 3: Miss optimization (out of range)
	// 1000000.0 (1e6) -> "1e+06" -> 5 chars -> 1 token.
	// If we treated as int -> 1000000 -> 7 digits -> 1 token.
	// Hard to distinguish by count alone for simple values.
	// But we need to ensure the *code path* is executed for coverage.
	missValRange := 2000000.0
	cRange, _, err := countTokensInValueSimpleFast(st, missValRange)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	// "2e+06" -> 5 chars -> 1 token.
	// If int: 2000000 -> 7 chars -> 1 token.
	if cRange != 1 {
		t.Errorf("Expected 1 token, got %d", cRange)
	}


	// 2. Slice []float64 optimization
	// We need a slice where some items hit and some miss, or all hit.

	// All hit
	sliceHit := []float64{1.0, 2.0, 3.0}
	cSlice, _, err := countTokensInValueSimpleFast(st, sliceHit)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if cSlice != 3 { // 1 + 1 + 1
		t.Errorf("Expected 3 tokens for slice hit, got %d", cSlice)
	}

	// Mixed (hit and miss)
	sliceMixed := []float64{
		1.0, // Hit -> 1
		12345.6789, // Miss -> 2
	}
	cMixed, _, err := countTokensInValueSimpleFast(st, sliceMixed)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if cMixed != 3 { // 1 + 2
		t.Errorf("Expected 3 tokens for mixed slice, got %d", cMixed)
	}
}
