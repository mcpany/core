// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/util"
)

func TestMiddlewareIPHandling_XFF(t *testing.T) {
	app := NewApplication()
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	// createAuthMiddleware(forcePrivateIPOnly bool, trustProxy bool)
	mw := app.createAuthMiddleware(false, true)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, ok := util.RemoteIPFromContext(r.Context())
		if !ok {
			t.Error("IP not found in context")
			return
		}
		// Expect clean IP
		expected := "::1"
		if ip != expected {
			t.Errorf("Expected IP %q, got %q", expected, ip)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "[::1]")
	req.RemoteAddr = "10.0.0.1:12345"

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr.Code)
	}
}
