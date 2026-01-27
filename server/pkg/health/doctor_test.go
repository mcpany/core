// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoctorHandler(t *testing.T) {
	doctor := NewDoctor()

	req, _ := http.NewRequest("GET", "/doctor", nil)
	w := httptest.NewRecorder()
	doctor.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var report DoctorReport
	err := json.Unmarshal(w.Body.Bytes(), &report)
	assert.NoError(t, err)
	assert.NotEmpty(t, report.Status)
	assert.Contains(t, report.Checks, "internet")
}
