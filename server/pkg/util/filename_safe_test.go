package util

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeFilename_Truncation(t *testing.T) {
	// Create a long filename with multibyte characters that ARE allowed (IsLetter=true)
	// '好' is 3 bytes and IsLetter=true.

	// We want to align it so truncation splits the character.
	// 254 'a's + '好'.
	// Length = 257.
	// Truncate at 255 -> 254 'a's + 1st byte of '好'.
	// This leaves an incomplete UTF-8 sequence.
	// With the fix, it should be truncated to 254 bytes (removing the partial char).

	prefix := strings.Repeat("a", 254)
	longName := prefix + "好"
	sanitized := SanitizeFilename(longName)

	assert.Equal(t, 254, len(sanitized), "Length should be 254 (partial char removed)")
	assert.True(t, utf8.ValidString(sanitized), "String should be valid UTF-8")
}
