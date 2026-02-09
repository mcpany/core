// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

func TestReadLastNLines_Advanced(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "advanced.log")

	t.Run("LongLine_GreaterThan64KB", func(t *testing.T) {
		// bufio.Scanner default buffer is 64KB (65536 bytes).
		// We create a line larger than that.
		longLineSize := 70 * 1024
		longLine := make([]byte, longLineSize)
		for i := range longLine {
			longLine[i] = 'a'
		}
		// Add a newline at the end
		longLine[len(longLine)-1] = '\n'

		if err := os.WriteFile(path, longLine, 0600); err != nil {
			t.Fatal(err)
		}

		lines, err := ReadLastNLines(path, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Currently, bufio.Scanner fails silently on long lines if Err() is ignored.
		// It might return a partial line or nothing depending on implementation details.
		// If it returns nothing, len(lines) == 0.
		// If it returns partial line, len(lines) == 1 but length < 70KB.
		if len(lines) == 0 {
			t.Errorf("expected 1 line, got 0")
		} else {
			if len(lines[0]) < longLineSize-1 { // -1 for newline stripped by Scanner
				t.Errorf("expected line length %d, got %d (likely truncated due to buffer limit)", longLineSize-1, len(lines[0]))
			}
		}
	})

	t.Run("ExactChunkSize", func(t *testing.T) {
		// ReadLastNLines uses 16KB chunk size internally.
		// We test file size exactly 16KB with newlines.
		chunkSize := 16 * 1024
		content := bytes.Repeat([]byte("a"), chunkSize-1)
		content = append(content, '\n')

		if err := os.WriteFile(path, content, 0600); err != nil {
			t.Fatal(err)
		}

		lines, err := ReadLastNLines(path, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lines) != 1 {
			t.Errorf("expected 1 line, got %d", len(lines))
		}
		if len(lines[0]) != chunkSize-1 {
			t.Errorf("expected line length %d, got %d", chunkSize-1, len(lines[0]))
		}
	})

	t.Run("ChunkBoundary_SplitLine", func(t *testing.T) {
		// Create a file where a line splits exactly across the 16KB boundary.
		// Chunk size is 16KB.
		// We write 16KB of 'a's, then newline.
		// So total size 16KB + 1 byte.
		// First chunk (backward) reads last 16KB: 15KB 'a's + newline.
		// Second chunk reads first 1KB: 1KB 'a's.
		// Join should work correctly.
		chunkSize := 16 * 1024
		content := bytes.Repeat([]byte("a"), chunkSize)
		content = append(content, '\n')

		if err := os.WriteFile(path, content, 0600); err != nil {
			t.Fatal(err)
		}

		lines, err := ReadLastNLines(path, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lines) != 1 {
			t.Errorf("expected 1 line, got %d", len(lines))
		}
		if len(lines[0]) != chunkSize {
			t.Errorf("expected line length %d, got %d", chunkSize, len(lines[0]))
		}
	})

	t.Run("MultipleChunks_ManyLines", func(t *testing.T) {
		// Write enough lines to span multiple chunks (skip "small file" optimization < 64KB)
		// 16KB chunk size. 64KB small file limit.
		// We need > 64KB file. Let's do 100KB.
		// Each line ~100 bytes. ~1000 lines.
		var buf bytes.Buffer
		lineCount := 1000
		expectedLastLine := "line 999"
		for i := 0; i < lineCount; i++ {
			// Pad to 100 bytes
			s := strings.Repeat("x", 90)
			if i == lineCount-1 {
				buf.WriteString(expectedLastLine + s + "\n")
			} else {
				buf.WriteString("line " + string(rune(i)) + s + "\n")
			}
		}

		if err := os.WriteFile(path, buf.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}

		lines, err := ReadLastNLines(path, 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lines) != 5 {
			t.Errorf("expected 5 lines, got %d", len(lines))
		}
		// Verify content of last line
		lastLine := string(lines[4])
		if !strings.HasPrefix(lastLine, expectedLastLine) {
			t.Errorf("expected last line to start with %q, got %q", expectedLastLine, lastLine)
		}
	})
}
