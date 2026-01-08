// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDoctorHandler(t *testing.T) {
	doctor := NewDoctor()
	handler := doctor.Handler()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/doctor", nil)
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var report DoctorReport
	err := json.Unmarshal(w.Body.Bytes(), &report)
	assert.NoError(t, err)

	// In some constrained environments (like sandbox), internet might not be available.
	// We check that status is either healthy or degraded, but the structure is correct.
	assert.Contains(t, []string{"healthy", "degraded"}, report.Status)
	assert.Contains(t, report.Checks, "internet")
	assert.WithinDuration(t, time.Now(), report.Timestamp, 5*time.Second)
}
