package sqlite

import (
    "context"
    "net/http"
    "os"
    "testing"
    "time"

    "github.com/mcpany/core/server/pkg/storage"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLogStorage(t *testing.T) {
    dbPath := "test_logs.db"
    defer os.Remove(dbPath)

    db, err := NewDB(dbPath)
    require.NoError(t, err)
    defer db.Close()

    store := NewStore(db)

    ctx := context.Background()

    entry := &storage.LogEntry{
        ID:        "req-1",
        Timestamp: time.Now().UTC(),
        Method:    "POST",
        Path:      "/api/v1/tools/execute",
        Status:    200,
        Duration:  100 * time.Millisecond,
        RequestHeaders: http.Header{
            "Content-Type": []string{"application/json"},
        },
        ResponseHeaders: http.Header{
            "Content-Type": []string{"application/json"},
        },
        RequestBody:  `{"tool": "test"}`,
        ResponseBody: `{"result": "success"}`,
    }

    // Save
    err = store.SaveLog(ctx, entry)
    require.NoError(t, err)

    // List
    logs, err := store.ListLogs(ctx, 10, 0)
    require.NoError(t, err)
    require.Len(t, logs, 1)

    // Verify
    got := logs[0]
    assert.Equal(t, entry.ID, got.ID)
    assert.Equal(t, entry.Method, got.Method)
    assert.Equal(t, entry.Path, got.Path)
    assert.Equal(t, entry.Status, got.Status)
    // Timestamp might lose precision or timezone info if not handled carefully, but we used UTC
    assert.WithinDuration(t, entry.Timestamp, got.Timestamp, time.Second)
    assert.Equal(t, entry.Duration, got.Duration)
    assert.Equal(t, entry.RequestHeaders.Get("Content-Type"), got.RequestHeaders.Get("Content-Type"))
    assert.Equal(t, entry.RequestBody, got.RequestBody)
    assert.Equal(t, entry.ResponseBody, got.ResponseBody)
}
