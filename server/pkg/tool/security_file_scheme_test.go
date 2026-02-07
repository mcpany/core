// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"strings"
	"testing"
)

func TestFileSchemeInDocker(t *testing.T) {
	// Attempt to use file scheme in Docker environment
	// Currently this passes (returns nil), but we want it to fail
	// because file:// can access container secrets/files or cause SSRF.
	err := validateSafePathAndInjection("file:///etc/passwd", true)
	if err == nil {
		t.Fatal("Expected error for file scheme in Docker, got nil")
	}
	if !strings.Contains(err.Error(), "file: scheme detected") {
		t.Fatalf("Expected 'file: scheme detected' error, got %v", err)
	}
}

func TestFileSchemeLocal(t *testing.T) {
	err := validateSafePathAndInjection("file:///etc/passwd", false)
	if err == nil {
		t.Fatal("Expected error for file scheme locally, got nil")
	}
	if !strings.Contains(err.Error(), "file: scheme detected") {
		t.Fatalf("Expected 'file: scheme detected' error, got %v", err)
	}
}

func TestAbsPathLocal(t *testing.T) {
	err := validateSafePathAndInjection("/etc/passwd", false)
	if err == nil {
		t.Fatal("Expected error for abs path locally, got nil")
	}
	if !strings.Contains(err.Error(), "absolute path detected") {
		t.Fatalf("Expected 'absolute path detected' error, got %v", err)
	}
}

func TestAbsPathDocker(t *testing.T) {
	// Abs path should be ALLOWED in Docker (relative to container root)
	// This ensures we don't break valid Docker usage where users specify absolute paths in the container.
	err := validateSafePathAndInjection("/etc/passwd", true)
	if err != nil {
		t.Fatalf("Expected nil for abs path in Docker, got %v", err)
	}
}
