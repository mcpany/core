// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/stretchr/testify/assert"
)

func TestHandleInitiateOAuth_Detailed(t *testing.T) {
	app := &Application{
		AuthManager: auth.NewManager(),
	}

	tests := []struct {
		name         string
		method       string
		body         string
		userID       string
		expectedCode int
		expectedBody []string
	}{
		{
			name:         "InvalidMethod",
			method:       http.MethodGet,
			body:         ``,
			userID:       "user1",
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: []string{"method not allowed"},
		},
		{
			name:         "InvalidJSON",
			method:       http.MethodPost,
			body:         `invalid`,
			userID:       "user1",
			expectedCode: http.StatusBadRequest,
			expectedBody: []string{"invalid request body"},
		},
		{
			name:         "MissingFields",
			method:       http.MethodPost,
			body:         `{"service_id":""}`,
			userID:       "user1",
			expectedCode: http.StatusBadRequest,
			expectedBody: []string{"service_id (or credential_id) and redirect_url are required"},
		},
		{
			name:         "Unauthorized",
			method:       http.MethodPost,
			body:         `{"service_id":"github","redirect_url":"http://localhost/callback"}`,
			userID:       "",
			expectedCode: http.StatusUnauthorized,
			expectedBody: []string{"unauthorized"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, "/auth/oauth/initiate", bytes.NewBufferString(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, "/auth/oauth/initiate", nil)
			}

			if tt.userID != "" {
				ctx := auth.ContextWithUser(req.Context(), tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			app.handleInitiateOAuth(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			for _, exp := range tt.expectedBody {
				assert.Contains(t, w.Body.String(), exp)
			}
		})
	}
}
