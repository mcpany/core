package util

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRedactDSNNumericPassword(t *testing.T) {
	// Case 1: Redis with numeric password and no user
	// redis://:123456@localhost:6379
	// Should redact 123456
	dsn := "redis://:123456@localhost:6379"
	redacted := RedactDSN(dsn)
	assert.NotContains(t, redacted, "123456")
	assert.Contains(t, redacted, "[REDACTED]")

	// Case 2: Just numeric password (fallback regex might trigger)
	// redis://:123456
	// This looks like a port in regex if we are not careful, but it has no host before it?
	// The regex is (://[^:]*):([^/@\s"?]+)
	// redis:// matches group 1.
	// 123456 matches group 2.
	// It is purely numeric.
	// If the code skips numeric passwords assuming they are ports, this will fail.
	dsn2 := "redis://:123456"
	redacted2 := RedactDSN(dsn2)
	assert.NotContains(t, redacted2, "123456", "Should redact numeric password even if it looks like a port")
}
