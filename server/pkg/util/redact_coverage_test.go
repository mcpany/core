package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanForSensitiveKeys_LongInput(t *testing.T) {
	// Create an input longer than 128 bytes
	// sensitiveKeys include "password"

	// Case 1: Long input with no sensitive keys
	longSafe := strings.Repeat("a", 200)
	assert.False(t, scanForSensitiveKeys([]byte(longSafe), false))

	// Case 2: Long input with sensitive key at the end
	longSensitive := strings.Repeat("a", 200) + "password"
	assert.True(t, scanForSensitiveKeys([]byte(longSensitive), false))

	// Case 3: Long input with sensitive key in the middle
    // Note: Use '.' as filler after password to ensure boundary check passes.
    // If we use 'b', it looks like "passwordbbbb..." which is treated as a different word.
	longSensitiveMid := strings.Repeat("a", 100) + "password" + strings.Repeat(".", 100)
	assert.True(t, scanForSensitiveKeys([]byte(longSensitiveMid), false))

    // Case 4: Long input with sensitive key that requires case folding match
    // "Password"
    longSensitiveUpper := strings.Repeat("a", 200) + "Password"
    assert.True(t, scanForSensitiveKeys([]byte(longSensitiveUpper), false))
}

func TestScanForSensitiveKeys_EdgeCases(t *testing.T) {
    // Test checkPotentialMatch boundary conditions

    // "auth" is a key. "author" is not (boundary check).
    // input: "author"
    assert.False(t, scanForSensitiveKeys([]byte("author"), false))

    // "auth" matches.
    assert.True(t, scanForSensitiveKeys([]byte("auth"), false))

    // CamelCase boundary: "authToken" -> matches "auth" if next is uppercase?
    // Code says:
    // if next >= 'A' && next <= 'Z' {
    //    if firstChar >= 'A' && firstChar <= 'Z' { continue }
    // }
    // "authToken": 'T' is upper. 'a' is lower. So it does NOT continue. It matches.
    assert.True(t, scanForSensitiveKeys([]byte("authToken"), false))

    // "AUTH" in "AUTHORITY"
    // "AUTH" matches "auth". First char 'A' is upper. Next 'O' is upper.
    // It continues (skips).
    assert.False(t, scanForSensitiveKeys([]byte("AUTHORITY"), false))
}
