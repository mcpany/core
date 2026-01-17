package util

import (
	"strings"
	"testing"
)

// Export for testing
var ScanForSensitiveKeys = scanForSensitiveKeys

func BenchmarkScanForSensitiveKeys(b *testing.B) {
	// 1KB string with no matches
	longSafe := strings.Repeat("not_sensitive_string_", 50)
	longSafeBytes := []byte(longSafe)

	// 1KB string with match at the end
    // "api_key" starts with 'a', which is first in sensitiveStartChars.
    // Baseline checks 'a' first, so it finds it fast, ignoring 's' and 't' in the prefix.
    // New approach stops at 't' and 's' first.
	longSensitive := longSafe + "api_key"
	longSensitiveBytes := []byte(longSensitive)

    // Match starting with 'x', which is last in sensitiveStartChars.
    // Baseline checks a, t, s, p, c THEN x.
    // "not_sensitive_string_" has 's' and 't'.
    // So baseline will still stop at 's' and 't' during their passes?
    // "not_sensitive_string_" has 't' (token), 's' (secret).
    // Baseline:
    // Pass 'a': scans full string (miss).
    // Pass 't': finds 't's, checks them (miss).
    // Pass 's': finds 's's, checks them (miss).
    // ...
    // Pass 'x': finds 'x' (hit).
    //
    // New:
    // Pass ANY: finds 't', check, finds 's', check... finds 'x', check (hit).
    //
    // The difference is that baseline ALSO did full scan for 'a', 'p', 'c'.
    // New saves those scans.
    longSensitiveX := longSafe + "x-api-key"
    longSensitiveXBytes := []byte(longSensitiveX)

    // 1KB string with many potential start chars but no actual matches
    // 'a', 'c', 'p', 's', 't', 'x' are start chars.
    // "apple", "cat", "pass", "snake", "test", "xylophone"
    longFalsePositives := strings.Repeat("apple_cat_pass_snake_test_xylophone_", 20)
    longFalsePositivesBytes := []byte(longFalsePositives)

	b.Run("LongSafe", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ScanForSensitiveKeys(longSafeBytes, false)
		}
	})

	b.Run("LongSensitive", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ScanForSensitiveKeys(longSensitiveBytes, false)
		}
	})

    b.Run("LongSensitiveX", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ScanForSensitiveKeys(longSensitiveXBytes, false)
		}
	})

    b.Run("LongFalsePositives", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            ScanForSensitiveKeys(longFalsePositivesBytes, false)
        }
    })
}
