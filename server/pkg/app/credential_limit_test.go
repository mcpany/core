// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCredentialHandler_LargeBody(t *testing.T) {
	app := setupTestApp()

	// Create a large body (> 1MB)
	largeBody := make([]byte, 2*1024*1024) // 2MB
	// Fill with some dummy JSON-like data so it's not just zeros (though zeros are also bad JSON)
	// But ReadAll doesn't care about content.
	largeBody[0] = '{'
	largeBody[len(largeBody)-1] = '}'

	req := httptest.NewRequest(http.MethodPost, "/api/v1/credentials", bytes.NewReader(largeBody))
	rr := httptest.NewRecorder()

	// Direct call to handler
	app.createCredentialHandler(rr, req)

	// Desired behavior: 413 Payload Too Large.
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Logf("Got status %d, body: %s", rr.Code, rr.Body.String())
		t.Errorf("Expected 413 Payload Too Large, got %d", rr.Code)
	} else {
		assert.Equal(t, http.StatusRequestEntityTooLarge, rr.Code)
	}
}

func TestUpdateCredentialHandler_LargeBody(t *testing.T) {
	app := setupTestApp()

	largeBody := make([]byte, 2*1024*1024)
	largeBody[0] = '{'
	largeBody[len(largeBody)-1] = '}'

	req := httptest.NewRequest(http.MethodPut, "/api/v1/credentials/123", bytes.NewReader(largeBody))
	rr := httptest.NewRecorder()

	app.updateCredentialHandler(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected 413 Payload Too Large, got %d", rr.Code)
	}
}
