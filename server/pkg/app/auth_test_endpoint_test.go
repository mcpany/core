// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleAuthTest(t *testing.T) {
	// Setup
	store := memory.NewStore()
	am := auth.NewManager()
	am.SetStorage(store)
	app := &Application{
		AuthManager: am,
		Storage:     store,
	}

	// Create a test server for HTTP testing
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer mytoken" {
			w.WriteHeader(http.StatusOK)
		} else if r.Header.Get("X-API-Key") == "mykey" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer ts.Close()

	// Seed Credential
	credID := "cred-http"
	loc := configv1.APIKeyAuth_HEADER
	cred := &configv1.Credential{
		Id: proto.String(credID),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_ApiKey{
				ApiKey: &configv1.APIKeyAuth{
					In:                &loc,
					ParamName:         proto.String("X-API-Key"),
					Value:             &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "mykey"}},
					VerificationValue: proto.String("mykey"), // Used for server-side verification, simpler to just match
				},
			},
		},
	}
	err := store.SaveCredential(context.Background(), cred)
	require.NoError(t, err)

	tests := []struct {
		name           string
		req            AuthTestRequest
		expectedStatus int
		expectSuccess  bool
		expectMessage  string
	}{
		{
			name: "HTTP Success with Credential",
			req: AuthTestRequest{
				CredentialID: credID,
				ServiceType:  "HTTP",
				ServiceConfig: map[string]any{
					"http_service": map[string]any{
						"address": ts.URL,
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name: "HTTP Failure (401) without Credential",
			req: AuthTestRequest{
				CredentialID: "none",
				ServiceType:  "HTTP",
				ServiceConfig: map[string]any{
					"http_service": map[string]any{
						"address": ts.URL,
					},
				},
			},
			// The handler catches error and returns success=false but HTTP 200 OK?
			// verifyAuthResponse sends 200 OK with success: false usually.
			// Let's check handleAuthTest logic.
			// err != nil -> writeAuthResponse(w, false, err.Error())
			// writeAuthResponse sets 200 OK mostly?
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
			expectMessage:  "server returned error status: 401",
		},
		{
			name: "Command Success",
			req: AuthTestRequest{
				ServiceType: "CMD",
				ServiceConfig: map[string]any{
					"command_line_service": map[string]any{
						"command": "ls", // Should exist on linux
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name: "Command Failure (Missing)",
			req: AuthTestRequest{
				ServiceType: "CMD",
				ServiceConfig: map[string]any{
					"command_line_service": map[string]any{
						"command": "nonexistent_command_12345",
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
			expectMessage:  "not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.req)
			req := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Register handler manually or call directly
			// It returns http.HandlerFunc
			handler := app.handleAuthTest()
			handler(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var resp AuthTestResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectSuccess, resp.Success)
			if tc.expectMessage != "" {
				assert.Contains(t, resp.Message, tc.expectMessage)
			}
		})
	}
}
