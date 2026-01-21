// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/doctor"
	"github.com/stretchr/testify/assert"
)

func ptr(s string) *string { return &s }

func TestDoctorRedaction_HTTP_URL_Leak(t *testing.T) {
	// configured URL triggers fallback regex (invalid scheme space) and has colon in password
	urlStr := "post gres://user:pass:word@localhost:5432/db"

	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: ptr("http-leak-test"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: ptr(urlStr),
					},
				},
			},
		},
	}

	results := doctor.RunChecks(context.Background(), cfg)
	assert.Len(t, results, 1)
	res := results[0]

	// The check should fail
	assert.Equal(t, doctor.StatusError, res.Status)

	// The message should contain the redacted DSN
	t.Logf("Doctor Message: %s", res.Message)

	// Ensure we DO NOT see the password "pass:word"
	assert.NotContains(t, res.Message, "pass:word", "Message should not contain the leaked password")

	// Ensure we DO see [REDACTED]
	assert.Contains(t, res.Message, "[REDACTED]", "Message should contain [REDACTED]")

	// Check for correct redaction structure if possible
	assert.Contains(t, res.Message, "user:[REDACTED]@localhost", "Message should contain correctly redacted string")
}
