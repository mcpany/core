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
	"google.golang.org/protobuf/proto"
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

	config := &configv1.SplunkConfig{
		HecUrl:     proto.String(ts.URL + "/services/collector/event"),
		Token:      proto.String("my-token"),
		Index:      proto.String("my-index"),
		Source:     proto.String("my-source"),
		Sourcetype: proto.String("my-sourcetype"),
	}

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
		newVal := atomic.AddInt32(&totalReceived, int32(count))
		w.WriteHeader(http.StatusOK)
		if newVal >= 5 {
			select {
			case done <- struct{}{}:
			default:
			}
		}
	}))
	defer ts.Close()

	config := &configv1.SplunkConfig{
		HecUrl: proto.String(ts.URL),
	}
	store := NewSplunkAuditStore(config)

	for i := 0; i < 5; i++ {
		_ = store.Write(context.Background(), Entry{ToolName: "test"})
	}

	// Wait a bit to ensure worker picks up everything (though Close should handle it)
	err := store.Close()
	assert.NoError(t, err)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		if atomic.LoadInt32(&totalReceived) < 5 {
			t.Fatal("timed out waiting for splunk batch")
		}
	}

	assert.Equal(t, int32(5), atomic.LoadInt32(&totalReceived))
}

func TestSplunkAuditStore_QueueFull(t *testing.T) {
	config := &configv1.SplunkConfig{HecUrl: proto.String("http://localhost")}
	store := NewSplunkAuditStore(config)

	// Close existing workers
	close(store.queue)
	store.wg.Wait()

	store.queue = make(chan Entry, splunkBufferSize)

	// Fill queue
	for i := 0; i < splunkBufferSize; i++ {
		store.queue <- Entry{}
	}

	// Next write should fail
	err := store.Write(context.Background(), Entry{})
	assert.Error(t, err)
	assert.Equal(t, "audit queue full", err.Error())
}

func TestSplunkAuditStore_Read_NotImplemented(t *testing.T) {
	store := NewSplunkAuditStore(nil)
	defer store.Close()
	_, err := store.Read(context.Background(), Filter{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read not implemented")
}

func TestSplunkAuditStore_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	config := &configv1.SplunkConfig{HecUrl: proto.String(ts.URL)}
	store := NewSplunkAuditStore(config)

	err := store.Write(context.Background(), Entry{ToolName: "test"})
	assert.NoError(t, err)

	store.Close()
}
