// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebhooksAPI(t *testing.T) {
	app := NewApplication()

	// Test List (Empty)
	req, _ := http.NewRequest("GET", "/webhooks", nil)
	rr := httptest.NewRecorder()
	app.listWebhooksHandler(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	// Default slice is empty, which marshals to []
	assert.JSONEq(t, "[]", rr.Body.String())

	// Test Create
	// but standard json often works for simple structs if fields are exported.
	// However, generated code uses special fields.
	// Let's use standard json if the struct has json tags (it usually does).
	// Or use protojson manually?
	// Handler uses json.NewDecoder(r.Body).Decode(&req) which uses standard json.
	// So we should send standard json.
	// But generated proto structs might be tricky.
	// Let's construct a map to be safe.
	input := map[string]interface{}{
		"name":     "TestHook",
		"url_path": "/test", // Try snake_case
	}
	body, _ := json.Marshal(input)
	req, _ = http.NewRequest("POST", "/webhooks", bytes.NewBuffer(body))
	rr = httptest.NewRecorder()
	app.createWebhookHandler(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Test List (Populated)
	req, _ = http.NewRequest("GET", "/webhooks", nil)
	rr = httptest.NewRecorder()
	app.listWebhooksHandler(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Decode response
	var hooks []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &hooks)
	assert.NoError(t, err)
	assert.Len(t, hooks, 1)
	assert.Equal(t, "TestHook", hooks[0]["name"])
}
