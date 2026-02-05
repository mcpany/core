// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"testing"
	"time"
)

func TestCountTokensInValue_ByteSlice_Performance(t *testing.T) {
	// Create a 1MB byte slice
	size := 1024 * 1024
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		data[i] = byte(i % 256)
	}

	st := NewSimpleTokenizer()

	start := time.Now()
	_, err := CountTokensInValue(st, data)
	if err != nil {
		t.Fatalf("CountTokensInValue failed: %v", err)
	}
	duration := time.Since(start)

	t.Logf("Time taken for 1MB []byte with SimpleTokenizer: %v", duration)

	// Fail if it takes too long (e.g., > 100ms).
	// Without optimization, it iterates 1M times with reflection and Sprintf, likely taking seconds.
	if duration > 100*time.Millisecond {
		t.Errorf("Performance regression: 1MB []byte took %v, expected < 100ms", duration)
	}
}
