// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// computeHash computes the hash for the audit entry using SHA-256.
// It uses a JSON array structure for unambiguous serialization.
<<<<<<< HEAD
//
// Summary: Computes hash for audit entry.
//
// Parameters:
//   - timestamp (string): The timestamp.
//   - toolName (string): The tool name.
//   - userID (string): The user ID.
//   - profileID (string): The profile ID.
//   - args (string): The arguments.
//   - result (string): The result.
//   - errorMsg (string): The error message.
//   - durationMs (int64): The duration in ms.
//   - prevHash (string): The previous hash.
//
// Returns:
//   - string: The computed hash.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
func computeHash(timestamp, toolName, userID, profileID, args, result, errorMsg string, durationMs int64, prevHash string) string {
	// Use JSON array for unambiguous serialization
	fields := []any{timestamp, toolName, userID, profileID, args, result, errorMsg, durationMs, prevHash}
	data, _ := json.Marshal(fields) // Error ignored as primitive types/strings should always marshal
	h := sha256.Sum256(data)
	return "v1:" + hex.EncodeToString(h[:])
}

// computeHashV0 computes the hash using the legacy method (vulnerable to collision).
// Kept for backward compatibility verification.
<<<<<<< HEAD
//
// Summary: Computes hash using legacy method.
//
// Parameters:
//   - timestamp (string): The timestamp.
//   - toolName (string): The tool name.
//   - userID (string): The user ID.
//   - profileID (string): The profile ID.
//   - args (string): The arguments.
//   - result (string): The result.
//   - errorMsg (string): The error message.
//   - durationMs (int64): The duration in ms.
//   - prevHash (string): The previous hash.
//
// Returns:
//   - string: The computed hash.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
func computeHashV0(timestamp, toolName, userID, profileID, args, result, errorMsg string, durationMs int64, prevHash string) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%d|%s",
		timestamp, toolName, userID, profileID, args, result, errorMsg, durationMs, prevHash)
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}
