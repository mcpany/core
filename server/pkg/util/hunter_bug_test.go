package util

import (
	"math"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestBytesToString(t *testing.T) {
	b := []byte("hello")
	s := BytesToString(b)
	assert.Equal(t, "hello", s)

	// Ensure no allocation (hard to test without benchmarks, but functionally it works)
}

func TestToString_FloatBoundary(t *testing.T) {
	// math.MaxInt64 = 9223372036854775807
	// float64(math.MaxInt64) = 9.223372036854776e+18 (rounded up to nearest even)
	// 9223372036854775807 is 0x7FFFFFFFFFFFFFFF
	// float64 representation loses precision.

	// Case 1: MaxInt64 exactly
	// It will be converted to float64, losing precision.
	// val = 9.223372036854776e+18
	// val >= float64(math.MinInt64) is true.
	// val < float64(math.MaxInt64) ?
	// float64(math.MaxInt64) is the same value. So val < val is false.
	// So it should fall through to FormatFloat.

	val := float64(math.MaxInt64)
	s := ToString(val)
	// We expect scientific notation or float string, NOT int string (which would be incorrect due to precision loss)
	assert.Contains(t, s, "e+", "Expected scientific notation for MaxInt64 as float64")

	// Case 2: A value that is safely representable in int64, but passed as float64.
	// 2^53 is safe limit for integer precision in float64.
	// let's try 2^60. 1152921504606846976
	valSafe := float64(1 << 60)
	sSafe := ToString(valSafe)
	assert.Equal(t, "1152921504606846976", sSafe)

	// Case 3: Negative boundary
	// math.MinInt64 = -9223372036854775808
	// float64(math.MinInt64) is exact because it's -2^63 (power of 2)
	valMin := float64(math.MinInt64)
	// val >= float64(math.MinInt64) is true (-2^63 >= -2^63)
	// val < float64(math.MaxInt64) is true (-2^63 < 2^63)
	// So it attempts to convert to int64.
	// int64(-2^63) is valid (math.MinInt64).
	sMin := ToString(valMin)
	assert.Equal(t, "-9223372036854775808", sMin)
}

func TestSanitizeID_DotInjection(t *testing.T) {
	// If I pass ["a.b"], I expect it to be sanitized because dot is not in allowedSanitizeIDChars.
	// allowedSanitizeIDChars includes alphanumeric, _, -
	// So "a.b" -> "ab_HASH"

	res, err := SanitizeID([]string{"a.b"}, false, 100, 8)
	assert.NoError(t, err)
	assert.NotEqual(t, "a.b", res)
	assert.Contains(t, res, "ab_")
}
