// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package doctor

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

func TestCheckService_Redaction_DSN_OverRedaction_Bug(t *testing.T) {
	// This test reproduces a specific bug where RedactDSN would over-redact
	// URLs that fail parsing (e.g. invalid port) and contain parameters with @.
	// Input: postgres://user:password@host:abc/db?email=foo@bar.com
	// Incorrect Output: postgres://user:[REDACTED]@bar.com
	// Correct Output: postgres://user:[REDACTED]@host:abc/db?email=foo@bar.com

	// We use HttpService because checkURL uses http.NewRequest which uses url.Parse.
	// We want url.Parse to fail.
	urlStr := "postgres://user:password@host:abc/db?email=foo@bar.com"
	service := &configv1.UpstreamServiceConfig{
		Name: stringPtr("test-http-redact-bug"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: stringPtr(urlStr),
			},
		},
	}

	res := CheckService(context.Background(), service)

	t.Logf("Result Message: %s", res.Message)

	// It should fail because of invalid port
	if res.Status != StatusError {
		t.Fatalf("Expected StatusError, got %s", res.Status)
	}

	// Verify Redaction
	if strings.Contains(res.Message, "password") {
		t.Errorf("Password leaked in error message: %s", res.Message)
	}

	if !strings.Contains(res.Message, "[REDACTED]") {
		t.Errorf("Expected [REDACTED] placeholder in error message: %s", res.Message)
	}

	// Verify Over-redaction (Host swallowing)
	if !strings.Contains(res.Message, "host:abc") {
		t.Errorf("Host was swallowed (over-redaction): %s", res.Message)
	}

	if !strings.Contains(res.Message, "email=foo@bar.com") {
		t.Errorf("Query params were swallowed (over-redaction): %s", res.Message)
	}
}
