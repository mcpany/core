// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestExportIntegration tests the integration of AuditMiddleware with the new exporters.
func TestExportIntegration(t *testing.T) {
	// Mock Splunk Server
	var splunkReceived int32
	splunkSignal := make(chan struct{}, 10)
	splunkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&splunkReceived, 1)
		w.WriteHeader(http.StatusOK)
		splunkSignal <- struct{}{}
	}))
	defer splunkServer.Close()

	// Configure middleware with Splunk
	storageType := configv1.AuditConfig_STORAGE_TYPE_SPLUNK
	config := configv1.AuditConfig_builder{
		Enabled:     proto.Bool(true),
		StorageType: &storageType,
		Splunk: configv1.SplunkConfig_builder{
			HecUrl:     proto.String(splunkServer.URL),
			Token:      proto.String("token"),
			Index:      proto.String("main"),
			Source:     proto.String("source"),
			Sourcetype: proto.String("json"),
		}.Build(),
	}.Build()

	m, err := NewAuditMiddleware(config)
	require.NoError(t, err)
	defer m.Close()

	// Execute a tool
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "result", nil
	}

	req := &tool.ExecutionRequest{
		ToolName: "test_integration_tool",
	}

	_, err = m.Execute(context.Background(), req, next)
	assert.NoError(t, err)

	// Wait for async log
	select {
	case <-splunkSignal:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for Splunk log")
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&splunkReceived))
}
