// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
	service := &configv1.UpstreamServiceConfig{
		Name: stringPtr("test-http-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: stringPtr(urlStr),
			},
		},
	}

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

func TestCheckService_Redaction_Bug_E2E(t *testing.T) {
	// This test specifically targets the bug where passwords with slashes
	// in malformed URLs (causing fallback to regex) were not redacted.
	// We use an opaque URL "http:..." instead of "http://..." to bypass the
	// dsnSchemeRegex (which handles "://") and force usage of dsnPasswordRegex.
	// Input: http:user:pass/word@host%  (The % at the end causes url.Parse to fail with "invalid URL escape")
	// Expected Error Message: parse "http:user:[REDACTED]@host%": ...

	urlStr := "http:user:pass/word@host%"
	service := &configv1.UpstreamServiceConfig{
		Name: stringPtr("test-redact-bug"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: stringPtr(urlStr),
			},
		},
	}

	res := CheckService(context.Background(), service)

	t.Logf("Result Message: %s", res.Message)

	// Check that we got an error (due to invalid URL)
	if res.Status != StatusError {
		t.Errorf("Expected StatusError, got %v", res.Status)
	}

	// The message should contain [REDACTED]
	if !strings.Contains(res.Message, "[REDACTED]") {
		t.Errorf("Expected error message to contain [REDACTED], but got: %s", res.Message)
	}

	// The message should NOT contain the password "pass/word"
	if strings.Contains(res.Message, "pass/word") {
		t.Errorf("Security Leak! Error message contains raw password: %s", res.Message)
	}
}
