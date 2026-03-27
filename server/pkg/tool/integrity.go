// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

// VerifyIntegrity provides verifyintegrity functionality.
//
// Summary: VerifyIntegrity.
//
// Parameters.
//   - t: The parameter.
//
// Returns.
//   - result: The result.
func VerifyIntegrity(t *v1.Tool) error {
	if !t.HasIntegrity() {
		return nil // No integrity check required
	}

	if t.GetIntegrity().GetAlgorithm() != "sha256" {
		return fmt.Errorf("unsupported integrity algorithm: %s", t.GetIntegrity().GetAlgorithm())
	}

	calculatedHash, err := CalculateHash(t)
	if err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}

	if calculatedHash != t.GetIntegrity().GetHash() {
		return fmt.Errorf("integrity check failed: expected %s, got %s", t.GetIntegrity().GetHash(), calculatedHash)
	}

	return nil
}

// VerifyConfigIntegrity provides verifyconfigintegrity functionality.
//
// Summary: VerifyConfigIntegrity.
//
// Parameters.
//   - t: The parameter.
//
// Returns.
//   - result: The result.
func VerifyConfigIntegrity(t *configv1.ToolDefinition) error {
	if t.GetIntegrity() == nil {
		return nil // No integrity check required
	}

	if t.GetIntegrity().GetAlgorithm() != "sha256" {
		return fmt.Errorf("unsupported integrity algorithm: %s", t.GetIntegrity().GetAlgorithm())
	}

	calculatedHash, err := CalculateConfigHash(t)
	if err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}

	if calculatedHash != t.GetIntegrity().GetHash() {
		return fmt.Errorf("integrity check failed: expected %s, got %s", t.GetIntegrity().GetHash(), calculatedHash)
	}

	return nil
}

// CalculateHash provides calculatehash functionality.
//
// Summary: CalculateHash.
//
// Parameters.
//   - t: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func CalculateHash(t *v1.Tool) (string, error) {
	// Create a copy of the tool without the integrity field to calculate the hash
	toolCopy := proto.Clone(t).(*v1.Tool)
	toolCopy.SetIntegrity(nil)

	// Marshal to Binary for hashing - deterministic
	marshaler := proto.MarshalOptions{
		Deterministic: true,
	}
	data, err := marshaler.Marshal(toolCopy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tool for integrity check: %w", err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// CalculateConfigHash provides calculateconfighash functionality.
//
// Summary: CalculateConfigHash.
//
// Parameters.
//   - t: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func CalculateConfigHash(t *configv1.ToolDefinition) (string, error) {
	// Create a copy of the tool to calculate the hash
	toolCopy := proto.Clone(t).(*configv1.ToolDefinition)
	toolCopy.SetIntegrity(nil)

	// Marshal to Binary for hashing - deterministic
	marshaler := proto.MarshalOptions{
		Deterministic: true,
	}
	data, err := marshaler.Marshal(toolCopy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tool for integrity check: %w", err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}
