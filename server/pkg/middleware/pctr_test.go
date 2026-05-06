// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPCTRMiddleware_SubsetEnforcement(t *testing.T) {
	pctr := NewPCTRMiddleware()

	// Setup initial state
	pctr.RegisterToken("token_123", []string{"pairing", "read:files", "write:files"})
	pctr.RegisterToken("token_admin", []string{"admin"})

	handler := pctr.APIHandler()

	t.Run("successful rotation with exact subset", func(t *testing.T) {
		reqBody := RotateRequest{
			OldToken:        "token_123",
			RequestedScopes: []string{"read:files"},
		}
		b, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/rotate", bytes.NewReader(b))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %v", rec.Code)
		}

		var resp RotateResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}

		if resp.NewToken != "token_123_rotated" {
			t.Errorf("unexpected token format: %v", resp.NewToken)
		}

		// Verify old token is gone and new is present
		if _, ok := pctr.activeTokens["token_123"]; ok {
			t.Error("old token was not invalidated")
		}
		if _, ok := pctr.activeTokens["token_123_rotated"]; !ok {
			t.Error("new token was not stored")
		}
	})

	t.Run("rejected rotation with escalated scopes", func(t *testing.T) {
		// New token from previous test is now "token_123_rotated" with scopes ["read:files"]
		reqBody := RotateRequest{
			OldToken:        "token_123_rotated",
			RequestedScopes: []string{"read:files", "write:files"}, // Escalation attempt!
		}
		b, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/rotate", bytes.NewReader(b))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden, got %v", rec.Code)
		}
	})

	t.Run("rejected rotation with unknown token", func(t *testing.T) {
		reqBody := RotateRequest{
			OldToken:        "unknown_token",
			RequestedScopes: []string{"read:files"},
		}
		b, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/auth/rotate", bytes.NewReader(b))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got %v", rec.Code)
		}
	})
}
