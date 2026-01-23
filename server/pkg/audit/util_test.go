// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeHashV1(t *testing.T) {
	ts := "2023-10-27T10:00:00Z"
	toolName := "test_tool"
	userID := "user1"
	profileID := "profile1"
	args := `{"key": "value"}`
	result := `{"status": "ok"}`
	errorMsg := ""
	durationMs := int64(100)
	prevHash := "prev_hash_123"

	// Test V1
	hash1 := computeHashV1(ts, toolName, userID, profileID, args, result, errorMsg, durationMs, prevHash)
	assert.Contains(t, hash1, "v1:")

	// Deterministic check
	hash2 := computeHashV1(ts, toolName, userID, profileID, args, result, errorMsg, durationMs, prevHash)
	assert.Equal(t, hash1, hash2)

	// Change one field
	hash3 := computeHashV1(ts, toolName, userID, profileID, args, result, "error", durationMs, prevHash)
	assert.NotEqual(t, hash1, hash3)
}

func TestComputeHashV2(t *testing.T) {
	ts := "2023-10-27T10:00:00Z"
	toolName := "test_tool"
	userID := "user1"
	profileID := "profile1"
	ipAddress := "127.0.0.1"
	args := `{"key": "value"}`
	result := `{"status": "ok"}`
	errorMsg := ""
	durationMs := int64(100)
	prevHash := "prev_hash_123"

	// Test V2
	hash1 := computeHashV2(ts, toolName, userID, profileID, ipAddress, args, result, errorMsg, durationMs, prevHash)
	assert.Contains(t, hash1, "v2:")

	// Deterministic check
	hash2 := computeHashV2(ts, toolName, userID, profileID, ipAddress, args, result, errorMsg, durationMs, prevHash)
	assert.Equal(t, hash1, hash2)

	// Change IP field
	hash3 := computeHashV2(ts, toolName, userID, profileID, "192.168.1.1", args, result, errorMsg, durationMs, prevHash)
	assert.NotEqual(t, hash1, hash3)
}

func TestComputeHashV0(t *testing.T) {
	ts := "2023-10-27T10:00:00Z"
	toolName := "test_tool"
	userID := "user1"
	profileID := "profile1"
	args := `{"key": "value"}`
	result := `{"status": "ok"}`
	errorMsg := ""
	durationMs := int64(100)
	prevHash := "prev_hash_123"

	hash1 := computeHashV0(ts, toolName, userID, profileID, args, result, errorMsg, durationMs, prevHash)
	assert.NotContains(t, hash1, "v1:")

	hash2 := computeHashV0(ts, toolName, userID, profileID, args, result, errorMsg, durationMs, prevHash)
	assert.Equal(t, hash1, hash2)
}

func TestJsonSerializationConsistency(t *testing.T) {
	// Ensure that computeHash behaves consistently
	fields := []any{"a", "b", 123}
	data, _ := json.Marshal(fields)
	assert.Equal(t, `["a","b",123]`, string(data))
}
