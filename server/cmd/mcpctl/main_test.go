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

	assert.Contains(t, b.String(), "Checking Configuration")
	assert.Contains(t, b.String(), "Checking Server Connectivity")
	assert.Contains(t, b.String(), "FAILED")
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

	err = cmd.ExecuteContext(context.Background())
	assert.NoError(t, err)
	assert.Contains(t, b.String(), "Checking Server Connectivity")
	assert.Contains(t, b.String(), "OK")
	assert.Contains(t, b.String(), "Checking System Health")
	assert.Contains(t, b.String(), "internet: OK")
}

func TestDoctorCmd_AddressParsing(t *testing.T) {
	tests := []struct {
		name string
		arg  string
	}{
		{"Port only", "50050"},
		{"Colon port", ":50050"},
		{"Localhost port", "127.0.0.1:50050"},
		{"IP port", "127.0.0.1:50050"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newRootCmd()
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			// We expect connection failure, but we want to make sure it parses correctly and attempts connection
			cmd.SetArgs([]string{"doctor", "--mcp-listen-address", tt.arg})
			_ = cmd.Execute()
			// It should fail connectivity but not crash
			assert.Contains(t, b.String(), "Checking Server Connectivity")
		})
	}
}

func TestDoctorCmd_ServerErrors(t *testing.T) {
	// 1. Server returns 500 on health
	t.Run("Health 500", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()
		_, port, _ := net.SplitHostPort(strings.TrimPrefix(ts.URL, "http://"))

		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"doctor", "--mcp-listen-address", ":" + port})
		_ = cmd.Execute()
		assert.Contains(t, b.String(), "Server returned status: 500 Internal Server Error")
	})

	// 2. Server returns 500 on doctor
	t.Run("Doctor 500", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}))
		defer ts.Close()
		_, port, _ := net.SplitHostPort(strings.TrimPrefix(ts.URL, "http://"))

		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"doctor", "--mcp-listen-address", ":" + port})
		_ = cmd.Execute()
		assert.Contains(t, b.String(), "Doctor endpoint returned status: 500")
	})

	// 3. Doctor returns invalid JSON
	t.Run("Doctor Invalid JSON", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json"))
			}
		}))
		defer ts.Close()
		_, port, _ := net.SplitHostPort(strings.TrimPrefix(ts.URL, "http://"))

		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"doctor", "--mcp-listen-address", ":" + port})
		_ = cmd.Execute()
		assert.Contains(t, b.String(), "Failed to decode doctor report")
	})

    // 4. Doctor returns degraded status
	t.Run("Doctor Degraded", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
			} else {
                w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
                w.Write([]byte(`{"status":"degraded","checks":{"db":{"status":"failed","message":"connection lost"}}}`))
			}
		}))
		defer ts.Close()
		_, port, _ := net.SplitHostPort(strings.TrimPrefix(ts.URL, "http://"))

		cmd := newRootCmd()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs([]string{"doctor", "--mcp-listen-address", ":" + port})
		_ = cmd.Execute()
		assert.Contains(t, b.String(), "DEGRADED")
		assert.Contains(t, b.String(), "db: FAIL")
        assert.Contains(t, b.String(), "connection lost")
	})
}
