package util

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_FalsePositive_Chunking(t *testing.T) {
	// We want to trigger a false positive for "authenticity" (which starts with "auth").
	// "auth" is sensitive. "authenticity" is not.
	// We need "auth" to fall exactly at the end of a chunk (4096 bytes).
	// scanEscapedKeyForSensitive is ONLY used if len(key) > maxUnescapeLimit (1MB).
	// So we need to construct a key > 1MB.

	// 1MB = 1048576 bytes.
	// We use a large number of chunks to ensure we exceed the limit.
	// Let's use 300 chunks of 4096 bytes.
	// 300 * 4096 = 1,228,800 bytes.
	chunkSize := 4096
	numChunks := 300
	targetEnd := numChunks * chunkSize

	// We want "auth" (4 bytes) to end exactly at targetEnd.
	// So padding length should be targetEnd - 4.
	paddingLen := targetEnd - 4
	padding := strings.Repeat("x", paddingLen)

	// We use "au\u0074h" to force the key to be treated as "containing escapes".
	// "au\u0074h" unescapes to "auth" (4 bytes).
	// Note: The length of the raw key string will be larger due to escape sequence,
	// which helps ensure we exceed maxUnescapeLimit.
	prefix := padding + `au\u0074h`
	suffix := "enticity"

	key := prefix + suffix

	// Verify the raw key length exceeds the default limit (1MB)
	// padding is ~1.2MB, so we are safe.
	if len(key) <= 1024*1024 {
		t.Fatalf("Key length %d is not > 1MB", len(key))
	}

	// Construct JSON
	// We need to be careful with memory, constructing huge JSON might be slow/heavy,
	// but 1.2MB is manageable in Go tests.
	input := []byte(`{"` + key + `": "safe_value"}`)

	output := RedactJSON(input)

	// Parse output
	var m map[string]interface{}
	err := json.Unmarshal(output, &m)
	assert.NoError(t, err)

	// Reconstruct the expected key (unescaped)
	expectedKey := strings.Repeat("x", paddingLen) + "authenticity"

	val, ok := m[expectedKey]
	if !ok {
		t.Fatalf("Key not found in output")
	}

	// It should NOT be redacted because "authenticity" is not sensitive.
	// If the bug exists, "auth" will be detected at the chunk boundary, causing redaction.
	assert.Equal(t, "safe_value", val, "Key should NOT be redacted (false positive due to chunking)")
}
