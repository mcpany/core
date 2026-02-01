/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// Package util provides utility functions.
package util

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir checks if a directory exists and creates it if not.
func EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		//nolint:gosec
		return os.MkdirAll(dir, 0750)
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
	//nolint:gosec
	return os.WriteFile(path, content, 0600)
}

// ReadLastNLines reads the last n lines from a file.
func ReadLastNLines(path string, n int) ([]string, error) {
	//nolint:gosec
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		//nolint:errcheck
		_ = file.Close()
	}()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
