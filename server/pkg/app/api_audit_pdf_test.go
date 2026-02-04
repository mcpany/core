// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleAuditExport_PDF(t *testing.T) {
	// Setup
	mockStore := new(MockAuditStore)

	// Create AuditMiddleware
	auditConfig := &configv1.AuditConfig{}
	auditConfig.SetEnabled(true)
	am, _ := middleware.NewAuditMiddleware(auditConfig)
	am.SetStore(mockStore)

	app := &Application{
		standardMiddlewares: &middleware.StandardMiddlewares{
			Audit: am,
		},
	}

	entries := []audit.Entry{
		{
			Timestamp:  time.Now(),
			ToolName:   "test-tool",
			UserID:     "user1",
			ProfileID:  "default",
			Result:     "success",
			DurationMs: 100,
		},
	}

	mockStore.On("Read", mock.Anything, mock.Anything).Return(entries, nil)

	req := httptest.NewRequest("GET", "/audit/export?format=pdf", nil)
	w := httptest.NewRecorder()

	app.handleAuditExport(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/pdf", resp.Header.Get("Content-Type"))

	// Check file signature
	body := w.Body.Bytes()
	assert.True(t, strings.HasPrefix(string(body), "%PDF"), "Body should start with %PDF")
}
