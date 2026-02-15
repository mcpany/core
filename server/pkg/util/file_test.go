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
