// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestDatadogAuditStore(t *testing.T) {
	var receivedCount int32
	received := make(chan struct{}, 10)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "my-api-key", r.Header.Get("DD-API-KEY"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, "mcpany", body["ddsource"])
		assert.Equal(t, "my-service", body["service"])
		assert.Equal(t, "env:prod", body["ddtags"])

		message, ok := body["message"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "test-tool", message["tool_name"])

		w.WriteHeader(http.StatusAccepted)
		atomic.AddInt32(&receivedCount, 1)
		received <- struct{}{}
	}))
	defer ts.Close()

	config := &configv1.DatadogConfig{
		ApiKey:  proto.String("my-api-key"),
		Site:    proto.String("datadoghq.com"),
		Service: proto.String("my-service"),
		Tags:    proto.String("env:prod"),
	}

	store := NewDatadogAuditStore(config)
	store.url = ts.URL
	defer store.Close()

	entry := AuditEntry{
		Timestamp:  time.Now(),
		ToolName:   "test-tool",
		Duration:   "100ms",
		DurationMs: 100,
	}

	err := store.Write(context.Background(), entry)
	assert.NoError(t, err)

	// Wait for processing
	select {
	case <-received:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for datadog event")
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&receivedCount))
}
