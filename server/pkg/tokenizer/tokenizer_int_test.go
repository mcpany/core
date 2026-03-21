// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"strconv"
	"testing"
)

func TestSimpleTokenizeInt_EdgeCases(t *testing.T) {
	tests := []int{
		0, 1, 9, 10, 99, 100, 999, 1000, 9999,
		9999999,  // 7 digits -> 1
		10000000, // 8 digits -> 2
		-1, -9, -10, -99, -100,
		-999999,  // 7 chars -> 1
		-1000000, // 8 chars -> 2
		-9999999, // 8 chars -> 2
        12345678,
        -12345678,
	}

	for _, n := range tests {
		s := strconv.Itoa(n)
		want := len(s) / 4
		if want < 1 {
			want = 1
		}
		got := simpleTokenizeInt(n)
		if got != want {
			t.Errorf("n=%d (%s) len=%d: got %d, want %d", n, s, len(s), got, want)
		}
	}
}

func TestSimpleTokenizeInt64_EdgeCases(t *testing.T) {
    tests := []int64{
		0, 1, 9, 10, 99, 100, 999, 1000, 9999,
		9999999,  // 7 digits -> 1
		10000000, // 8 digits -> 2
		-1, -9, -10, -99, -100,
		-999999,  // 7 chars -> 1
		-1000000, // 8 chars -> 2
		-9999999, // 8 chars -> 2
        12345678,
        -12345678,
	}

	for _, n := range tests {
		s := strconv.FormatInt(n, 10)
		want := len(s) / 4
		if want < 1 {
			want = 1
		}
		got := simpleTokenizeInt64(n)
		if got != want {
			t.Errorf("n=%d (%s) len=%d: got %d, want %d", n, s, len(s), got, want)
		}
	}
}
