// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestAPISecurity_RBAC(t *testing.T) {
	store := memory.NewStore()
	app := NewApplication()
	app.Storage = store

	tests := []struct {
		name          string
		method        string
		handler       func(storage.Storage) http.HandlerFunc
		directHandler http.HandlerFunc
		roles         []string
		wantStatus    int
	}{
		{
			name:       "Settings POST - No Roles",
			method:     http.MethodPost,
			handler:    app.handleSettings,
			roles:      nil,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Settings POST - User Role",
			method:     http.MethodPost,
			handler:    app.handleSettings,
			roles:      []string{"user"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Settings POST - Admin Role",
			method:     http.MethodPost,
			handler:    app.handleSettings,
			roles:      []string{"admin"},
			wantStatus: http.StatusOK,
		},
		{
			name:          "Secret Create - No Roles",
			method:        http.MethodPost,
			directHandler: app.createSecretHandler,
			roles:         nil,
			wantStatus:    http.StatusForbidden,
		},
		{
			name:          "Secret Create - Admin Role",
			method:        http.MethodPost,
			directHandler: app.createSecretHandler,
			roles:         []string{"admin"},
			wantStatus:    http.StatusCreated,
		},
		{
			name:          "Credential Create - No Roles",
			method:        http.MethodPost,
			directHandler: app.createCredentialHandler,
			roles:         nil,
			wantStatus:    http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h http.HandlerFunc
			if tt.handler != nil {
				h = tt.handler(store)
			} else {
				h = tt.directHandler
			}

			body := "{}"
			if tt.wantStatus == http.StatusCreated || tt.wantStatus == http.StatusOK {
				if tt.name == "Secret Create - Admin Role" {
					secret := &configv1.Secret{}
					secret.SetName("test")
					secret.SetKey("key")
					secret.SetValue("val")
					b, _ := protojson.Marshal(secret)
					body = string(b)
				}
				if tt.name == "Credential Create - Admin Role" {
					// createCredentialHandler checks Name or ID
					cred := &configv1.Credential{}
					cred.SetName("test")
					// We don't need to set Authentication to pass basic validation if optional,
					// but it's good practice.
					// However, proto unmarshal won't fail if fields are missing unless validation requires it.
					// createCredentialHandler only checks ID/Name.

					// createCredentialHandler validates request body first.
					// If we pass empty body to Admin role, it might fail 400 Bad Request.
					// But our expectation is 201 Created for success case.
					// For failure case (Forbidden), empty body is fine if RBAC check is before unmarshal.
					// In my implementation, RBAC check IS before unmarshal.
					b, _ := protojson.Marshal(cred)
					body = string(b)
				}
			}

			req := httptest.NewRequest(tt.method, "/test", bytes.NewBufferString(body))
			ctx := req.Context()
			if len(tt.roles) > 0 {
				ctx = auth.ContextWithRoles(ctx, tt.roles)
			}
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			h(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", w.Code, tt.wantStatus)
			}
		})
	}
}
