// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package audit

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
)

func TestSplunkAuditStore(t *testing.T) {
	var receivedCount int32
	received := make(chan struct{}, 10)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/services/collector/event", r.URL.Path)
		assert.Equal(t, "Splunk my-token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		decoder := json.NewDecoder(r.Body)
		for decoder.More() {
			var body map[string]interface{}
			err := decoder.Decode(&body)
			require.NoError(t, err)

			assert.Equal(t, "my-source", body["source"])
			assert.Equal(t, "my-sourcetype", body["sourcetype"])
			assert.Equal(t, "my-index", body["index"])

			event, ok := body["event"].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, "test-tool", event["tool_name"])

			atomic.AddInt32(&receivedCount, 1)
		}
		w.WriteHeader(http.StatusOK)
		received <- struct{}{}
	}))
	defer ts.Close()

	config := &configv1.SplunkConfig{}
	config.SetHecUrl(ts.URL + "/services/collector/event")
	config.SetToken("my-token")
	config.SetIndex("my-index")
	config.SetSource("my-source")
	config.SetSourcetype("my-sourcetype")

	store := NewSplunkAuditStore(config)

	entry := Entry{
		Timestamp:  time.Now(),
		ToolName:   "test-tool",
		Duration:   "100ms",
		DurationMs: 100,
	}

	err := store.Write(context.Background(), entry)
	assert.NoError(t, err)

	// Close to flush
	err = store.Close()
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

func TestSplunkAuditStore_Batch(t *testing.T) {
	var totalReceived int32
	done := make(chan struct{}, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		count := 0
		for decoder.More() {
			var body map[string]interface{}
			_ = decoder.Decode(&body)
			count++
		}
		atomic.AddInt32(&totalReceived, int32(count))
		w.WriteHeader(http.StatusOK)
		if atomic.LoadInt32(&totalReceived) >= 5 {
			done <- struct{}{}
		}
	}))
	defer ts.Close()

	config := &configv1.SplunkConfig{}
	config.SetHecUrl(ts.URL)
	store := NewSplunkAuditStore(config)

	for i := 0; i < 5; i++ {
		_ = store.Write(context.Background(), Entry{ToolName: "test"})
	}

	_ = store.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}

	assert.Equal(t, int32(5), atomic.LoadInt32(&totalReceived))
}
