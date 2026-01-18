// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

// VerifyIntegrity checks if the tool definition matches its expected hash.
func VerifyIntegrity(t *v1.Tool) error {
	if t.Integrity == nil {
		return nil // No integrity check required
	}

	if t.GetIntegrity().GetAlgorithm() != "sha256" {
		return fmt.Errorf("unsupported integrity algorithm: %s", t.GetIntegrity().GetAlgorithm())
	}

	// Create a copy of the tool without the integrity field to calculate the hash
	// We need to clone it because we are modifying it.
	toolCopy := proto.Clone(t).(*v1.Tool)
	toolCopy.Integrity = nil

	// Marshal to Binary for hashing - usually more deterministic for this purpose if we use the option
	marshaler := proto.MarshalOptions{
		Deterministic: true,
	}
	data, err := marshaler.Marshal(toolCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal tool for integrity check: %w", err)
	}

	hash := sha256.Sum256(data)
	calculatedHash := hex.EncodeToString(hash[:])

	if calculatedHash != t.GetIntegrity().GetHash() {
		return fmt.Errorf("integrity check failed: expected %s, got %s", t.GetIntegrity().GetHash(), calculatedHash)
	}

	return nil
}
