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
//
// Summary: Represents a Span.
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
//
// Summary: Represents a Trace.
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

func toSpan(entry audit.Entry) Span {
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
		b, err := json.Marshal(entry.Result)
		if err == nil {
			_ = json.Unmarshal(b, &output)
		}
	}

	return Span{
		ID:           entry.SpanID,
		Name:         entry.ToolName,
		Type:         "tool",
		StartTime:    startTime,
		EndTime:      endTime,
		Status:       status,
		Input:        input,
		Output:       output,
		ErrorMessage: entry.Error,
		Children:     []Span{},
	}
}

func buildSpanTree(parentID string, entries []audit.Entry) []Span {
	var children []Span
	for _, e := range entries {
		if e.ParentID == parentID {
			span := toSpan(e)
			span.Children = buildSpanTree(e.SpanID, entries)
			children = append(children, span)
		}
	}
	return children
}

func buildSingleTraceTree(traceID string, entries []audit.Entry) *Trace {
	var rootEntry audit.Entry
	var hasRoot bool
	for _, e := range entries {
		if e.ParentID == "" {
			rootEntry = e
			hasRoot = true
			break
		}
	}

	if !hasRoot {
		if len(entries) > 0 {
			rootEntry = entries[0]
		} else {
			return nil
		}
	}

	rootSpan := toSpan(rootEntry)
	rootSpan.Children = buildSpanTree(rootEntry.SpanID, entries)

	status := statusSuccess
	if rootEntry.Error != "" {
		status = statusError
	}

	durationMs := rootEntry.DurationMs
	if len(entries) > 1 {
		// Calculate total duration as max end time minus min start time
		minStart := rootSpan.StartTime
		maxEnd := rootSpan.EndTime
		for _, e := range entries {
			span := toSpan(e)
			if span.StartTime < minStart {
				minStart = span.StartTime
			}
			if span.EndTime > maxEnd {
				maxEnd = span.EndTime
			}
		}
		durationMs = maxEnd - minStart
	}

	return &Trace{
		ID:            traceID,
		RootSpan:      rootSpan,
		Timestamp:     rootEntry.Timestamp.Format(time.RFC3339),
		TotalDuration: durationMs,
		Status:        status,
		Trigger:       "user",
	}
}

func buildTraceTrees(entries []audit.Entry) []*Trace {
	// Group entries by TraceID
	grouped := make(map[string][]audit.Entry)
	for _, e := range entries {
		// If TraceID is empty, generate a fallback ID
		tID := e.TraceID
		if tID == "" {
			// Use fallback ID based on unique execution so traces don't group together unintentionally
			data := fmt.Sprintf("%d-%s-%s-%s-%s", e.Timestamp.UnixNano(), e.ToolName, e.UserID, e.ProfileID, e.SpanID)
			hash := sha256.Sum256([]byte(data))
			tID = hex.EncodeToString(hash[:])
		}

		// If SpanID is empty, generate fallback based on trace ID
		if e.SpanID == "" {
			e.SpanID = tID + "-0"
		}
		grouped[tID] = append(grouped[tID], e)
	}

	var traces []*Trace
	for traceID, group := range grouped {
		trace := buildSingleTraceTree(traceID, group)
		if trace != nil {
			traces = append(traces, trace)
		}
	}

	return traces
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

			var entries []audit.Entry
			for _, item := range history {
				if entry, ok := item.(audit.Entry); ok {
					entries = append(entries, entry)
				}
			}

			traces = buildTraceTrees(entries)

			// ⚡ BOLT: Optimized trace retrieval
			// Randomized Selection from Top 5 High-Impact Targets
			// Only return the requested number of recent traces
			limitStr := r.URL.Query().Get("limit")
			limit := len(traces)
			if limitStr != "" {
				if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
					limit = parsed
				}
			}

			// Sort by timestamp descending (newest first)
			// (Assuming buildTraceTrees doesn't guarantee order, though map iteration is random)
			for i := 0; i < len(traces); i++ {
				for j := i + 1; j < len(traces); j++ {
					// Use ascending check inside condition to push older items to the end
					if traces[j].Timestamp > traces[i].Timestamp {
						traces[i], traces[j] = traces[j], traces[i]
					} else if traces[j].Timestamp == traces[i].Timestamp {
						// Tie breaker on StartTime or Trace ID to be stable
						if traces[j].RootSpan.StartTime > traces[i].RootSpan.StartTime {
							traces[i], traces[j] = traces[j], traces[i]
						} else if traces[j].RootSpan.StartTime == traces[i].RootSpan.StartTime && traces[j].ID > traces[i].ID {
							traces[i], traces[j] = traces[j], traces[i]
						}
					}
				}
			}

			if len(traces) > limit {
				traces = traces[:limit]
			}
		}

		if traces == nil {
			traces = []*Trace{}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(traces)
	}
}

func (a *Application) handleClearTraces() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if a.standardMiddlewares != nil && a.standardMiddlewares.Audit != nil {
			a.standardMiddlewares.Audit.ClearHistory()
			logging.GetLogger().Info("Cleared trace history via API")
		}

		w.WriteHeader(http.StatusNoContent)
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

		// Build and send history
		var entries []audit.Entry
		for _, msg := range history {
			if entry, ok := msg.(audit.Entry); ok {
				entries = append(entries, entry)
			}
		}

		initialTraces := buildTraceTrees(entries)
		for _, trace := range initialTraces {
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

				// Rebuild the trace tree for this specific TraceID
				// Fetch the latest history to get all spans for this trace
				currentHistory := a.standardMiddlewares.Audit.GetHistory()
				var traceEntries []audit.Entry
				for _, hItem := range currentHistory {
					if hEntry, hOk := hItem.(audit.Entry); hOk && hEntry.TraceID == entry.TraceID {
						traceEntries = append(traceEntries, hEntry)
					}
				}

				// Ensure the incoming entry is included in case it wasn't in history yet
				found := false
				for _, e := range traceEntries {
					if e.SpanID == entry.SpanID {
						found = true
						break
					}
				}
				if !found {
					traceEntries = append(traceEntries, entry)
				}

				trace := buildSingleTraceTree(entry.TraceID, traceEntries)
				if trace == nil {
					continue
				}

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
		"query":   "Analyze Q3 financial report",
		"context": "user-session-123",
	})
	child1Args, _ := json.Marshal(map[string]any{
		"query": "Q3 2024 financials",
	})
	child2Args, _ := json.Marshal(map[string]any{
		"files": []string{"data_q3.xlsx"},
	})

	entries := []audit.Entry{
		{
			Timestamp: now,
			ToolName:  "orchestrator-task",
			UserID:    "system",
			ProfileID: "default",
			TraceID:   traceID,
			SpanID:    traceID + "-0",
			ParentID:  "",
			Arguments: json.RawMessage(rootArgs),
			Result: map[string]any{
				"summary":    "Revenue up 15%",
				"confidence": 0.98,
			},
			Duration:   "1250ms",
			DurationMs: 1250,
		},
		{
			Timestamp: now.Add(50 * time.Millisecond),
			ToolName:  "search-tool",
			UserID:    "system",
			ProfileID: "default",
			TraceID:   traceID,
			SpanID:    traceID + "-1",
			ParentID:  traceID + "-0",
			Arguments: json.RawMessage(child1Args),
			Result: map[string]any{
				"results": []string{"report_q3.pdf", "data_q3.xlsx"},
			},
			Duration:   "400ms",
			DurationMs: 400,
		},
		{
			Timestamp: now.Add(500 * time.Millisecond),
			ToolName:  "data-analyzer",
			UserID:    "system",
			ProfileID: "default",
			TraceID:   traceID,
			SpanID:    traceID + "-2",
			ParentID:  traceID + "-0",
			Arguments: json.RawMessage(child2Args),
			Result: map[string]any{
				"analysis": "Growth detected",
				"metrics": map[string]any{
					"revenue": 1.15,
				},
			},
			Duration:   "700ms",
			DurationMs: 700,
		},
	}
	return entries
}
