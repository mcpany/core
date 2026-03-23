// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestManager_TestWebhook_Comprehensive(t *testing.T) {
	// Setup standard mock server to simulate non-2xx responses
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound) // Simulate 404
	}))
	defer ts.Close()

	tests := []struct {
		name          string
		webhookConfig *WebhookConfig
		testID        string
		expectError   bool
		expectedErr   string
		expectedStatus string
	}{
		{
			name:        "Webhook Not Found",
			webhookConfig: nil, // Do not add
			testID:      "non-existent",
			expectError: true,
			expectedErr: "webhook not found",
			expectedStatus: "",
		},
		{
			name: "Invalid URL Format",
			webhookConfig: &WebhookConfig{
				URL:    "http://192.168.0.%31",
				Events: []string{"test"},
				Active: true,
			},
			testID:      "", // Assigned dynamically
			expectError: true,
			expectedErr: "invalid URL escape", // Go's built-in url parse error string fragment
			expectedStatus: "",
		},
		{
			name: "Dial Failure Invalid Scheme",
			webhookConfig: &WebhookConfig{
				URL:    "invalid-scheme://foo",
				Events: []string{"test"},
				Active: true,
			},
			testID:      "", // Assigned dynamically
			expectError: true,
			expectedErr: "unsupported protocol scheme",
			expectedStatus: "failure",
		},
		{
			name: "Non-2xx Status Code",
			webhookConfig: &WebhookConfig{
				URL:    ts.URL,
				Events: []string{"test"},
				Active: true,
			},
			testID:      "", // Assigned dynamically
			expectError: true,
			expectedErr: "status code: 404",
			expectedStatus: "failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			idToTest := tt.testID

			// Setup Webhook
			if tt.webhookConfig != nil {
				m.AddWebhook(tt.webhookConfig)

				// Find dynamic ID
				for _, hw := range m.ListWebhooks() {
					if hw.URL == tt.webhookConfig.URL {
						idToTest = hw.ID
						break
					}
				}
			}

			err := m.TestWebhook(context.Background(), idToTest)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error containing '%s', got nil", tt.expectedErr)
				}
				if tt.expectedErr != "" && !containsStr(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error to contain '%s', but got: %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
			}

			// Verify status mutations
			if tt.expectedStatus != "" {
				updated, ok := m.GetWebhook(idToTest)
				if !ok {
					t.Fatalf("Expected to find webhook '%s' to check status", idToTest)
				}
				if updated.Status != tt.expectedStatus {
					t.Errorf("Expected status to be '%s', got '%s'", tt.expectedStatus, updated.Status)
				}
			}
		})
	}
}

// Helper to check substring since strings.Contains might not be imported
func containsStr(s, substr string) bool {
	// Basic implementation or use strings.Contains
	// Doing it simply to avoid modifying imports if possible
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
