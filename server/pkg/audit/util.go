package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// computeHash computes the hash for the audit entry using SHA-256.
// It uses a JSON array structure for unambiguous serialization.
func computeHash(timestamp, toolName, userID, profileID, args, result, errorMsg string, durationMs int64, prevHash string) string {
	// Use JSON array for unambiguous serialization
	fields := []any{timestamp, toolName, userID, profileID, args, result, errorMsg, durationMs, prevHash}
	data, _ := json.Marshal(fields) // Error ignored as primitive types/strings should always marshal
	h := sha256.Sum256(data)
	return "v1:" + hex.EncodeToString(h[:])
}

// computeHashV0 computes the hash using the legacy method (vulnerable to collision).
// Kept for backward compatibility verification.
func computeHashV0(timestamp, toolName, userID, profileID, args, result, errorMsg string, durationMs int64, prevHash string) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%d|%s",
		timestamp, toolName, userID, profileID, args, result, errorMsg, durationMs, prevHash)
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}
