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
		// Powers of 10 boundaries
		9999999999, // 10 digits -> 2
		10000000000, // 11 digits -> 2
		-9999999999, // 11 chars -> 2
		-10000000000, // 12 chars -> 3
		// More large numbers (assuming 64-bit int)
		100000000000, // 12 digits -> 3
		-100000000000, // 13 chars -> 3
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
		// Powers of 10 boundaries (positive)
		9999999999, // 10 digits -> 2
		10000000000, // 11 digits -> 2 (up to 11/4=2)
		99999999999, // 11 digits -> 2
		100000000000, // 12 digits -> 3
		999999999999999, // 15 digits -> 3
		1000000000000000, // 16 digits -> 4
		// Powers of 10 boundaries (negative)
		-9999999999, // 11 chars -> 2
		-10000000000, // 12 chars -> 3
		-99999999999, // 12 chars -> 3
		-100000000000, // 13 chars -> 3
		-99999999999999, // 15 chars -> 3
		-100000000000000, // 16 chars -> 4
		-999999999999999999, // 19 chars -> 4
		-1000000000000000000, // 20 chars -> 5
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
