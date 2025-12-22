// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestRBACMiddleware_RequireRole(t *testing.T) {
	m := NewRBACMiddleware()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name          string
		userRoles     []string
		requiredRole  string
		expectedCode  int
	}{
		{
			name:          "User has role",
			userRoles:     []string{"admin", "editor"},
			requiredRole:  "admin",
			expectedCode:  http.StatusOK,
		},
		{
			name:          "User missing role",
			userRoles:     []string{"editor"},
			requiredRole:  "admin",
			expectedCode:  http.StatusForbidden,
		},
		{
			name:          "No roles",
			userRoles:     nil,
			requiredRole:  "admin",
			expectedCode:  http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.userRoles != nil {
				req = req.WithContext(auth.ContextWithRoles(req.Context(), tt.userRoles))
			}

			rr := httptest.NewRecorder()
			middleware := m.RequireRole(tt.requiredRole)
			middleware(handler).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
		})
	}
}

func TestRBACMiddleware_EnforcePolicy(t *testing.T) {
	m := NewRBACMiddleware()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	// Policy always true, but EnforcePolicy returns NotImplemented
	middleware := m.EnforcePolicy(func(u *configv1.User) bool { return true })
	middleware(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotImplemented, rr.Code)
}

func TestRBACMiddleware_RequireAnyRole(t *testing.T) {
	m := NewRBACMiddleware()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name          string
		userRoles     []string
		requiredRoles []string
		expectedCode  int
	}{
		{
			name:          "User has one of the roles",
			userRoles:     []string{"editor"},
			requiredRoles: []string{"admin", "editor"},
			expectedCode:  http.StatusOK,
		},
		{
			name:          "User has none of the roles",
			userRoles:     []string{"viewer"},
			requiredRoles: []string{"admin", "editor"},
			expectedCode:  http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.userRoles != nil {
				req = req.WithContext(auth.ContextWithRoles(req.Context(), tt.userRoles))
			}

			rr := httptest.NewRecorder()
			middleware := m.RequireAnyRole(tt.requiredRoles...)
			middleware(handler).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
		})
	}
}
