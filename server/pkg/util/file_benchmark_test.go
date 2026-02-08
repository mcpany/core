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
	// Create a large file
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "benchmark.log")

	line := []byte("this is a relatively long line of log data to simulate realistic usage scenarios 1234567890\n")
	repeats := 100000 // ~8MB file

	f, err := os.Create(path)
	if err != nil {
		b.Fatal(err)
	}

	// Buffered write for speed
	var buf bytes.Buffer
	for i := 0; i < 1000; i++ {
		buf.Write(line)
	}
	chunk := buf.Bytes()

	for i := 0; i < repeats/1000; i++ {
		if _, err := f.Write(chunk); err != nil {
			b.Fatal(err)
		}
	}
	f.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Read last 5000 lines
		_, err := ReadLastNLines(path, 5000)
		if err != nil {
			b.Fatal(err)
		}
	}
}
