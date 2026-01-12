package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanForSensitiveKeys_LargeInput(t *testing.T) {
	t.Parallel()
    // Test the large input optimization path in scanForSensitiveKeys (len > 128)

    // Construct a large string > 128 bytes
    prefix := strings.Repeat("not_sensitive_", 10) // 14 * 10 = 140 bytes
    input := prefix + "api_key"

    // Should find it
    assert.True(t, scanForSensitiveKeys([]byte(input), false))

    // Should not find if missing
    assert.False(t, scanForSensitiveKeys([]byte(prefix), false))

    // Test mixed case with large input
    inputMixed := prefix + "API_KEY"
    assert.True(t, scanForSensitiveKeys([]byte(inputMixed), false))

    // Test match at the very end
    inputEnd := prefix + "token"
    assert.True(t, scanForSensitiveKeys([]byte(inputEnd), false))
}

func TestIsKey(t *testing.T) {
    t.Parallel()
    // Test isKey function directly

    // Valid key scenarios
    // isKey starts scanning at startOffset.

    assert.True(t, isKey([]byte(`":`), 0))
    assert.True(t, isKey([]byte(`" :`), 0))

    // Invalid key scenarios
    assert.False(t, isKey([]byte(`" ,`), 0))
    assert.False(t, isKey([]byte(`" }`), 0))
    assert.False(t, isKey([]byte(`"`), 0)) // EOF

    // Max scan limit
    // Should return true (conservative) if limit reached
    assert.True(t, isKey([]byte(strings.Repeat("a", 300)), 0))

    // Escape sequence (conservative)
    // If it finds escape, it returns true immediately
    assert.True(t, isKey([]byte(`\"notquote"`), 0))
}

func TestIsKeyColon(t *testing.T) {
    t.Parallel()

    assert.True(t, isKeyColon([]byte(`: value`), 0))
    assert.True(t, isKeyColon([]byte(`   : value`), 0))
    assert.False(t, isKeyColon([]byte(` value`), 0))
    assert.False(t, isKeyColon([]byte(``), 0))
}

func TestCheckPotentialMatch_NextCharMask(t *testing.T) {
    t.Parallel()
    // Test the sensitiveNextCharMask optimization

    // "abracadabra" starts with 'a'. Second char 'b'.
    // 'b' is not in the mask for 'a'.
    // So it should quickly return false.
    assert.False(t, scanForSensitiveKeys([]byte("abracadabra"), false))

    // "access" starts with 'a'. Second char 'c'.
    // 'c' is not in mask for 'a'.
    assert.False(t, scanForSensitiveKeys([]byte("access"), false))
}

func TestScanForSensitiveKeys_OptimizedLoop(t *testing.T) {
    t.Parallel()
    // Test the optimized loop for large inputs (>128) with multiple potential matches

    // Create input with many 'a's but no full match, then a match
    // "aaaa..."
    input := strings.Repeat("a", 200) + "pi_key"
    // "aaaa...pi_key" -> ends with "api_key"
    assert.True(t, scanForSensitiveKeys([]byte(input), false))

    // Uppercase start
    inputUp := strings.Repeat("A", 200) + "PI_KEY"
    assert.True(t, scanForSensitiveKeys([]byte(inputUp), false))
}

func TestSkipJSONValue_Coverage(t *testing.T) {
    t.Parallel()
    // Cover cases in skipJSONValue, skipObject, skipArray, etc.

    // skipObject with strings containing braces
    jsonObj := `{"a": "{", "b": "}"}`
    assert.Equal(t, len(jsonObj), skipObject([]byte(jsonObj), 0))

    // skipArray with strings containing brackets
    jsonArr := `["[", "]"]`
    assert.Equal(t, len(jsonArr), skipArray([]byte(jsonArr), 0))

    // skipLiteral
    jsonLit := `null,`
    // skipLiteral returns index after 'null'
    assert.Equal(t, 4, skipLiteral([]byte(jsonLit), 0))

    // skipNumber
    jsonNum := `1234,`
    assert.Equal(t, 4, skipNumber([]byte(jsonNum), 0))
}
