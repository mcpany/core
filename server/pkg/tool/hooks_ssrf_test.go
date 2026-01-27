// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestWebhookClient_SSRFProtection(t *testing.T) {
	// 1. Start a local server (target for SSRF)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 2. Configure webhook to point to it
	config := configv1.WebhookConfig_builder{
		Url: server.URL,
	}.Build()

	// 3. Ensure env vars are cleared so we test default secure behavior
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "")
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "")
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "")

	// 4. Create client
	client := NewWebhookClient(config)

	// 5. Call
	_, err := client.Call(context.Background(), "test.event", map[string]string{"foo": "bar"})

	// 6. Assert failure
	assert.Error(t, err)
	if err != nil {
		// cloudevents client wraps the error
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
	}
}
