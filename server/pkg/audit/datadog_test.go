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

	config := &configv1.DatadogConfig{}
	config.SetApiKey("my-api-key")
	config.SetSite("datadoghq.com")
	config.SetService("my-service")
	config.SetTags("env:prod")

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
	done := make(chan struct{})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var logs []map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&logs)
		atomic.AddInt32(&totalReceived, int32(len(logs)))
		w.WriteHeader(http.StatusAccepted)
		if atomic.LoadInt32(&totalReceived) >= 5 {
			done <- struct{}{}
		}
	}))
	defer ts.Close()

	config := &configv1.DatadogConfig{}
	config.SetApiKey("key")
	store := NewDatadogAuditStore(config)
	store.url = ts.URL

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

func TestDatadogAuditStore_Read(t *testing.T) {
	store := NewDatadogAuditStore(nil)
	entries, err := store.Read(context.Background(), Filter{})
	assert.Error(t, err)
	assert.Nil(t, entries)
	assert.Equal(t, "read not implemented for datadog audit store", err.Error())
	_ = store.Close()
}
