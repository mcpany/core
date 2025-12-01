// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthenticationMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		headerKey      string
		expectedStatus int
	}{
		{
			name:           "No API Key Required",
			apiKey:         "",
			headerKey:      "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Correct API Key",
			apiKey:         "test-key",
			headerKey:      "test-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Incorrect API Key",
			apiKey:         "test-key",
			headerKey:      "wrong-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing API Key",
			apiKey:         "test-key",
			headerKey:      "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			if tt.headerKey != "" {
				req.Header.Set("X-API-Key", tt.headerKey)
			}

			rr := httptest.NewRecorder()
			handler := AuthenticationMiddleware(tt.apiKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}
