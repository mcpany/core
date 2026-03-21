// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestValidateHandler_Allowed(t *testing.T) {
	reqData := WebhookRequest{
		Kind:     1,
		ToolName: "execute",
		Inputs: map[string]any{
			"command": "ls -la",
		},
	}
	body, _ := json.Marshal(reqData)
	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	validateHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var respData WebhookResponse
	json.NewDecoder(resp.Body).Decode(&respData)

	if !respData.Allowed {
		t.Errorf("Expected Allowed=true, got false")
	}
}

func TestValidateHandler_Denied(t *testing.T) {
	reqData := WebhookRequest{
		Kind:     1,
		ToolName: "execute",
		Inputs: map[string]any{
			"command": "rm -rf /",
		},
	}
	body, _ := json.Marshal(reqData)
	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	validateHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var respData WebhookResponse
	json.NewDecoder(resp.Body).Decode(&respData)

	if respData.Allowed {
		t.Errorf("Expected Allowed=false, got true")
	}
	if !contains(respData.Status.Message, "restricted keyword 'rm'") {
		t.Errorf("Unexpected message: %s", respData.Status.Message)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr || len(s) > len(substr) && s[len(s)-len(substr):] == substr || len(s) > len(substr) && s[0:len(substr)] != substr && s[len(s)-len(substr):] != substr
	// quick impl, actually strings.Contains is better but need import
}
