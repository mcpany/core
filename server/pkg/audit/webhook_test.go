package audit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookAuditStore(t *testing.T) {
	// Setup test server
	var mu sync.Mutex
	var receivedEntries []Entry
	var receivedHeaders http.Header
	done := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		receivedHeaders = r.Header
		var batch []Entry
		err := json.NewDecoder(r.Body).Decode(&batch)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		receivedEntries = append(receivedEntries, batch...)
		w.WriteHeader(http.StatusOK)
		select {
		case <-done:
		default:
			close(done)
		}
	}))
	defer server.Close()

	// Setup store
	headers := map[string]string{
		"X-Custom-Header": "test-value",
	}
	store := NewWebhookAuditStore(server.URL, headers)

	// Create test entry
	entry := Entry{
		Timestamp:  time.Now().Truncate(time.Second),
		ToolName:   "test-tool",
		UserID:     "test-user",
		DurationMs: 100,
		Arguments:  json.RawMessage(`{"arg": "value"}`),
		Result:     "success",
		Error:      "",
	}

	// Test Write
	err := store.Write(context.Background(), entry)
	require.NoError(t, err)

	// Flush and close
	err = store.Close()
	require.NoError(t, err)

	// Wait for server to receive
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for webhook")
	}

	// Verify received data
	mu.Lock()
	defer mu.Unlock()
	require.Len(t, receivedEntries, 1)
	receivedEntry := receivedEntries[0]
	assert.Equal(t, entry.ToolName, receivedEntry.ToolName)
	assert.Equal(t, entry.UserID, receivedEntry.UserID)
	assert.Equal(t, entry.DurationMs, receivedEntry.DurationMs)
	assert.JSONEq(t, string(entry.Arguments), string(receivedEntry.Arguments))

	// Verify headers
	assert.Equal(t, "test-value", receivedHeaders.Get("X-Custom-Header"))
	assert.Equal(t, "application/json", receivedHeaders.Get("Content-Type"))
}

func TestWebhookAuditStore_Batch(t *testing.T) {
	var mu sync.Mutex
	var receivedEntries []Entry
	done := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var batch []Entry
		_ = json.NewDecoder(r.Body).Decode(&batch)
		mu.Lock()
		receivedEntries = append(receivedEntries, batch...)
		if len(receivedEntries) >= 5 {
			select {
			case <-done:
			default:
				close(done)
			}
		}
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := NewWebhookAuditStore(server.URL, nil)

	for i := 0; i < 5; i++ {
		_ = store.Write(context.Background(), Entry{ToolName: "tool"})
	}

	// Close to flush
	_ = store.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, receivedEntries, 5)
}
