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

func TestHandleInitiateOAuth(t *testing.T) {
	// Setup
	store := memory.NewStore()
	am := auth.NewManager()
	am.SetStorage(store)
	app := &Application{
		AuthManager: am,
	}

	// Create valid service and credential with OAuth config
	ctx := context.Background()
	svcID := "github"
	credID := "cred1"

	// Seed Service with OAuth
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String(svcID),
		UpstreamAuth: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_Oauth2{
				Oauth2: &configv1.OAuth2Auth{
					ClientId:         &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "client-id"}},
					ClientSecret:     &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "client-secret"}},
					AuthorizationUrl: proto.String("https://github.com/login/oauth/authorize"),
					TokenUrl:         proto.String("https://github.com/login/oauth/access_token"),
					Scopes:           proto.String("read:user"),
				},
			},
		},
	}
	err := store.SaveService(ctx, svc)
	require.NoError(t, err)

	// Seed Credential with OAuth
	cred := &configv1.Credential{
		Id: proto.String(credID),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_Oauth2{
				Oauth2: &configv1.OAuth2Auth{
					ClientId:         &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "cred-client-id"}},
					ClientSecret:     &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "cred-client-secret"}},
					AuthorizationUrl: proto.String("https://example.com/auth"),
					TokenUrl:         proto.String("https://example.com/token"),
				},
			},
		},
	}
	err = store.SaveCredential(ctx, cred)
	require.NoError(t, err)

	tests := []struct {
		name           string
		method         string
		body           map[string]string
		userInContext  string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "MethodNotAllowed",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "InvalidBody",
			method:         http.MethodPost,
			body:           nil, // Will fail decode if we pass nil reader, but we need to pass something that fails decode? Or just close early.
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "MissingParams",
			method: http.MethodPost,
			body: map[string]string{
				"service_id": "foo",
				// redirect_url missing
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Unauthorized",
			method: http.MethodPost,
			body: map[string]string{
				"service_id":   svcID,
				"redirect_url": "http://localhost:8080/cb",
			},
			userInContext:  "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "Success_Service",
			method: http.MethodPost,
			body: map[string]string{
				"service_id":   svcID,
				"redirect_url": "http://localhost:8080/cb",
			},
			userInContext:  "user1",
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Success_Credential",
			method: http.MethodPost,
			body: map[string]string{
				"credential_id": credID,
				"redirect_url":  "http://localhost:8080/cb",
			},
			userInContext:  "user1",
			expectedStatus: http.StatusOK,
		},
		{
			name:   "ServiceNotFound",
			method: http.MethodPost,
			body: map[string]string{
				"service_id":   "unknown",
				"redirect_url": "http://localhost:8080/cb",
			},
			userInContext:  "user1",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var bodyBytes []byte
			if tc.name == "InvalidBody" {
				bodyBytes = []byte("invalid-json")
			} else if tc.body != nil {
				bodyBytes, _ = json.Marshal(tc.body)
			} else {
				bodyBytes = []byte("{}")
			}

			req := httptest.NewRequest(tc.method, "/auth/oauth/initiate", bytes.NewReader(bodyBytes))
			if tc.userInContext != "" {
				req = req.WithContext(auth.ContextWithUser(req.Context(), tc.userInContext))
			}

			w := httptest.NewRecorder()
			app.handleInitiateOAuth(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedStatus == http.StatusOK {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp["authorization_url"])
				assert.NotEmpty(t, resp["state"])
			} else if tc.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tc.expectedBody)
			}
		})
	}
}

func TestHandleOAuthCallback(t *testing.T) {
	// Setup
	store := memory.NewStore()
	am := auth.NewManager()
	am.SetStorage(store)
	app := &Application{
		AuthManager: am,
	}

	// NOTE: mocking Execute is hard without dependency injection on Manager for OAuth config/exchange.
	// Manager uses "golang.org/x/oauth2" directly.
	// However, we can test validation logic and error paths before Exchange.
	// Real Exchange will fail because we can't easily mock the token endpoint without
	// creating a real server or changing AuthManager to use a swappable factory.
	//
	// Given the constraints, we will test:
	// 1. Validation errors (method, body, params, auth)
	// 2. Storage lookup errors (service/cred not found)
	// 3. Exchange failure (which confirms we reached that point)

	tests := []struct {
		name           string
		method         string
		body           map[string]string
		userInContext  string
		expectedStatus int
		errorContains  string
	}{
		{
			name:           "MethodNotAllowed",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "InvalidBody",
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest, // Invalid URL encoded body for JSON decoder? or just verify logic
		},
		{
			name:   "MissingParams",
			method: http.MethodPost,
			body: map[string]string{
				"service_id": "foo",
				// missing code/redirect
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Unauthorized",
			method: http.MethodPost,
			body: map[string]string{
				"service_id":   "foo",
				"code":         "123",
				"redirect_url": "cb",
			},
			userInContext:  "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "ServiceNotFound",
			method: http.MethodPost,
			body: map[string]string{
				"service_id":   "unknown",
				"code":         "123",
				"redirect_url": "cb",
			},
			userInContext:  "user1",
			expectedStatus: http.StatusInternalServerError,
			errorContains:  "service unknown not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var bodyBytes []byte
			if tc.name == "InvalidBody" {
				bodyBytes = []byte("not-json")
			} else if tc.body != nil {
				bodyBytes, _ = json.Marshal(tc.body)
			} else {
				bodyBytes = []byte("{}")
			}

			req := httptest.NewRequest(tc.method, "/auth/oauth/callback", bytes.NewReader(bodyBytes))
			if tc.userInContext != "" {
				req = req.WithContext(auth.ContextWithUser(req.Context(), tc.userInContext))
			}

			w := httptest.NewRecorder()
			app.handleOAuthCallback(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.errorContains != "" {
				assert.Contains(t, w.Body.String(), tc.errorContains)
			}
		})
	}
}
