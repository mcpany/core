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

func TestDatadogAuditStore(t *testing.T) {
	var receivedCount int32
	received := make(chan struct{}, 10)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "my-api-key", r.Header.Get("DD-API-KEY"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var logs []map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&logs)
		require.NoError(t, err)

		for _, body := range logs {
			assert.Equal(t, "mcpany", body["ddsource"])
			assert.Equal(t, "my-service", body["service"])
			assert.Equal(t, "env:prod", body["ddtags"])

			message, ok := body["message"].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, "test-tool", message["tool_name"])

			atomic.AddInt32(&receivedCount, 1)
		}

		w.WriteHeader(http.StatusAccepted)
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
		t.Fatal("timed out waiting for datadog event")
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&receivedCount))
}

func TestDatadogAuditStore_Batch(t *testing.T) {
	var totalReceived int32
	done := make(chan struct{}, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var logs []map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&logs)
		// We might receive less than 5 in one request if batching sends partial results
		newVal := atomic.AddInt32(&totalReceived, int32(len(logs)))
		w.WriteHeader(http.StatusAccepted)
		if newVal >= 5 {
			select {
			case done <- struct{}{}:
			default:
			}
		}
	}))
	defer ts.Close()

	config := &configv1.DatadogConfig{
		ApiKey: proto.String("key"),
	}
	store := NewDatadogAuditStore(config)
	store.url = ts.URL

	for i := 0; i < 5; i++ {
		_ = store.Write(context.Background(), Entry{ToolName: "test"})
	}

	// Call store.Close() to ensure everything is flushed.
	// This waits for workers to finish.
	err := store.Close()
	assert.NoError(t, err)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		// Check if we already received enough (maybe done send failed)
		if atomic.LoadInt32(&totalReceived) < 5 {
			t.Fatal("timed out waiting for datadog batch")
		}
	}

	assert.Equal(t, int32(5), atomic.LoadInt32(&totalReceived))
}

func TestDatadogAuditStore_QueueFull(t *testing.T) {
	config := &configv1.DatadogConfig{ApiKey: proto.String("key")}
	store := NewDatadogAuditStore(config)

	// Close queue to stop workers
	close(store.queue)
	store.wg.Wait()

	// Create new full queue
	store.queue = make(chan Entry, datadogBufferSize)
	for i := 0; i < datadogBufferSize; i++ {
		store.queue <- Entry{}
	}

	// Next write should fail
	err := store.Write(context.Background(), Entry{})
	assert.Error(t, err)
	assert.Equal(t, "audit queue full", err.Error())
}

func TestDatadogAuditStore_ConfigDefaults(t *testing.T) {
	store := NewDatadogAuditStore(nil)
	assert.Equal(t, "https://http-intake.logs.datadoghq.com/api/v2/logs", store.url)
	store.Close()

	config := &configv1.DatadogConfig{Site: proto.String("datadoghq.eu")}
	store = NewDatadogAuditStore(config)
	assert.Equal(t, "https://http-intake.logs.datadoghq.eu/api/v2/logs", store.url)
	store.Close()
}

func TestDatadogAuditStore_Read_NotImplemented(t *testing.T) {
	store := NewDatadogAuditStore(nil)
	defer store.Close()
	_, err := store.Read(context.Background(), Filter{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read not implemented")
}

func TestDatadogAuditStore_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	store := NewDatadogAuditStore(nil)
	store.url = ts.URL

	// Should not return error on Write, but log to stderr (hard to test stderr)
	// We just ensure it doesn't panic or block forever
	err := store.Write(context.Background(), Entry{ToolName: "test"})
	assert.NoError(t, err)

	store.Close()
}
