package tool

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

// VerifyIntegrity checks if the tool definition matches its expected hash.
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

// VerifyConfigIntegrity checks if the config tool definition matches its expected hash.
// Currently ToolDefinition in config does not have an integrity field, so this is a placeholder.
func VerifyConfigIntegrity(_ *configv1.ToolDefinition) error {
	return nil
}

// CalculateHash computes the SHA256 hash of a runtime tool definition.
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

// CalculateConfigHash computes the SHA256 hash of a configuration tool definition.
func CalculateConfigHash(t *configv1.ToolDefinition) (string, error) {
	// Create a copy of the tool to calculate the hash
	toolCopy := proto.Clone(t).(*configv1.ToolDefinition)

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
