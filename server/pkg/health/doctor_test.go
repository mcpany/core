// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockTransport struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestDoctorHandler(t *testing.T) {
	doctor := NewDoctor()

	// Mock the HTTP client to avoid external calls
	doctor.httpClient = &http.Client{
		Transport: &mockTransport{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString("OK")),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	req, _ := http.NewRequest("GET", "/doctor", nil)
	w := httptest.NewRecorder()
	doctor.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var report DoctorReport
	err := json.Unmarshal(w.Body.Bytes(), &report)
	assert.NoError(t, err)
	assert.NotEmpty(t, report.Status)
	assert.Contains(t, report.Checks, "internet")
	assert.Equal(t, "ok", report.Checks["internet"].Status)
}

func TestDoctor_AddCheck(t *testing.T) {
	doctor := NewDoctor()
	// Mock client
	doctor.httpClient = &http.Client{
		Transport: &mockTransport{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString("OK")),
				}, nil
			},
		},
	}

	doctor.AddCheck("custom", func(ctx context.Context) CheckResult {
		return CheckResult{
			Status:  "ok",
			Message: "All good",
		}
	})

	doctor.AddCheck("custom_fail", func(ctx context.Context) CheckResult {
		return CheckResult{
			Status:  "error",
			Message: "Not good",
		}
	})

	req, _ := http.NewRequest("GET", "/doctor", nil)
	w := httptest.NewRecorder()
	doctor.Handler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var report DoctorReport
	err := json.Unmarshal(w.Body.Bytes(), &report)
	assert.NoError(t, err)

	assert.Equal(t, "degraded", report.Status)
	assert.Contains(t, report.Checks, "custom")
	assert.Equal(t, "ok", report.Checks["custom"].Status)
	assert.Equal(t, "All good", report.Checks["custom"].Message)

	assert.Contains(t, report.Checks, "custom_fail")
	assert.Equal(t, "error", report.Checks["custom_fail"].Status)
}

func TestHistoryHandler(t *testing.T) {
	doctor := NewDoctor()

	// Seed history
	AddHealthStatus("test-service", "ok")
	AddHealthStatus("test-service", "error")

	req, _ := http.NewRequest("GET", "/doctor/history", nil)
	w := httptest.NewRecorder()
	doctor.HistoryHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var history map[string][]HistoryPoint
	err := json.Unmarshal(w.Body.Bytes(), &history)
	assert.NoError(t, err)

	assert.Contains(t, history, "test-service")
	assert.GreaterOrEqual(t, len(history["test-service"]), 2)
	assert.Equal(t, "ok", history["test-service"][0].Status)
	assert.Equal(t, "error", history["test-service"][1].Status)
}
