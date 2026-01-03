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

func TestSplunkAuditStore(t *testing.T) {
	var receivedCount int32
	received := make(chan struct{}, 10)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/services/collector/event", r.URL.Path)
		assert.Equal(t, "Splunk my-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, "my-source", body["source"])
		assert.Equal(t, "my-sourcetype", body["sourcetype"])
		assert.Equal(t, "my-index", body["index"])

		event, ok := body["event"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "test-tool", event["tool_name"])

		w.WriteHeader(http.StatusOK)
		atomic.AddInt32(&receivedCount, 1)
		received <- struct{}{}
	}))
	defer ts.Close()

	config := &configv1.SplunkConfig{
		HecUrl:     proto.String(ts.URL + "/services/collector/event"),
		Token:      proto.String("my-token"),
		Index:      proto.String("my-index"),
		Source:     proto.String("my-source"),
		Sourcetype: proto.String("my-sourcetype"),
	}

	store := NewSplunkAuditStore(config)
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
		t.Fatal("timed out waiting for splunk event")
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&receivedCount))
}
