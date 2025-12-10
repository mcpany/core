// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestHttpChecker_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}))
	defer server.Close()

	cfg := &config.HttpHealthCheck{
		Url:        server.URL,
		Interval:   durationpb.New(100 * time.Millisecond),
		ExpectedCode: http.StatusOK,
	}

	mockUpstream := &upstream.MockUpstream{}
	checker, err := NewHttpChecker("test-http-success", cfg, mockUpstream)
	require.NoError(t, err)

	err = checker.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestHttpChecker_WrongStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &config.HttpHealthCheck{
		Url:        server.URL,
		Interval:   durationpb.New(100 * time.Millisecond),
		ExpectedCode: http.StatusOK,
	}

	mockUpstream := &upstream.MockUpstream{}
	checker, err := NewHttpChecker("test-http-wrong-status", cfg, mockUpstream)
	require.NoError(t, err)

	err = checker.HealthCheck(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code")
}

func TestHttpChecker_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Sleep longer than the timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.HttpHealthCheck{
		Url:      server.URL,
		Interval: durationpb.New(100 * time.Millisecond),
		Timeout:  durationpb.New(50 * time.Millisecond),
	}

	mockUpstream := &upstream.MockUpstream{}
	checker, err := NewHttpChecker("test-http-timeout", cfg, mockUpstream)
	require.NoError(t, err)

	err = checker.HealthCheck(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestHttpChecker_ResponseBodyCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Service is running")
	}))
	defer server.Close()

	mockUpstream := &upstream.MockUpstream{}

	// Test case 1: Successful body check
	cfgSuccess := &config.HttpHealthCheck{
		Url:                          server.URL,
		ExpectedResponseBodyContains: "running",
	}
	checkerSuccess, err := NewHttpChecker("test-body-success", cfgSuccess, mockUpstream)
	require.NoError(t, err)
	err = checkerSuccess.HealthCheck(context.Background())
	assert.NoError(t, err)

	// Test case 2: Failed body check
	cfgFail := &config.HttpHealthCheck{
		Url:                          server.URL,
		ExpectedResponseBodyContains: "stopped",
	}
	checkerFail, err := NewHttpChecker("test-body-fail", cfgFail, mockUpstream)
	require.NoError(t, err)
	err = checkerFail.HealthCheck(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not contain expected string")
}

func TestNewHttpChecker_Validation(t *testing.T) {
	mockUpstream := &upstream.MockUpstream{}

	_, err := NewHttpChecker("test-nil-config", nil, mockUpstream)
	assert.Error(t, err, "Should return error for nil config")

	cfgNoURL := &config.HttpHealthCheck{}
	_, err = NewHttpChecker("test-no-url", cfgNoURL, mockUpstream)
	assert.Error(t, err, "Should return error for empty URL")
}
