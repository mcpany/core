// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// resolveRoot finds the best matching root path for the given virtual path.
func (p *LocalProvider) resolveRoot(virtualPath string) (string, string, error) {
	if len(p.rootPaths) == 0 {
		return "", "", fmt.Errorf("no root paths defined")
	}

	// 1. Find the best matching root path (longest prefix match)
	var bestMatchVirtual string
	var bestMatchReal string

	for vRoot, rRoot := range p.rootPaths {
		// Ensure vRoot has a clean format
		cleanVRoot := vRoot
		if !strings.HasPrefix(cleanVRoot, "/") {
			cleanVRoot = "/" + cleanVRoot
		}

		// Ensure virtualPath starts with /
		checkPath := virtualPath
		if !strings.HasPrefix(checkPath, "/") {
			checkPath = "/" + checkPath
		}

		if strings.HasPrefix(checkPath, cleanVRoot) {
			// Ensure it matches as a directory component
			// Matches if checkPath == cleanVRoot OR checkPath starts with cleanVRoot + "/"
			validMatch := checkPath == cleanVRoot ||
				strings.HasPrefix(checkPath, cleanVRoot+"/") ||
				cleanVRoot == "/"

			if validMatch {
				if len(cleanVRoot) > len(bestMatchVirtual) {
					bestMatchVirtual = cleanVRoot
					bestMatchReal = rRoot
				}
			}
		}
	}

	if bestMatchVirtual == "" {
		// Try fallback: if rootPaths has "/" key, use it.
		if val, ok := p.rootPaths["/"]; ok {
			bestMatchVirtual = "/"
			bestMatchReal = val
		} else {
			return "", "", fmt.Errorf("path %s is not allowed (no matching root)", virtualPath)
		}
	}
	return bestMatchVirtual, bestMatchReal, nil
}

// resolveSymlinks resolves symlinks in the path, handling non-existent paths by resolving the deepest existing ancestor.
func resolveSymlinks(path string) (string, error) {
	canonical, err := filepath.EvalSymlinks(path)
	if err == nil {
		return canonical, nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// If file doesn't exist, find deepest existing ancestor
	currentPath := path
	var existingPath string
	var remainingPath string

	for {
		dir := filepath.Dir(currentPath)
		if dir == currentPath {
			// Reached root without finding existing path
			return "", fmt.Errorf("failed to resolve path (root not found): %s", path)
		}

		if _, err := os.Stat(dir); err == nil {
			existingPath = dir
			var relErr error
			remainingPath, relErr = filepath.Rel(existingPath, path)
			if relErr != nil {
				return "", fmt.Errorf("failed to calculate relative path: %w", relErr)
			}
			break
		}
		currentPath = dir
	}

	existingPathCanonical, err := filepath.EvalSymlinks(existingPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve ancestor path symlinks: %w", err)
	}

	return filepath.Join(existingPathCanonical, remainingPath), nil
}

// checkAccess verifies if the path is allowed based on allowed/denied patterns.
func (p *LocalProvider) checkAccess(targetPathCanonical string) error {
	if len(p.allowedPaths) > 0 {
		allowed := false
		for _, pattern := range p.allowedPaths {
			matched, err := filepath.Match(pattern, targetPathCanonical)
			if err != nil || !matched {
				if strings.HasPrefix(targetPathCanonical, pattern) {
					matched = true
				}
			}

			if matched {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("access denied: path not in allowed list")
		}
	}

	if len(p.deniedPaths) > 0 {
		for _, pattern := range p.deniedPaths {
			matched, err := filepath.Match(pattern, targetPathCanonical)
			if err != nil || !matched {
				if strings.HasPrefix(targetPathCanonical, pattern) {
					matched = true
				}
			}

			if matched {
				return fmt.Errorf("access denied: path is in denied list")
			}
		}
	}
	return nil
}
