// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
)

// CFIAConfig represents the public CFIAConfig entity.
//
// Summary: Defines the structured data model representing a config.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type CFIAConfig struct {
	// Enabled determines if the CFIA middleware is active.
	Enabled bool `json:"enabled"`
	// TargetTools lists the tools that should be inspected for file reads (e.g., fs:read, read_file).
	TargetTools []string `json:"target_tools"`
	// AttestedHashes maps file paths to their expected, hardware-attested SHA-256 hashes.
	AttestedHashes map[string]string `json:"attested_hashes"`
	// ArgumentName specifies the argument name that contains the file path in the tool request.
	ArgumentName string `json:"argument_name"`
}

// CFIAMiddleware represents the public CFIAMiddleware entity.
//
// Summary: Defines the structured data model representing a middleware.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type CFIAMiddleware struct {
	config CFIAConfig
}

// NewCFIAMiddleware serves as a public interface for interacting with NewCFIAMiddleware.
//
// Summary: Constructs and returns an initialized cfia middleware ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewCFIAMiddleware(config CFIAConfig) *CFIAMiddleware {
	return &CFIAMiddleware{
		config: config,
	}
}

// Execute serves as a public interface for interacting with Execute.
//
// Summary: Execute the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
