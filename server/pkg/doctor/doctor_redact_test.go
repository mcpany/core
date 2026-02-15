package doctor

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

func stringPtr(s string) *string {
	return &s
}

func TestCheckService_Redaction_Mailto_HTTP(t *testing.T) {
	// This integration test verifies that the doctor check, which uses RedactDSN internally,
	// does NOT redact mailto: links in error messages.
	// This covers the flow: CheckService -> checkHTTPService -> checkURL -> RedactDSN.

	urlStr := "mailto:bob@example.com"
	service := configv1.UpstreamServiceConfig_builder{
		Name: stringPtr("test-http-service"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: stringPtr(urlStr),
		}.Build(),
	}.Build()

	res := CheckService(context.Background(), service)

	t.Logf("Result Message: %s", res.Message)

	// If RedactDSN was buggy, it would replace "mailto:bob@example.com" with "mailto:[REDACTED]@example.com".
	if strings.Contains(res.Message, "[REDACTED]") {
		t.Errorf("Expected mailto link not to be redacted in error message, but got: %s", res.Message)
	}

	// Double check that the message actually contains the URL (to ensure we are testing the right thing)
	if !strings.Contains(res.Message, "mailto:bob@example.com") {
		t.Errorf("Expected error message to contain original URL, but got: %s", res.Message)
	}
}
