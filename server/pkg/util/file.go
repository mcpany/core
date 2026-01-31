/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// Package util provides utility functions.
// nolint:revive
package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir checks if a directory exists and creates it if not.
func EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// FileExists checks if a file exists.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// WriteFile writes content to a file, creating the directory if needed.
func WriteFile(path string, content []byte) error {
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return os.WriteFile(path, content, 0644)
}
