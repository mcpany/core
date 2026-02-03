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
	// Setup a large file once
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "bench.log")

	// Create a 10MB file with ~200 byte lines -> ~50,000 lines
	line := []byte("this is a relatively long line that represents a log entry in the system with some metadata and timestamps and other things to make it reasonably realistic in size for a benchmark test case\n")
	fileSize := 10 * 1024 * 1024 // 10MB
	iterations := fileSize / len(line)

	f, err := os.Create(path)
	if err != nil {
		b.Fatal(err)
	}

	// Use buffered writer for fast setup
	var buf bytes.Buffer
	for i := 0; i < iterations; i++ {
		buf.Write(line)
	}
	if _, err := f.Write(buf.Bytes()); err != nil {
		b.Fatal(err)
	}
	f.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Read last 10,000 lines (approx 2MB)
		// This forces multiple chunks to be read and triggered the O(N^2) behavior
		_, err := ReadLastNLines(path, 10000)
		if err != nil {
			b.Fatal(err)
		}
	}
}
