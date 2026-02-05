// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/stretchr/testify/require"
)

func TestLogPersistence(t *testing.T) {
	// 1. Setup DB Paths
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "mcpany.db")

	// 2. Pre-seed Logs
	// We need to use the SAME logic as the server to open the DB.
	// Since we are linking against the same codebase, we can use logging.NewSQLiteLogStore directly.
	store, err := logging.NewSQLiteLogStore(dbPath)
	require.NoError(t, err)

	historicMsg := fmt.Sprintf("Historic Log %d", time.Now().UnixNano())
	entry := logging.LogEntry{
		ID:        "historic-1",
		Timestamp: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		Level:     "INFO",
		Message:   historicMsg,
		Source:    "test-seeder",
	}
	require.NoError(t, store.Write(entry))
	require.NoError(t, store.Close())

	// 3. Start Server (External Process)
	// We pass --db-path
	serverInfo := integration.StartMCPANYServer(t, "LogPersistence", "--db-path", dbPath)
	defer serverInfo.CleanupFunc()

	// 4. Connect WebSocket
	// The endpoint is /ws/logs. The helper gives us HTTP/JSONRPC endpoints.
	require.NotEmpty(t, serverInfo.JSONRPCEndpoint)

	wsBase := serverInfo.JSONRPCEndpoint
	if len(wsBase) > 4 && wsBase[:4] == "http" {
		wsBase = "ws" + wsBase[4:]
	}

	// Extract API Key from HTTPEndpoint (helper adds it there)
	u, err := url.Parse(serverInfo.HTTPEndpoint)
	require.NoError(t, err)
	apiKey := u.Query().Get("api_key")

	wsURL := fmt.Sprintf("%s/api/v1/ws/logs?api_key=%s", wsBase, apiKey)

	t.Logf("Connecting to WS: %s", wsURL)

	dialer := websocket.Dialer{HandshakeTimeout: 5 * time.Second}
	conn, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// 5. Verify Logs
	found := false
	timeout := time.After(10 * time.Second)

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for historic log")
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				// If server closes connection (e.g. shutdown), we fail
				t.Fatalf("Read error: %v", err)
			}

			var recvEntry logging.LogEntry
			if err := json.Unmarshal(msg, &recvEntry); err == nil {
				if recvEntry.Message == historicMsg {
					found = true
					goto DONE
				}
			}
		}
	}
DONE:
	require.True(t, found, "Historic log not found in stream")
}
