package doctor

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

func TestCheckService_Redaction_Redis_Bug(t *testing.T) {
	// This test simulates the bug where RedactDSN fails to redact passwords in DSNs
	// that fail URL parsing (e.g., redis://:password) and are passed to RedactDSN
	// via error messages.

	// Use a sensitive password
	password := "mysecretpassword"
	urlStr := "redis://:" + password

	service := configv1.UpstreamServiceConfig_builder{
		Name: stringPtr("test-redis-service"),
		HttpService: configv1.HttpUpstreamService_builder{
			// Using HttpService to trigger checkURL which calls http.NewRequest
			// which fails and calls RedactDSN(err.Error())
			Address: stringPtr(urlStr),
		}.Build(),
	}.Build()

	res := CheckService(context.Background(), service)

	t.Logf("Result Message: %s", res.Message)

	// The message should NOT contain the password
	if strings.Contains(res.Message, password) {
		t.Errorf("Security leak! Error message contains password: %s", res.Message)
	}

	// The message SHOULD contain [REDACTED] (asserting the fix works as intended)
	// Currently this will FAIL because of the bug.
	if !strings.Contains(res.Message, "[REDACTED]") {
		t.Errorf("Expected redaction in error message, but got: %s", res.Message)
	}
}
