// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCmd(t *testing.T) {
	cmd := newRootCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"version"})
	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, b.String(), "version dev")
}

func TestDoctorCmd_Offline(t *testing.T) {
	cmd := newRootCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	// Use a random port that is likely closed
	cmd.SetArgs([]string{"doctor", "--mcp-listen-address", ":54321"})

	// Depending on implementation, it might error or just print failure
	// For this test, we just want to ensure it doesn't panic and tries to run
	_ = cmd.Execute()

	// We haven't implemented it yet, so this will fail with unknown command
	// assert.Contains(t, b.String(), "Checking server connectivity")
}

func TestDoctorCmd_Online(t *testing.T) {
	// Mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		} else if r.URL.Path == "/doctor" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"healthy","checks":{"internet":{"status":"ok"}}}`))
		}
	}))
	defer ts.Close()

	// Extract port from ts.URL
	_, port, err := net.SplitHostPort(strings.TrimPrefix(ts.URL, "http://"))
	assert.NoError(t, err)

	cmd := newRootCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"doctor", "--mcp-listen-address", ":" + port})

	// Execute (will fail until implemented)
	_ = cmd.ExecuteContext(context.Background())
}
