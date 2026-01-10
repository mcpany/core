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
	fs           afero.Fs
	rootPaths    map[string]string
	allowedPaths []string
	deniedPaths  []string
}

// NewLocalProvider creates a new LocalProvider from the given configuration.
func NewLocalProvider(_ *configv1.OsFs, rootPaths map[string]string, allowedPaths, deniedPaths []string) *LocalProvider {
	return &LocalProvider{
		fs:           afero.NewOsFs(),
		rootPaths:    rootPaths,
		allowedPaths: allowedPaths,
		deniedPaths:  deniedPaths,
	}
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

	// 4. Allowed/Denied Paths Check
	// Note: We check against the VIRTUAL path provided by the user, as the rules are likely written against that.
	// OR do we check against the resolved REAL path?
	// The issue description implies "configure allowed directories inside [the base directory]".
	// If the user configured root "/" -> "/home/user", and allowed "/home/user/public", they probably mean the real path
	// IF they are thinking about the server configuration.
	// BUT if they are thinking about the MCP interface, they might mean "/public" (virtual).
	// Given MCP Any configures the "Server", and "root_paths" maps Virtual -> Real.
	// Allowed/Denied paths are usually relative to the "exposed" view or the absolute paths on disk.
	// Since `root_paths` can map anything anywhere, checking absolute paths on disk is safest and least ambiguous for the server operator.
	// So we check `targetPathCanonical`.

	if len(p.allowedPaths) > 0 {
		allowed := false
		for _, pattern := range p.allowedPaths {
			// We support glob matching.
			// If pattern is absolute, match against targetPathCanonical.
			// If pattern is relative, it's ambiguous. Let's assume absolute paths are required for security config.
			// Or we can treat them as relative to the matched root? No, that's complex.
			// Let's assume patterns match the REAL path on disk.

			matched, err := filepath.Match(pattern, targetPathCanonical)
			if err != nil {
				// Try prefix match if glob fails or as fallback?
				// Simple prefix check is often what users want for directories.
				if strings.HasPrefix(targetPathCanonical, pattern) {
					matched = true
				}
			} else if !matched {
				// filepath.Match matches the whole string.
				// If the user allows "/tmp/foo", they probably mean "/tmp/foo" AND "/tmp/foo/bar".
				// filepath.Match("/tmp/foo", "/tmp/foo/bar") is false.
				// So we should check if pattern is a prefix.
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
			return "", fmt.Errorf("access denied: path not in allowed list")
		}
	}

	if len(p.deniedPaths) > 0 {
		for _, pattern := range p.deniedPaths {
			matched, err := filepath.Match(pattern, targetPathCanonical)
			if err != nil {
				if strings.HasPrefix(targetPathCanonical, pattern) {
					matched = true
				}
			} else if !matched {
				if strings.HasPrefix(targetPathCanonical, pattern) {
					matched = true
				}
			}

			if matched {
				return "", fmt.Errorf("access denied: path is in denied list")
			}
		}
	}

	return targetPathCanonical, nil
}

// Close closes the provider.
func (p *LocalProvider) Close() error {
	return nil
}
