// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDoctorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	doctor := NewDoctor()
	r := gin.New()
	r.GET("/doctor", doctor.Handler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/doctor", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var report DoctorReport
	err := json.Unmarshal(w.Body.Bytes(), &report)
	assert.NoError(t, err)

	assert.Equal(t, "healthy", report.Status) // Assuming internet is available in CI, otherwise might fail
	assert.Contains(t, report.Checks, "internet")
	assert.WithinDuration(t, time.Now(), report.Timestamp, 5*time.Second)
}
