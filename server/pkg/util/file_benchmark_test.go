// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkReadLastNLines(b *testing.B) {
	// Setup a large file (~10MB)
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "bench.log")

	line := []byte("this is a log line that has some reasonable length to simulate real logs\n")
	// 10MB / 70 bytes ~= 150,000 lines
	iterations := 150000

	var buf bytes.Buffer
	buf.Grow(len(line) * iterations)
	for i := 0; i < iterations; i++ {
		buf.Write(line)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0600); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	b.Run("1000 lines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := ReadLastNLines(path, 1000)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("10000 lines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := ReadLastNLines(path, 10000)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

    b.Run("50000 lines", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := ReadLastNLines(path, 50000)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
