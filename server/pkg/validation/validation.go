// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package validation provides validation utilities for config files and other inputs.
package validation

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// IsValidBindAddress checks if a given string is a valid bind address.
// A valid bind address is in the format "host:port".
//
// Summary: Validates a bind address string.
//
// Parameters:
//   - s: string. The address string to validate.
//
// Returns:
//   - error: An error if validation fails.
func IsValidBindAddress(s string) error {
	_, port, err := net.SplitHostPort(s)
	if err != nil {
		// If the error is due to missing port in address (which happens when no colon is present),
		// we check if it is a valid port-only string by prepending ":".
		if !strings.Contains(s, ":") {
			// Ensure it is a numeric port to distinguish from missing-port hostnames like "localhost"
			if p, err := strconv.Atoi(s); err == nil {
				if p < 0 || p > 65535 {
					return fmt.Errorf("port must be between 0 and 65535")
				}
				_, port, err = net.SplitHostPort(":" + s)
				if err == nil && port != "" {
					return nil
				}
			}
		}
		return err
	}
	if port == "" {
		return fmt.Errorf("port is required")
	}
	// Check if port is numeric and within range
	p, err := strconv.Atoi(port)
	if err != nil {
		// Non-numeric ports (service names) are allowed.
		return nil //nolint:nilerr // Intentional: we allow non-numeric ports
	}
	if p < 0 || p > 65535 {
		return fmt.Errorf("port must be between 0 and 65535")
	}
	return nil
}

// IsSecurePath checks if a given file path is secure and does not contain any
// path traversal sequences ("../" or "..\\"). This function is crucial for
// preventing directory traversal attacks, where a malicious actor could
// otherwise access or manipulate files outside of the intended directory.
//
// Summary: Checks for path traversal attempts.
//
// Parameters:
//   - path: string. The path to check.
//
// Returns:
//   - error: An error if the path is insecure.
//
// IsSecurePath is a variable to allow mocking in tests.
var IsSecurePath = func(path string) error {
	// ⚡ BOLT: Fast path to avoid expensive string splitting for safe paths.
	// Randomized Selection from Top 5 High-Impact Targets
	if strings.Contains(path, "..") {
		// Check original path for traversal attempts before cleaning
		parts := strings.Split(path, string(os.PathSeparator))
		for _, part := range parts {
			if part == ".." {
				return fmt.Errorf("path contains '..', which is not allowed")
			}
		}

		clean := filepath.Clean(path)
		parts = strings.Split(clean, string(os.PathSeparator))
		for _, part := range parts {
			if part == ".." {
				return fmt.Errorf("path contains '..', which is not allowed")
			}
		}
	}
	return nil
}

// IsSecureRelativePath checks if a given file path is secure, relative, and does not contain any
// path traversal sequences. It strictly disallows absolute paths and drive letters.
//
// Summary: Checks if a path is secure and relative.
var IsSecureRelativePath = func(path string) error {
	// 1. Basic security check (no ..)
	if err := IsSecurePath(path); err != nil {
		return err
	}

	// 2. Check for absolute path
	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute paths are not allowed: %s", path)
	}

	// 3. Check for drive letter (Windows)
	if filepath.VolumeName(path) != "" {
		return fmt.Errorf("paths with volume names are not allowed: %s", path)
	}

	// 4. Check for leading separator (to prevent being treated as absolute or root-relative on some systems)
	// On Windows, \foo is not Abs, but is rooted. We want to disallow it.
	if strings.HasPrefix(path, string(os.PathSeparator)) {
		return fmt.Errorf("paths starting with separator are not allowed: %s", path)
	}

	// 5. Check for alternate separator if different (e.g. / on Windows, \ on Linux)
	// We want to be strict.
	if os.PathSeparator == '\\' && strings.HasPrefix(path, "/") {
		return fmt.Errorf("paths starting with '/' are not allowed")
	}
	// Note: We don't check for '\\' on Linux as it is a valid filename character,
	// although unusual. Blocking it might break valid filenames.

	return nil
}

var (
	allowedPaths []string
)

// SetAllowedPaths sets the list of allowed paths for file operations.
//
// Summary: Sets the global allowed paths list.
//
// Parameters:
//   - paths: []string. The list of allowed paths.
func SetAllowedPaths(paths []string) {
	allowedPaths = paths
}

// IsSensitivePath checks if a given file path points to a sensitive file or directory.
// It blocks access to:
// - .env files and variants
// - .git directory
// - Server configuration files (config.yaml, config.json)
// - Database files (mcpany.db)
// - Private keys (id_rsa, id_dsa, *.pem, *.key)
//
// Summary: Checks for sensitive file patterns.
var IsSensitivePath = func(path string) error {
	base := filepath.Base(path)
	baseLower := strings.ToLower(base)

	// Block .env files
	if strings.HasPrefix(baseLower, ".env") {
		return fmt.Errorf("access to sensitive file %q is denied", base)
	}

	// Block .git directory
	if baseLower == ".git" {
		return fmt.Errorf("access to sensitive directory %q is denied", base)
	}

	// Block server configuration files
	if baseLower == "config.yaml" || baseLower == "config.yml" || baseLower == "config.json" {
		return fmt.Errorf("access to sensitive configuration file %q is denied", base)
	}

	// Block database files
	if baseLower == "mcpany.db" || strings.HasSuffix(baseLower, ".db") || strings.HasSuffix(baseLower, ".sqlite") {
		return fmt.Errorf("access to database file %q is denied", base)
	}

	// Block private keys
	if baseLower == "id_rsa" || baseLower == "id_dsa" || baseLower == "id_ed25519" || baseLower == "id_ecdsa" {
		return fmt.Errorf("access to private key %q is denied", base)
	}
	if strings.HasSuffix(baseLower, ".pem") || strings.HasSuffix(baseLower, ".key") {
		return fmt.Errorf("access to private key file %q is denied", base)
	}

	return nil
}

// IsAllowedPath checks if a given file path is allowed (inside CWD or AllowedPaths)
// and does not contain any path traversal sequences ("../").
// It is a variable to allow mocking in tests.
//
// Summary: Checks if a path is within allowed directories.
var IsAllowedPath = func(path string) error {
	// 1. Basic security check (no .. in the path string itself)
	if err := IsSecurePath(path); err != nil {
		return err
	}

	// 2. Check for sensitive files (on input path)
	if err := IsSensitivePath(path); err != nil {
		return err
	}

	// 3. Check Location (shared logic, sensitiveCheck=true)
	return checkLocation(path, true)
}

// IsAllowedLocation checks if a given file path is allowed (inside CWD or AllowedPaths)
// and does not contain any path traversal sequences ("../").
// Unlike IsAllowedPath, it does NOT check for sensitive files (like .pem), making it suitable for
// configuration fields that expect sensitive files (e.g., TLS certificates).
var IsAllowedLocation = func(path string) error {
	// 1. Basic security check (no .. in the path string itself)
	if err := IsSecurePath(path); err != nil {
		return err
	}

	// 2. Check Location (shared logic, sensitiveCheck=false)
	return checkLocation(path, false)
}

func checkLocation(path string, checkSensitive bool) error {
	// Resolve to absolute path
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

	// Check for sensitive files again (on resolved path) if requested
	if checkSensitive {
		if err := IsSensitivePath(realPath); err != nil {
			return fmt.Errorf("resolved path %q points to sensitive file: %w", realPath, err)
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

	// Check if inside CWD
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	if isInside(cwd, realPath) {
		return nil
	}

	// Check Allowed Paths
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

// allowedOpaqueSchemes are schemes that are allowed to not have a host component.
var allowedOpaqueSchemes = map[string]bool{
	"dns":         true,
	"unix":        true,
	"passthrough": true,
	"mailto":      true,
	"data":        true,
	"file":        true,
}

// IsValidURL checks if a given string is a valid URL. This function performs
// several checks, including for length, whitespace, the presence of a scheme,
// and host, considering special cases for schemes like "unix" or "mailto" that
// do not require a host.
//
// Summary: Validates a URL string.
//
// Parameters:
//   - s: string. The URL string.
//
// Returns:
//   - bool: True if valid.
func IsValidURL(s string) bool {
	if len(s) > 2048 || strings.TrimSpace(s) != s {
		return false
	}

	// ⚡ Bolt Optimization: Check for whitespace and control characters in a single pass over bytes.
	// This avoids:
	// 1. strings.Contains (scan)
	// 2. range loop (UTF-8 decoding)
	//
	// ASCII control characters are 0-31 and 127.
	// Space is 32.
	// So if b <= 32 || b == 127, it's invalid.
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b <= 32 || b == 127 {
			return false
		}
	}

	u, err := url.Parse(s)
	if err != nil {
		return false
	}

	if u.Scheme == "" {
		return false
	}
	// If a host is NOT present, the scheme must be one that allows an opaque part.
	if u.Host == "" {
		if !allowedOpaqueSchemes[u.Scheme] {
			return false
		}
		// Also ensure there is *something* after the scheme.
		if u.Opaque == "" && u.Path == "" {
			return false
		}
	} else if strings.HasPrefix(u.Host, ":") { // If a host is present, it must not be only a port.
		return false
	}

	return true
}

// ValidateHTTPServiceDefinition checks the validity of an HttpCallDefinition.
// It ensures that the endpoint path is specified and correctly formatted, and
// that a valid HTTP method is set.
//
// Summary: Validates an HTTP service definition.
//
// Parameters:
//   - def: *configv1.HttpCallDefinition. The definition to validate.
//
// Returns:
//   - error: An error if validation fails.
func ValidateHTTPServiceDefinition(def *configv1.HttpCallDefinition) error {
	if def == nil {
		return fmt.Errorf("http call definition cannot be nil")
	}
	if strings.TrimSpace(def.GetEndpointPath()) == "" {
		return fmt.Errorf("path is required")
	}
	if !strings.HasPrefix(def.GetEndpointPath(), "/") {
		return fmt.Errorf("path must start with a '/'")
	}
	u, err := url.Parse(def.GetEndpointPath())
	if err != nil {
		return fmt.Errorf("path contains invalid characters")
	}
	if u.RawQuery != "" {
		return fmt.Errorf("path must not contain query parameters")
	}
	if def.GetMethod() == configv1.HttpCallDefinition_HTTP_METHOD_UNSPECIFIED {
		return fmt.Errorf("method is required")
	}
	return nil
}

// FileExists checks if a file exists at the given path.
//
// Summary: Checks file existence.
var FileExists = func(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
}
