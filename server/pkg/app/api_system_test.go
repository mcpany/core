// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleSystemDiagnostics(t *testing.T) {
	app := &Application{}

	req, err := http.NewRequest("GET", "/system/diagnostics", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.handleSystemDiagnostics)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp SystemDiagnosticsResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, runtime.GOOS, resp.OS)
	assert.Equal(t, runtime.GOARCH, resp.Arch)
	// Docker check might be true or false depending on where this test runs,
	// so we just ensure the field exists and boolean (which it is by type).
	// We check if EnvVars is not nil.
	assert.NotNil(t, resp.EnvVars)
	// Check that we have at least one of the expected keys
	_, exists := resp.EnvVars["HTTP_PROXY"]
	assert.True(t, exists)
}
