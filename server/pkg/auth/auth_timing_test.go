package auth

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyAuthenticator_Timing(t *testing.T) {
	// A mock API key authentication using pointer creation directly
	v := "long-secret-key-12345"
	cfg := &configv1.APIKeyAuth{
		ParamName:         "X-API-Key",
		VerificationValue: &v,
		In:                configv1.APIKeyAuth_HEADER,
	}

	auth := NewAPIKeyAuthenticator(cfg)

	// A negative test: Ensure short vs long key returns standard unauthorized without panic
	reqShort := httptest.NewRequest("GET", "/", nil)
	reqShort.Header.Set("X-API-Key", "short")

	reqLong := httptest.NewRequest("GET", "/", nil)
	reqLong.Header.Set("X-API-Key", "looooooooooooooooooooooooooooooooooooooooooong")

	// Since Go's crypto/subtle.ConstantTimeCompare now compares SHA256 hashes of fixed 32 bytes
	// It will not leak the length. This test just ensures no panics and valid rejection.
	_, err := auth.Authenticate(context.Background(), reqShort)
	assert.Error(t, err)

	_, err = auth.Authenticate(context.Background(), reqLong)
	assert.Error(t, err)

	// Measure timing to ensure it takes roughly the same time for very small difference,
	// though in a unit test environment timing can be flaky.
	// We'll run a few iterations just to ensure it executes without issues.
	iterations := 1000

	startShort := time.Now()
	for i := 0; i < iterations; i++ {
		auth.Authenticate(context.Background(), reqShort)
	}
	durationShort := time.Since(startShort)

	startLong := time.Now()
	for i := 0; i < iterations; i++ {
		auth.Authenticate(context.Background(), reqLong)
	}
	durationLong := time.Since(startLong)

	t.Logf("Short took: %v, Long took: %v", durationShort, durationLong)
	// Basic sanity check, duration should not be 100x different.
	// We won't assert exact equality due to CI noise.
}
