// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"

// CFIAConfig defines the configuration for Context-File Integrity Attestation.
//
// Summary: Configuration for Context-File Integrity Attestation (CFIA) Middleware.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type CFIAConfig struct {
	// Enabled determines if the CFIA middleware is active.
	Enabled bool `json:"enabled"`
	// TargetTools lists the tools that should be inspected for file reads (e.g., fs:read, read_file).
// CFIAMiddleware implements Context-File Integrity Attestation.
// It intercepts requests to read local files, calculates their hashes,
// and ensures they match known-good, hardware-attested manifests to
// prevent Deceptive Context Injection.
//
// Summary: Represents the CFIA Middleware.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type CFIAMiddleware struct {
	config CFIAConfig
}

// NewCFIAMiddleware creates a new CFIAMiddleware instance.
//
// Summary: Creates a new Context-File Integrity Attestation middleware instance.
//
// Parameters:
//   - config (CFIAConfig): The configuration settings.
//
// Returns:
//   - *CFIAMiddleware: The resulting CFIA middleware instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewCFIAMiddleware(config CFIAConfig) *CFIAMiddleware {
	return &CFIAMiddleware{
		config: config,
	}
}

// Execute enforces context-file integrity before proceeding to the next handler.
//
// Summary: Executes the Context-File Integrity Attestation check on the request.
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
func (m *CFIAMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if !m.config.Enabled {
		return next(ctx, req)
	}

	// Check if the current tool is in the target list
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

	logger := logging.GetLogger().With("component", "cfia_middleware")

	// Extract the file path argument
	if req.Arguments == nil {
		return next(ctx, req)
	}

	argName := m.config.ArgumentName
	if argName == "" {
		argName = "path" // default
	}

	pathRaw, hasPath := req.Arguments[argName]
	if !hasPath {
		return next(ctx, req)
	}

	path, ok := pathRaw.(string)
	if !ok || path == "" {
		return next(ctx, req) // Could be a dir or invalid arg type, let upstream handle
	}

	// Check if this path requires attestation
	expectedHash, requiresAttestation := m.config.AttestedHashes[path]
	if !requiresAttestation {
		return next(ctx, req)
	}

	logger.Debug("CFIA Middleware intercepted target file access", "tool", req.ToolName, "path", path)

	// Read and hash the file
	content, err := os.ReadFile(path)
	if err != nil {
		logger.Warn("CFIA Broker rejected request: failed to read file for hashing", "path", path, "error", err)
		return nil, fmt.Errorf("Context-File Integrity Attestation Failed: unable to read file '%s'", path)
	}

	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	if hash != expectedHash {
		logger.Warn("CFIA Broker rejected request: hash mismatch detected (Deceptive Context Injection suspected)", "path", path, "expected", expectedHash, "actual", hash)
		return nil, fmt.Errorf("Context-File Integrity Attestation Failed: hash mismatch for file '%s'. Deceptive Context Injection suspected", path)
	}

	logger.Info("CFIA validation passed", "path", path)

	// Context verified, proceed
	return next(ctx, req)
}
