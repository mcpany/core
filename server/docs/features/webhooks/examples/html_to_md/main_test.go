// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConvertHandler(t *testing.T) {
	reqData := WebhookRequest{
		Kind:     2,
		ToolName: "get_page_content",
		Result:   "<h1>Hello World</h1><p>This is a test.</p>",
	}
	body, _ := json.Marshal(reqData)
	req := httptest.NewRequest("POST", "/convert", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	convertHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var respData WebhookResponse
	json.NewDecoder(resp.Body).Decode(&respData)

	respMap, ok := respData.ReplacementObject.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected ReplacementObject to be a map, got %T", respData.ReplacementObject)
	}

	content, ok := respMap["content"].(string)
	if !ok {
		t.Errorf("Expected content field in replacement object")
	}

	if !strings.Contains(content, "# Hello World") {
		t.Errorf("Expected markdown title, got: %s", content)
	}
}
