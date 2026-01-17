// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IsSecurePath checks if a given file path is secure and does not contain any
// path traversal sequences ("../" or "..\\"). This function is crucial for
// preventing directory traversal attacks, where a malicious actor could
// otherwise access or manipulate files outside of the intended directory.
//
// path is the file path to be validated.
// It returns an error if the path is found to be insecure, and nil otherwise.
// IsSecurePath checks if a given file path is secure.
// It is a variable to allow mocking in tests.
var IsSecurePath = func(path string) error {
	clean := filepath.Clean(path)
	parts := strings.Split(clean, string(os.PathSeparator))
	for _, part := range parts {
		if part == ".." {
			return fmt.Errorf("path contains '..', which is not allowed")
		}
	}
	return nil
}

var (
	allowedPaths []string
)

// SetAllowedPaths checks if a given file path is relative and does not contain any
// path traversal sequences ("../").
func SetAllowedPaths(paths []string) {
	allowedPaths = paths
}

// IsAllowedPath checks if a given file path is allowed (inside CWD or AllowedPaths)
// and does not contain any path traversal sequences ("../").
// It is a variable to allow mocking in tests.
var IsAllowedPath = func(path string) error {
	// 1. Basic security check (no .. in the path string itself)
	if err := IsSecurePath(path); err != nil {
		return err
	}

	// 2. Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Resolve Symlinks to ensure we are checking the real path
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If file does not exist, we need to check the first existing parent directory.
			// We want to ensure that even if the file is created later, it won't traverse out.
			// Since we already checked `IsSecurePath` (no `..`), if the non-existing parts don't have `..`,
			// the only risk is if an existing parent component is a symlink pointing out.
			current := absPath
			for {
				// If current exists, we are good to stop and resolve it.
				if _, err := os.Stat(current); err == nil {
					break
				}
				parent := filepath.Dir(current)
				if parent == current {
					// We reached root and it doesn't exist? Should not happen on unix.
					break
				}
				current = parent
			}

			realCurrent, err := filepath.EvalSymlinks(current)
			if err != nil {
				return fmt.Errorf("failed to resolve path %q: %w", path, err)
			}

			// Now calculate the suffix that didn't exist
			rel, err := filepath.Rel(current, absPath)
			if err != nil {
				return fmt.Errorf("failed to calculate relative path: %w", err)
			}

			realPath = filepath.Join(realCurrent, rel)
		} else {
			return fmt.Errorf("failed to resolve symlinks for %q: %w", path, err)
		}
	}

	// Helper to check if child is inside parent
	isInside := func(parent, child string) bool {
		// We use EvalSymlinks on parent too just in case CWD is symlinked
		realParent, err := filepath.EvalSymlinks(parent)
		if err == nil {
			parent = realParent
		}

		rel, err := filepath.Rel(parent, child)
		if err != nil {
			return false
		}
		return !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != ".."
	}

	// 3. Check if inside CWD
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	if isInside(cwd, realPath) {
		return nil
	}

	// 4. Check Allowed Paths
	for _, allowedDir := range allowedPaths {
		allowedDir = strings.TrimSpace(allowedDir)
		if allowedDir == "" {
			continue
		}

		// Resolve allowedDir to absolute too
		allowedAbs, err := filepath.Abs(allowedDir)
		if err == nil && isInside(allowedAbs, realPath) {
			return nil
		}
	}

	return fmt.Errorf("path %q is not allowed (must be in CWD or in allowed paths)", path)
}
