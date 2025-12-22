// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestRBACMiddleware_EnforcePolicy(t *testing.T) {
	m := NewRBACMiddleware()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := m.EnforcePolicy(func(user *configv1.User) bool {
		return true
	})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotImplemented, rr.Code)
}
