// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/logging"
)

const (
	statusError   = "error"
	statusSuccess = "success"
)

// Trace represents a full trace as expected by the UI.
type Trace struct {
	ID            string `json:"id"`
	RootSpan      Span   `json:"rootSpan"`
	Timestamp     string `json:"timestamp"` // ISO8601
	TotalDuration int64  `json:"totalDuration"`
	Status        string `json:"status"`
	Trigger       string `json:"trigger"`
}

// Span represents a span in a trace.
type Span struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	StartTime    int64       `json:"startTime"`
	EndTime      int64       `json:"endTime"`
	Status       string      `json:"status"`
	Input        interface{} `json:"input,omitempty"`
	Output       interface{} `json:"output,omitempty"`
	ErrorMessage string      `json:"errorMessage,omitempty"`
}

// handleTraces handles GET /api/v1/traces.
func (a *Application) handleTraces() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if a.standardMiddlewares == nil || a.standardMiddlewares.Audit == nil {
			http.Error(w, "Audit middleware not configured", http.StatusServiceUnavailable)
			return
		}

		// We use SubscribeWithHistory to get the history buffer from the broadcaster.
		// We immediately unsubscribe because we only want the snapshot for HTTP GET.
		ch, historyBytes := a.standardMiddlewares.Audit.SubscribeWithHistory()
		if ch != nil {
			a.standardMiddlewares.Audit.Unsubscribe(ch)
		}

		traces := make([]Trace, 0, len(historyBytes))
		// Iterate backwards to show newest first? Or frontend handles sort?
		// Usually API returns chronological or reverse chronological.
		// Broadcaster history is chronological (oldest first).
		// Let's reverse it so newest is first.
		for i := len(historyBytes) - 1; i >= 0; i-- {
			var entry audit.Entry
			if err := json.Unmarshal(historyBytes[i], &entry); err != nil {
				continue
			}
			traces = append(traces, auditEntryToTrace(entry))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(traces)
	}
}

// handleTracesWS handles WebSocket connections for trace streaming.
func (a *Application) handleTracesWS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logging.GetLogger().Error("failed to upgrade to websocket", "error", err)
			return
		}
		defer func() {
			if err := conn.Close(); err != nil {
				logging.GetLogger().Error("failed to close websocket connection", "error", err)
			}
		}()

		if a.standardMiddlewares == nil || a.standardMiddlewares.Audit == nil {
			logging.GetLogger().Error("Audit middleware not configured for trace streaming")
			return
		}

		// Subscribe to audit logs
		ch, historyBytes := a.standardMiddlewares.Audit.SubscribeWithHistory()
		if ch == nil {
			logging.GetLogger().Error("Failed to subscribe to audit logs")
			return
		}
		defer a.standardMiddlewares.Audit.Unsubscribe(ch)

		// Set write deadline
		if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
			logging.GetLogger().Error("failed to set write deadline", "error", err)
			return
		}
		conn.SetPongHandler(func(string) error {
			return conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		})

		// Send history
		for _, msgBytes := range historyBytes {
			var entry audit.Entry
			if err := json.Unmarshal(msgBytes, &entry); err != nil {
				continue
			}
			trace := auditEntryToTrace(entry)
			traceBytes, _ := json.Marshal(trace)

			if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				logging.GetLogger().Error("failed to set write deadline", "error", err)
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, traceBytes); err != nil {
				logging.GetLogger().Error("failed to write history trace to websocket", "error", err)
				return
			}
		}

		// Send ping periodically
		go func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
					return
				}
			}
		}()

		// Stream new entries
		for msgBytes := range ch {
			var entry audit.Entry
			if err := json.Unmarshal(msgBytes, &entry); err != nil {
				logging.GetLogger().Error("Failed to unmarshal audit entry", "error", err)
				continue
			}
			trace := auditEntryToTrace(entry)
			traceBytes, _ := json.Marshal(trace)

			if err := conn.WriteMessage(websocket.TextMessage, traceBytes); err != nil {
				logging.GetLogger().Error("failed to write trace to websocket", "error", err)
				return
			}
			if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				logging.GetLogger().Error("failed to set write deadline", "error", err)
				return
			}
		}
	}
}

func auditEntryToTrace(entry audit.Entry) Trace {
	// Generate a deterministic ID based on timestamp and content if possible, or just random
	// Since we don't have a unique ID in audit entry yet, we use UUID.
	// Note: This means ID will change on every fetch! This is bad for React keys.
	// Ideally AuditEntry should have an ID.
	// But since we are only reading, maybe we can hash the entry?
	// For now, let's use a deterministic UUID based on content to keep it stable across re-fetches if possible.
	// But broadcasting happens once.
	// However, `handleTraces` (GET) re-reads history. If we generate random UUIDs, they will differ from WS stream?
	// Yes.
	// FIX: We should rely on `entry` having an ID?
	// `audit.Entry` struct in `server/pkg/audit/types.go` likely doesn't have ID.
	// Let's generate one based on Timestamp + ToolName + Duration (entropy).
	// This is "stable enough" for display.
	traceID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("%d-%s-%d", entry.Timestamp.UnixNano(), entry.ToolName, entry.DurationMs))).String()

	status := statusSuccess
	if entry.Error != "" {
		status = statusError
	}

	startTime := entry.Timestamp.UnixMilli()
	endTime := startTime + entry.DurationMs

	// Parse arguments if possible
	var input interface{}
	if len(entry.Arguments) > 0 {
		_ = json.Unmarshal(entry.Arguments, &input)
	}
	// Result is already interface{}

	return Trace{
		ID:            traceID,
		Timestamp:     entry.Timestamp.Format(time.RFC3339),
		TotalDuration: entry.DurationMs,
		Status:        status,
		Trigger:       "user", // Default
		RootSpan: Span{
			ID:           traceID + "-span",
			Name:         entry.ToolName,
			Type:         "tool",
			StartTime:    startTime,
			EndTime:      endTime,
			Status:       status,
			Input:        input,
			Output:       entry.Result,
			ErrorMessage: entry.Error,
		},
	}
}
