// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// computeHash calculates the SHA-256 hash of the audit entry fields.
// Format: "v1:<sha256_hex>"
func computeHash(ts, toolName, userID, profileID, args, result, errorMsg string, durationMs int64, prevHash string) string {
	// Construct the content string to be hashed using JSON array for collision resistance.
	// We include the version prefix "v1:" in the returned hash, but the content being hashed
	// is purely the JSON array of fields.

	// Use []any to serialize as a JSON array
	fields := []any{
		ts,
		toolName,
		userID,
		profileID,
		args,
		result,
		errorMsg,
		durationMs,
		prevHash,
	}

	// We ignore error here because we are encoding basic types that are always valid JSON
	contentBytes, _ := json.Marshal(fields)

	hash := sha256.Sum256(contentBytes)
	return "v1:" + hex.EncodeToString(hash[:])
}

// computeHashV0 calculates the legacy hash (pipe separated).
// Kept for backward compatibility verification or if we encounter old records.
func computeHashV0(ts, toolName, userID, profileID, args, result, errorMsg string, durationMs int64, prevHash string) string {
	content := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%d|%s",
		ts, toolName, userID, profileID, args, result, errorMsg, durationMs, prevHash)

	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
