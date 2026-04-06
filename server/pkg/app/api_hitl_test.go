// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/bus"
)

func TestHITLApprovalMFA(t *testing.T) {
	app := &Application{
		busProvider: bus.NewProvider(), // Provides memory bus for tests
	}
	mux := http.NewServeMux()
	app.mountHITL(mux)

	time.Sleep(10 * time.Millisecond)

	tests := []struct {
		name       string
		reqBody    map[string]interface{}
		statusCode int
	}{
		{
			name: "Valid MFA Format",
			reqBody: map[string]interface{}{
				"action":  "approved",
				"mfaCode": "123456",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "Invalid MFA Length",
			reqBody: map[string]interface{}{
				"action":  "approved",
				"mfaCode": "12345",
			},
			statusCode: http.StatusUnauthorized,
		},
		{
			name: "Invalid MFA Characters",
			reqBody: map[string]interface{}{
				"action":  "approved",
				"mfaCode": "abcdef",
			},
			statusCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			globalHITLState.mu.Lock()
			globalHITLState.approvals["test-mfa"] = hitlApprovalRequest{
				ExecutionID: "test-mfa",
				ToolName:    "db.drop",
				RequireMFA:  true,
			}
			globalHITLState.mu.Unlock()

			body, _ := json.Marshal(tc.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/hitl/approvals/test-mfa", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			res := w.Result()
			if res.StatusCode != tc.statusCode {
				t.Errorf("Expected status %d, got %v", tc.statusCode, res.StatusCode)
			}
		})
	}
}
