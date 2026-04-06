// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookHandler_ListWebhooks(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/config.yaml", []byte(`
upstream_services:
  - name: test-service
    pre_call_hooks:
      - name: hook1
        webhook:
          url: http://localhost/hook1
    post_call_hooks:
      - name: hook2
        webhook:
          url: http://localhost/hook2
`), 0644)
	store := config.NewFileStore(fs, []string{"/config.yaml"})
	handler := NewWebhookHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/webhooks", nil)
	rr := httptest.NewRecorder()

	handler.ListWebhooks(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var webhooks []WebhookConfig
	err := json.Unmarshal(rr.Body.Bytes(), &webhooks)
	require.NoError(t, err)
	assert.Len(t, webhooks, 2)

	// pre_call
	assert.Equal(t, "hook1", webhooks[0].ID)
	assert.Equal(t, "http://localhost/hook1", webhooks[0].URL)
	assert.Contains(t, webhooks[0].Events, "pre_call")

	// post_call
	assert.Equal(t, "hook2", webhooks[1].ID)
	assert.Equal(t, "http://localhost/hook2", webhooks[1].URL)
	assert.Contains(t, webhooks[1].Events, "post_call")
}

func TestWebhookHandler_AddWebhook(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/config.yaml", []byte(`
upstream_services:
  - name: test-service
`), 0644)
	store := config.NewFileStore(fs, []string{"/config.yaml"})
	handler := NewWebhookHandler(store)

	payload := `{"url":"http://test.com/hook","events":["post_call"],"active":true}`
	req := httptest.NewRequest("POST", "/api/v1/webhooks", bytes.NewBufferString(payload))
	rr := httptest.NewRecorder()

	handler.AddWebhook(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var res map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.Equal(t, "success", res["status"])
	assert.NotEmpty(t, res["id"])

	// Verify it was saved
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)
	require.Len(t, cfg.UpstreamServices, 1)
	assert.Len(t, cfg.UpstreamServices[0].PostCallHooks, 1)
	assert.Equal(t, "http://test.com/hook", cfg.UpstreamServices[0].PostCallHooks[0].GetWebhook().GetUrl())
}

func TestWebhookHandler_DeleteWebhook(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/config.yaml", []byte(`
upstream_services:
  - name: test-service
    pre_call_hooks:
      - name: hook1
        webhook:
          url: http://localhost/hook1
`), 0644)
	store := config.NewFileStore(fs, []string{"/config.yaml"})
	handler := NewWebhookHandler(store)

	req := httptest.NewRequest("DELETE", "/api/v1/webhooks/hook1", nil)
	rr := httptest.NewRecorder()

	handler.DeleteWebhook(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	// Verify it was saved/deleted
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)
	require.Len(t, cfg.UpstreamServices, 1)
	assert.Len(t, cfg.UpstreamServices[0].PreCallHooks, 0)
}

func TestWebhookHandler_TestWebhook(t *testing.T) {
	// Start a mock test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/cloudevents+json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/config.yaml", []byte(`
upstream_services:
  - name: test-service
    pre_call_hooks:
      - name: hook1
        webhook:
          url: `+ts.URL+`
`), 0644)
	store := config.NewFileStore(fs, []string{"/config.yaml"})
	handler := NewWebhookHandler(store)

	req := httptest.NewRequest("POST", "/api/v1/webhooks/hook1/test", nil)
	rr := httptest.NewRecorder()

	handler.TestWebhook(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
}
