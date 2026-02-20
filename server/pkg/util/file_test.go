// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
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

func TestReadLastNLines_BufferCompaction(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "compaction.log")

	// Create a file larger than 4KB (bufio default buffer) but smaller than 64KB
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	// Write distinct lines. 1000 lines * ~7 bytes = ~7000 bytes.
	// Enough to trigger buffer compaction (default buf size is 4KB).
	for i := 0; i < 1000; i++ {
		if _, err := f.WriteString("line" + strconv.Itoa(i) + "\n"); err != nil {
			t.Fatal(err)
		}
	}
	f.Close()

	lines, err := ReadLastNLines(path, 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(lines) != 1000 {
		t.Errorf("expected 1000 lines, got %d", len(lines))
	}

	// Verify first line (which is most likely to be overwritten/corrupted)
	if string(lines[0]) != "line0" {
		t.Errorf("lines[0]: expected 'line0', got '%s'", string(lines[0]))
	}
	// Verify a line from the middle
	if string(lines[500]) != "line500" {
		t.Errorf("lines[500]: expected 'line500', got '%s'", string(lines[500]))
	}
	// Verify last line
	if string(lines[999]) != "line999" {
		t.Errorf("lines[999]: expected 'line999', got '%s'", string(lines[999]))
	}
}
