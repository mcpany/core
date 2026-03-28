// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
)

// CCIGConfig defines the configuration for CI/CD Cache Integrity Guard.
//
// Summary: Configuration for CI/CD Cache Integrity Guard (CCIG) Middleware.
type CCIGConfig struct {
	// Enabled determines if the CCIG middleware is active.
	Enabled bool `json:"enabled"`
	// TargetTools lists the tools that interact with build caches (e.g., fs:read, cache:restore).
	TargetTools []string `json:"target_tools"`
	// AttestedHashes maps cache file paths to their expected, hardware-attested SHA-256 hashes.
	AttestedHashes map[string]string `json:"attested_hashes"`
	// ArgumentName specifies the argument name that contains the file path in the tool request.
	ArgumentName string `json:"argument_name"`
}

// CCIGMiddleware implements CI/CD Cache Integrity Guard.
// It intercepts requests to read cache files, calculates their hashes,
// and ensures they match known-good, hardware-attested manifests to
// prevent supply chain poisoning.
//
// Summary: Represents the CCIG Middleware.
type CCIGMiddleware struct {
	config CCIGConfig
}

// NewCCIGMiddleware creates a new CCIGMiddleware instance.
//
// Summary: Creates a new CI/CD Cache Integrity Guard middleware instance.
//
// Parameters:
//   - config (CCIGConfig): The configuration settings.
//
// Returns:
//   - *CCIGMiddleware: The resulting CCIG middleware instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewCCIGMiddleware(config CCIGConfig) *CCIGMiddleware {
	return &CCIGMiddleware{
		config: config,
	}
}

// Execute enforces cache file integrity before proceeding to the next handler.
//
// Summary: Executes the CI/CD Cache Integrity Guard check on the request.
//
// Parameters:
//   - ctx (context.Context): The execution context.
//   - req (*tool.ExecutionRequest): The tool execution request.
//   - next (tool.ExecutionFunc): The next handler in the chain.
//
// Returns:
//   - any: The execution result if the file hash matches or isn't required.
//   - error: An error if integrity verification fails (hash mismatch or missing).
//
// Errors:
//   - Returns an error if the requested file's hash does not match the attested hash.
//   - Returns an error if the file cannot be read for hashing.
//
// Side Effects:
//   - Reads the target file from the filesystem.
//   - Logs validation outcomes (success or failure).
func (m *CCIGMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if !m.config.Enabled {
		return next(ctx, req)
	}

	isTarget := false
	for _, t := range m.config.TargetTools {
		if t == req.ToolName {
			isTarget = true
			break
		}
	}

	if !isTarget {
		return next(ctx, req)
	}

	logger := logging.GetLogger().With("component", "ccig_middleware")

	if req.Arguments == nil {
		return next(ctx, req)
	}

	argName := m.config.ArgumentName
	if argName == "" {
		argName = "path" // default argument name
	}

	pathRaw, hasPath := req.Arguments[argName]
	if !hasPath {
		return next(ctx, req)
	}

	path, ok := pathRaw.(string)
	if !ok || path == "" {
		return next(ctx, req)
	}

	// Normalize path to prevent directory traversal bypasses
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		logger.Warn("CCIG Broker rejected request: invalid path provided", "path", path, "error", err)
		return nil, fmt.Errorf("CI/CD Cache Integrity Guard Failed: invalid path '%s'", path)
	}

	// Check against the normalized absolute path. We assume AttestedHashes uses absolute normalized paths.
	// If it uses relative paths (not recommended), we should normalize both, but absolute is safer.
	// To be robust against configs using relative paths or absolute paths interchangeably,
	// let's check against the provided exact configuration keys which users provide,
	// but after cleaning the input. However, since AttestedHashes is a map, it's a strict string match.
	// We'll normalize the map keys during initialization or assume the user configures them as clean paths.
	// The safest approach is to match against the clean path, and if not found, try the absolute path.
	expectedHash, requiresAttestation := m.config.AttestedHashes[cleanPath]
	if !requiresAttestation {
		expectedHash, requiresAttestation = m.config.AttestedHashes[absPath]
	}

	if !requiresAttestation {
		return next(ctx, req)
	}

	logger.Debug("CCIG Middleware intercepted target file access", "tool", req.ToolName, "path", path, "cleanPath", cleanPath, "absPath", absPath)

	file, err := os.Open(cleanPath)
	if err != nil {
		logger.Warn("CCIG Broker rejected request: failed to open file for hashing", "path", cleanPath, "error", err)
		return nil, fmt.Errorf("CI/CD Cache Integrity Guard Failed: unable to read file '%s'", cleanPath)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		logger.Warn("CCIG Broker rejected request: failed to hash file", "path", cleanPath, "error", err)
		return nil, fmt.Errorf("CI/CD Cache Integrity Guard Failed: unable to hash file '%s'", cleanPath)
	}

	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	if hash != expectedHash {
		logger.Warn("CCIG Broker rejected request: hash mismatch detected (Supply Chain Poisoning suspected)", "path", cleanPath, "expected", expectedHash, "actual", hash)
		return nil, fmt.Errorf("CI/CD Cache Integrity Guard Failed: hash mismatch for file '%s'. Supply Chain Poisoning suspected", cleanPath)
	}

	logger.Info("CCIG validation passed", "path", path)

	return next(ctx, req)
}
