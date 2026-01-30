// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/logging"
)

// Span represents a span in a trace.
type Span struct {
	// ID is the unique identifier for the span.
	ID string `json:"id"`
	// Name is the name of the operation.
	Name string `json:"name"`
	// Type is the type of the span (e.g., "tool", "http").
	Type string `json:"type"`
	// StartTime is when the span started (Unix millis).
	StartTime int64 `json:"startTime"` // Unix millis
	// EndTime is when the span ended (Unix millis).
	EndTime int64 `json:"endTime"` // Unix millis
	// Status is the status of the operation (success, error, pending).
	Status string `json:"status"` // success, error, pending
	// Input is the input data for the operation.
	Input map[string]any `json:"input,omitempty"`
	// Output is the output data from the operation.
	Output map[string]any `json:"output,omitempty"`
	// ErrorMessage describes any error that occurred.
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// Trace represents a full trace.
type Trace struct {
	// ID is the unique identifier for the trace.
	ID string `json:"id"`
	// RootSpan is the root span of the trace.
	RootSpan Span `json:"rootSpan"`
	// Timestamp is when the trace started (ISO 8601).
	Timestamp string `json:"timestamp"` // ISO 8601
	// TotalDuration is the total duration of the trace in milliseconds.
	TotalDuration int64 `json:"totalDuration"`
	// Status is the overall status of the trace.
	Status string `json:"status"`
	// Trigger describes what triggered the trace.
	Trigger string `json:"trigger"`
}

const (
	statusSuccess = "success"
	statusError   = "error"
)

func toTrace(entry audit.Entry) *Trace {
	// Generate deterministic ID based on content to prevent duplicates during history replay
	data := fmt.Sprintf("%d-%s-%s-%s", entry.Timestamp.UnixNano(), entry.ToolName, entry.UserID, entry.ProfileID)
	hash := sha256.Sum256([]byte(data))
	traceID := hex.EncodeToString(hash[:])

	// Span ID can be same or derived
	spanID := traceID + "-0"

	status := statusSuccess
	if entry.Error != "" {
		status = statusError
	}

	startTime := entry.Timestamp.UnixMilli()
	durationMs := entry.DurationMs
	endTime := startTime + durationMs

	var input map[string]any
	if len(entry.Arguments) > 0 {
		_ = json.Unmarshal(entry.Arguments, &input)
	}

	var output map[string]any
	if entry.Result != nil {
		// entry.Result is already an interface{}, but if it's a map/struct it works.
		// If it's a primitive, we might want to wrap it?
		// For now assume map or convertible.
		b, err := json.Marshal(entry.Result)
		if err == nil {
			_ = json.Unmarshal(b, &output)
		}
	}

	span := Span{
		ID:           spanID,
		Name:         entry.ToolName,
		Type:         "tool",
		StartTime:    startTime,
		EndTime:      endTime,
		Status:       status,
		Input:        input,
		Output:       output,
		ErrorMessage: entry.Error,
	}

	return &Trace{
		ID:            traceID,
		RootSpan:      span,
		Timestamp:     entry.Timestamp.Format(time.RFC3339),
		TotalDuration: durationMs,
		Status:        status,
		Trigger:       "user", // Default to user for now
	}
}

func (a *Application) handleTraces() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if a.standardMiddlewares == nil || a.standardMiddlewares.Audit == nil {
			// If audit is disabled, return empty list
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("[]"))
			return
		}

		history := a.standardMiddlewares.Audit.GetHistory()
		var traces []*Trace

		for _, msg := range history {
			var entry audit.Entry
			if err := json.Unmarshal(msg, &entry); err == nil {
				traces = append(traces, toTrace(entry))
			}
		}

		// Reverse to show newest first? Or keeps chronological?
		// The history appends new items. So it's oldest first.
		// UI usually expects newest first or handles sort.
		// Let's send chronological (oldest -> newest).

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(traces)
	}
}

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
			// If audit is disabled, just close or keep open but send nothing?
			// Better to send a close message.
			_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Audit disabled"), time.Now().Add(time.Second))
			return
		}

		// Subscribe to traces with history
		logCh, history := a.standardMiddlewares.Audit.SubscribeWithHistory()
		defer a.standardMiddlewares.Audit.Unsubscribe(logCh)

		// Set write deadline
		if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
			logging.GetLogger().Error("failed to set write deadline", "error", err)
			return
		}
		conn.SetPongHandler(func(string) error {
			return conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		})

		// Send history
		for _, msg := range history {
			var entry audit.Entry
			if err := json.Unmarshal(msg, &entry); err != nil {
				continue
			}
			trace := toTrace(entry)

			if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				logging.GetLogger().Error("failed to set write deadline", "error", err)
				return
			}
			if err := conn.WriteJSON(trace); err != nil {
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

		for msg := range logCh {
			var entry audit.Entry
			if err := json.Unmarshal(msg, &entry); err != nil {
				continue
			}
			trace := toTrace(entry)

			if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				logging.GetLogger().Error("failed to set write deadline", "error", err)
				return
			}
			if err := conn.WriteJSON(trace); err != nil {
				logging.GetLogger().Error("failed to write trace to websocket", "error", err)
				return
			}
		}
	}
}
