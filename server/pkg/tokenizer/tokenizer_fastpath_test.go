// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"fmt"
	"math"
	"testing"
)

func TestCountTokensInValue_FastPathConsistency(t *testing.T) {
	tok := NewSimpleTokenizer()

	tests := []struct {
		name  string
		input interface{}
	}{
		{"Zero", 0},
		{"Small positive", 123},
		{"Small negative", -123},
		{"MaxInt", math.MaxInt},
		{"MinInt", math.MinInt},
		{"MaxInt64", int64(math.MaxInt64)},
		{"MinInt64", int64(math.MinInt64)},
		{"Bool true", true},
		{"Bool false", false},
		{"Nil", nil},
		{"Nested Map", map[string]interface{}{"a": 1, "b": true}},
		{"Nested Slice", []interface{}{1, true, nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fast path result
			got, err := CountTokensInValue(tok, tt.input)
			if err != nil {
				t.Fatalf("CountTokensInValue failed: %v", err)
			}

			// Slow path equivalent (manually calculate using string conversion for primitives, or manual sum for complex)
			// For complex types, we can't easily replicate "slow path" without bypassing the fast path in the code.
			// But we can check primitives directly.
			// For complex types, we can trust that if primitives are correct, the structural recursion is correct
			// (as verified by existing TestCountTokensInValue).

			// Let's rely on string representation for primitives:
			switch v := tt.input.(type) {
			case int, int64, bool:
				str := fmt.Sprintf("%v", v)
				want, _ := tok.CountTokens(str)
				if got != want {
					t.Errorf("Mismatch for %v (str: %q): fast=%d, slow=%d", tt.input, str, got, want)
				}
			case nil:
				// fmt.Sprintf("%v", nil) is "<nil>" (5 chars). 5/4 = 1.
				// Fast path returns 1.
				if got != 1 {
					t.Errorf("Mismatch for nil: fast=%d, want 1", got)
				}
			}
		})
	}
}
