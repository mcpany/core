package util

import (
	"strings"
	"testing"
)

func TestScanForSensitiveKeys_LongInput_Hunter(t *testing.T) {
	// Generate input > 128 bytes to trigger SIMD/long string optimization
	// Use '.' as padding to avoid creating false word boundaries (e.g. "passwordz" or "passwordx").
	padding := strings.Repeat(".", 200)

	// Case 1: Lowercase key (password) - triggers idxU == -1 path
	input1 := []byte(padding + "password" + padding)
	if !scanForSensitiveKeys(input1, false) {
		t.Error("Failed to find sensitive key in long input (lowercase)")
	}

	// Case 2: Uppercase key (PASSWORD) - triggers idxL == -1 path
	input2 := []byte(padding + "PASSWORD" + padding)
	if !scanForSensitiveKeys(input2, false) {
		t.Error("Failed to find sensitive key in long input (uppercase)")
	}

	// Case 3: Both (password ... PASSWORD) - triggers default path
	input3 := []byte(padding + "password" + "PASSWORD" + padding)
	if !scanForSensitiveKeys(input3, false) {
		t.Error("Failed to find sensitive key in long input (both)")
	}
}

func TestUnescapeKeySmall_BufferTooSmall(t *testing.T) {
	input := []byte("test")
	buf := make([]byte, 2) // Smaller than input
	_, ok := unescapeKeySmall(input, buf)
	if ok {
		t.Error("Expected failure when buffer is too small")
	}
}

func TestUnescapeKeySmall_EOF_InsideEscape(t *testing.T) {
	input := []byte("test\\") // Ends with backslash
	buf := make([]byte, 10)
	_, ok := unescapeKeySmall(input, buf)
	if ok {
		t.Error("Expected failure when input ends with backslash")
	}
}

func TestScanEscapedKeyForSensitive_Coverage(t *testing.T) {
	// Trigger scanEscapedKeyForSensitive by using invalid escape that json.Unmarshal fails on
	// \uZZZZ is invalid hex.

	// Also test unknown escape \q
	// Use space to avoid boundary check failure (passwordq looks like a word)
	key := []byte("password \\q")
	// json.Unmarshal will fail on \q.
	// It will fall back to scanEscapedKeyForSensitive.
	// "password" should be found.
	if !isKeySensitive(key) {
		t.Error("Failed to detect sensitive key with unknown escape")
	}

	// Test invalid hex \uZZZZ
	// Use space to avoid boundary check failure (passwordu looks like a word)
	key2 := []byte("password \\uZZZZ")
	if !isKeySensitive(key2) {
		t.Error("Failed to detect sensitive key with invalid hex escape")
	}

	// Test EOF inside escape
	// This is hard to pass to isKeySensitive because it assumes valid-ish json string content inside quotes?
	// But isKeySensitive takes raw bytes.
	key3 := []byte("password\\")
	// This might fail unescapeKeySmall and fall back to raw check?
	// unescapeKeySmall fails.
	// fall back to "keyToCheck = keyContent".
	// scanForSensitiveKeys("password\\", false) -> true.
	if !isKeySensitive(key3) {
		t.Error("Failed to detect sensitive key with trailing backslash")
	}
}

func TestScanEscapedKeyForSensitive_BufferFill(t *testing.T) {
	// Force scanEscapedKeyForSensitive usage by making key very large
	// maxUnescapeLimit is var, we can change it for test?
	// It is unexported var in redact_fast.go.
	// But since we are in same package, we can access it!

	originalLimit := maxUnescapeLimit
	maxUnescapeLimit = 10 // Set low limit to force streaming path
	defer func() { maxUnescapeLimit = originalLimit }()

	// create a key that is larger than 4097 (internal buffer of scanEscapedKeyForSensitive)
	// to trigger buffer refill logic.
	// 4097 + padding.

	// We want to test the boundary check logic.
	// "token" split across boundary.
	// internal bufSize is 4097. overlap 64.

	padding := strings.Repeat("x", 4096)
	// We want "token" to be at the end.
	// buf[4096] is 'z' (dummy).
	// valid data up to 4096.

	// If we put "token" at 4094..4099.
	// "to" at 4094, 4095.
	// "ken" at 4096, 4097, 4098.

	// escape chars counting?
	// scanEscapedKeyForSensitive unescapes.
	// So we pass plain text (no backslashes) to make it easy to count.

	input := padding + "token"
	// len = 4096 + 5 = 4101.
	// chunk 1: 0..4096. (4097 chars).
	// 4096 is 't'.
	// bufIdx will be 4097.
	// Trigger flush.
	// "token" is split?
	// "t" is at index 4096.
	// It should be preserved in overlap.

	if !scanEscapedKeyForSensitive([]byte(input)) {
		t.Error("Failed to find sensitive key across buffer boundary")
	}
}
