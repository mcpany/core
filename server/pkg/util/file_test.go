// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestReadLastNLines(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.log")

	// Case 1: Empty file
	if err := os.WriteFile(path, []byte(""), 0600); err != nil {
		t.Fatal(err)
	}
	lines, err := ReadLastNLines(path, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("expected 0 lines, got %d", len(lines))
	}

	// Case 2: File with fewer lines than N
	content := "line1\nline2\nline3\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	lines, err = ReadLastNLines(path, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	if string(lines[0]) != "line1" {
		t.Errorf("expected line1, got %s", string(lines[0]))
	}

	// Case 3: File with exactly N lines
	content = "1\n2\n3\n4\n5\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	lines, err = ReadLastNLines(path, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d", len(lines))
	}
	if string(lines[0]) != "1" {
		t.Errorf("expected 1, got %s", string(lines[0]))
	}

	// Case 4: File with more than N lines
	content = "1\n2\n3\n4\n5\n6\n7\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	lines, err = ReadLastNLines(path, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	if string(lines[0]) != "5" {
		t.Errorf("expected 5, got %s", string(lines[0]))
	}
	if string(lines[2]) != "7" {
		t.Errorf("expected 7, got %s", string(lines[2]))
	}

	// Case 5: Large file (larger than chunk size 16KB)
	var large bytes.Buffer
	for i := 0; i < 3000; i++ {
		large.WriteString("this is a relatively long line to fill up space quickly \n")
	}
	if err := os.WriteFile(path, large.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}

	lines, err = ReadLastNLines(path, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 10 {
		t.Errorf("expected 10 lines, got %d", len(lines))
	}
}

func BenchmarkReadLastNLines(b *testing.B) {
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "bench.log")

	// Create a large file (approx 10MB)
	f, err := os.Create(path)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	line := []byte("this is a log line for benchmarking purposes\n")
	// 45 bytes * 250000 = ~11MB
	for i := 0; i < 250000; i++ {
		f.Write(line)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Read last 5000 lines. This should require multiple chunks (chunkSize is 16KB).
		// 5000 * 45 = 225KB. 225KB / 16KB = ~14 chunks.
		// To trigger more chunks, let's read more.
		// Read last 50000 lines -> 2.25MB -> ~140 chunks.
		// 140 iterations of append(small, big...) is noticeable but maybe not huge.
		// Let's read last 200000 lines -> 9MB -> ~560 chunks.
		// 560^2 is 313,600 copies.
		_, err := ReadLastNLines(path, 50000)
		if err != nil {
			b.Fatal(err)
		}
	}
}
