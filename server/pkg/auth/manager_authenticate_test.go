// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestManager_Authenticate_GlobalAPIKey(t *testing.T) {
	am := auth.NewManager()
	am.SetAPIKey("secret-key")

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedError  bool
		expectedAPIKey string
	}{
		{
			name: "valid header",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/foo", nil)
				req.Header.Set("X-API-Key", "secret-key")
				return req
			},
			expectedError:  false,
			expectedAPIKey: "secret-key",
		},
		{
			name: "invalid header",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/foo", nil)
				req.Header.Set("X-API-Key", "wrong-key")
				return req
			},
			expectedError: true,
		},
		{
			name: "missing header",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/foo", nil)
				return req
			},
			expectedError: true,
		},
		{
			name: "query param (should fail)",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/foo?api_key=secret-key", nil)
				return req
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := tt.setupRequest()

			// We pass "unknown-service" so it falls back to global API key check
			newCtx, err := am.Authenticate(ctx, "unknown-service", req)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				key, ok := auth.APIKeyFromContext(newCtx)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedAPIKey, key)
			}
		})
	}
}

func TestManager_Authenticate_ServiceSpecific(t *testing.T) {
	am := auth.NewManager()

	// Register an authenticator for "service1"
	loc := configv1.APIKeyAuth_HEADER
	authConfig := &configv1.APIKeyAuth{
		ParamName: proto.String("X-Service-Key"),
		In:        &loc,
		VerificationValue: proto.String("service-secret"),
	}
	authenticator := auth.NewAPIKeyAuthenticator(authConfig)
	require.NotNil(t, authenticator)

	err := am.AddAuthenticator("service1", authenticator)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/foo", nil)
	req.Header.Set("X-Service-Key", "service-secret")

	ctx := context.Background()
	newCtx, err := am.Authenticate(ctx, "service1", req)
	assert.NoError(t, err)

	key, ok := auth.APIKeyFromContext(newCtx)
	assert.True(t, ok)
	assert.Equal(t, "service-secret", key)
}
