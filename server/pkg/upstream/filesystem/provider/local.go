// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
)

// LocalProvider provides access to the local filesystem.
type LocalProvider struct {
	fs        afero.Fs
	rootPaths map[string]string
}

// NewLocalProvider creates a new LocalProvider from the given configuration.
func NewLocalProvider(_ *configv1.OsFs, rootPaths map[string]string) (*LocalProvider, error) {
	// Validate root paths
	for _, realPath := range rootPaths {
		info, err := os.Stat(realPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("root path does not exist: %s", realPath)
			}
			return nil, fmt.Errorf("failed to access root path %s: %w", realPath, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("root path is not a directory: %s", realPath)
		}
	}

	return &LocalProvider{
		fs:        afero.NewOsFs(),
		rootPaths: rootPaths,
	}, nil
}

// GetFs returns the underlying filesystem.
func (p *LocalProvider) GetFs() afero.Fs {
	return p.fs
}

// ResolvePath resolves the virtual path to a real path in the local filesystem.
func (p *LocalProvider) ResolvePath(virtualPath string) (string, error) {
	if len(p.rootPaths) == 0 {
		return "", fmt.Errorf("no root paths defined")
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
			// cleanVRoot is "/data", checkPath is "/database" -> startsWith is true, but not a child.
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
			return "", fmt.Errorf("path %s is not allowed (no matching root)", virtualPath)
		}
	}

	// 2. Resolve the path
	relativePath := strings.TrimPrefix(virtualPath, bestMatchVirtual)
	// handle case where virtualPath matched exactly or with trailing slash
	relativePath = strings.TrimPrefix(relativePath, "/")

	realRootAbs, err := filepath.Abs(bestMatchReal)
	if err != nil {
		return "", fmt.Errorf("failed to resolve root path: %w", err)
	}

	// Resolve symlinks in the root path to ensure we have the canonical path
	realRootCanonical, err := filepath.EvalSymlinks(realRootAbs)
	if err != nil {
		// If root doesn't exist, we can't really secure it, so error out.
		return "", fmt.Errorf("failed to resolve root path symlinks: %w", err)
	}

	targetPath := filepath.Join(realRootCanonical, relativePath)
	targetPathAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target path: %w", err)
	}

	// Resolve symlinks for the target path too
	targetPathCanonical, err := filepath.EvalSymlinks(targetPathAbs)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, we need to find the deepest existing ancestor
			// to ensure that no symlinks in the path point outside the root.
			currentPath := targetPathAbs
			var existingPath string
			var remainingPath string

			for {
				dir := filepath.Dir(currentPath)
				if dir == currentPath {
					// Reached root without finding existing path (should not happen if realRoot exists)
					return "", fmt.Errorf("failed to resolve path (root not found): %s", targetPathAbs)
				}

				// Check if dir exists
				if _, err := os.Stat(dir); err == nil {
					existingPath = dir
					var relErr error
					remainingPath, relErr = filepath.Rel(existingPath, targetPathAbs)
					if relErr != nil {
						return "", fmt.Errorf("failed to calculate relative path: %w", relErr)
					}
					break
				}
				currentPath = dir
			}

			// Resolve symlinks for the existing ancestor
			existingPathCanonical, err := filepath.EvalSymlinks(existingPath)
			if err != nil {
				return "", fmt.Errorf("failed to resolve ancestor path symlinks: %w", err)
			}

			// Construct the canonical path
			targetPathCanonical = filepath.Join(existingPathCanonical, remainingPath)

			// Note: We don't check if the "remainingPath" contains ".." because filepath.Rel and Join should handle it,
			// and we are constructing it from absolute paths.
			// However, since the intermediate directories don't exist, they can't be symlinks pointing elsewhere.
		} else {
			return "", fmt.Errorf("failed to resolve target path symlinks: %w", err)
		}
	}

	// 3. Security Check: Ensure targetPathCanonical starts with realRootCanonical
	if !strings.HasPrefix(targetPathCanonical, realRootCanonical+string(os.PathSeparator)) && targetPathCanonical != realRootCanonical {
		return "", fmt.Errorf("access denied: path traversal detected")
	}

	return targetPathCanonical, nil
}

// Close closes the provider.
func (p *LocalProvider) Close() error {
	return nil
}
