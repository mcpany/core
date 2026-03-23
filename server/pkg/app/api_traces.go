// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"math/rand"

	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/logging"
)

// Span represents a span in a trace.
type Span struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	ServiceName  string         `json:"serviceName,omitempty"`
	Type         string         `json:"type"`
	StartTime    int64          `json:"startTime"` // Unix millis
	EndTime      int64          `json:"endTime"`   // Unix millis
	Status       string         `json:"status"`    // success, error, pending
	Input        map[string]any `json:"input,omitempty"`
	Output       map[string]any `json:"output,omitempty"`
	ErrorMessage string         `json:"errorMessage,omitempty"`
	Children     []Span         `json:"children,omitempty"`
}

// Trace represents a full trace.
type Trace struct {
	ID            string `json:"id"`
	RootSpan      Span   `json:"rootSpan"`
	Timestamp     string `json:"timestamp"` // ISO 8601
	TotalDuration int64  `json:"totalDuration"`
	Status        string `json:"status"`
	Trigger       string `json:"trigger"`
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

		var traces []*Trace

		// 1. Get real audit logs
		if a.standardMiddlewares != nil && a.standardMiddlewares.Audit != nil {
			history := a.standardMiddlewares.Audit.GetHistory()

			// ⚡ BOLT: Optimized trace retrieval
			// Randomized Selection from Top 5 High-Impact Targets
			// Only unmarshal the requested number of recent traces to save CPU and bandwidth.
			limitStr := r.URL.Query().Get("limit")
			limit := len(history)
			if limitStr != "" {
				if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
					limit = parsed
				}
			}

			// Determine start index. History is chronological (oldest -> newest).
			// We want the last `limit` items.
			startIdx := 0
			if len(history) > limit {
				startIdx = len(history) - limit
			}

			// Iterate backwards from end to startIdx to return newest first
			for i := len(history) - 1; i >= startIdx; i-- {
				if entry, ok := history[i].(audit.Entry); ok {
					traces = append(traces, toTrace(entry))
				}
			}
		}

		if traces == nil {
			traces = []*Trace{}
		}

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
			entry, ok := msg.(audit.Entry)
			if !ok {
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

		pingTicker := time.NewTicker(5 * time.Second)
		defer pingTicker.Stop()

		for {
			select {
			case <-pingTicker.C:
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
					return
				}
			case msg, ok := <-logCh:
				if !ok {
					return
				}
				entry, ok := msg.(audit.Entry)
				if !ok {
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
}

func (a *Application) handleDebugSeedTraces() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if a.standardMiddlewares == nil || a.standardMiddlewares.Audit == nil {
			http.Error(w, "Audit middleware not enabled", http.StatusInternalServerError)
			return
		}

		entries := generateMockAuditEntries()

		for _, entry := range entries {
			if err := a.standardMiddlewares.Audit.Write(r.Context(), entry); err != nil {
				logging.GetLogger().Error("failed to seed trace to audit db", "error", err)
				http.Error(w, "Failed to seed trace", http.StatusInternalServerError)
				return
			}
		}

		logging.GetLogger().Info("Seeded debug trace to database", "id", entries[0].TraceID)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "seeded", "id": entries[0].TraceID})
	}
}

func generateMockAuditEntries() []audit.Entry {
	now := time.Now()
	traceID := fmt.Sprintf("trace-seed-%d", rand.Intn(10000)) //nolint:gosec // Testing only

	rootArgs, _ := json.Marshal(map[string]any{
		"issue_id": "INC-8492",
		"description": "Database latency spike in eu-west-1",
	})

	slackArgs, _ := json.Marshal(map[string]any{
		"channel": "#incident-response",
		"message": "Investigating INC-8492: Database latency spike in eu-west-1",
	})

	queryArgs, _ := json.Marshal(map[string]any{
		"query": "SELECT avg(latency) FROM metrics WHERE service='db-primary' AND region='eu-west-1' AND time > now() - 1h",
	})

	getLogsArgs, _ := json.Marshal(map[string]any{
		"service": "db-primary",
		"level": "ERROR",
		"limit": 100,
	})

	analyzeArgs, _ := json.Marshal(map[string]any{
		"logs_context": "Found 45 deadlock errors in the last 15 minutes.",
		"metrics_context": "Average latency is 1450ms, p99 is 4500ms.",
	})

	jiraArgs, _ := json.Marshal(map[string]any{
		"issue_id": "INC-8492",
		"status": "In Progress",
		"comment": "Root cause identified as deadlocks. Applying mitigation plan.",
	})

	entries := []audit.Entry{
		{
			Timestamp:  now,
			ToolName:   "incident-triage-agent",
			UserID:     "admin-user",
			ProfileID:  "prod-incident-profile",
			TraceID:    traceID,
			SpanID:     traceID + "-0",
			ParentID:   "",
			Arguments:  json.RawMessage(rootArgs),
			Result: map[string]any{
				"status": "mitigated",
				"resolution_time_ms": 4250,
				"confidence": 0.95,
			},
			Duration:   "4250ms",
			DurationMs: 4250,
		},
		{
			Timestamp:  now.Add(100 * time.Millisecond),
			ToolName:   "slack-notify",
			UserID:     "admin-user",
			ProfileID:  "prod-incident-profile",
			TraceID:    traceID,
			SpanID:     traceID + "-1",
			ParentID:   traceID + "-0",
			Arguments:  json.RawMessage(slackArgs),
			Result: map[string]any{
				"message_ts": "1672531200.000100",
				"delivered": true,
			},
			Duration:   "150ms",
			DurationMs: 150,
		},
		{
			Timestamp:  now.Add(300 * time.Millisecond),
			ToolName:   "query-metrics",
			UserID:     "admin-user",
			ProfileID:  "prod-incident-profile",
			TraceID:    traceID,
			SpanID:     traceID + "-2",
			ParentID:   traceID + "-0",
			Arguments:  json.RawMessage(queryArgs),
			Result: map[string]any{
				"avg_latency_ms": 1450,
				"p99_latency_ms": 4500,
				"status": "critical",
			},
			Duration:   "850ms",
			DurationMs: 850,
		},
		{
			Timestamp:  now.Add(1200 * time.Millisecond),
			ToolName:   "fetch-logs",
			UserID:     "admin-user",
			ProfileID:  "prod-incident-profile",
			TraceID:    traceID,
			SpanID:     traceID + "-3",
			ParentID:   traceID + "-0",
			Arguments:  json.RawMessage(getLogsArgs),
			Error:      "Timeout exceeding 2000ms while fetching logs from Elasticsearch cluster",
			Duration:   "2000ms",
			DurationMs: 2000,
		},
		{
			Timestamp:  now.Add(3250 * time.Millisecond),
			ToolName:   "analyze-incident-data",
			UserID:     "admin-user",
			ProfileID:  "prod-incident-profile",
			TraceID:    traceID,
			SpanID:     traceID + "-4",
			ParentID:   traceID + "-0",
			Arguments:  json.RawMessage(analyzeArgs),
			Result: map[string]any{
				"root_cause": "Deadlock in inventory_updates table",
				"recommended_action": "Kill blocked transactions and restart connection pool",
			},
			Duration:   "500ms",
			DurationMs: 500,
		},
		{
			Timestamp:  now.Add(3800 * time.Millisecond),
			ToolName:   "jira-update",
			UserID:     "admin-user",
			ProfileID:  "prod-incident-profile",
			TraceID:    traceID,
			SpanID:     traceID + "-5",
			ParentID:   traceID + "-0",
			Arguments:  json.RawMessage(jiraArgs),
			Result: map[string]any{
				"success": true,
				"issue_url": "https://jira.example.com/browse/INC-8492",
			},
			Duration:   "400ms",
			DurationMs: 400,
		},
	}
	return entries
}
