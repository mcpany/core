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

// Span - Auto-generated documentation.
//
// Summary: Span represents a span in a trace.
//
// Fields:
//   - Various fields for Span.
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

// Trace - Auto-generated documentation.
//
// Summary: Trace represents a full trace.
//
// Fields:
//   - Various fields for Trace.
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

		// 2. Append seeded traces
		a.seededTracesMu.RLock()
		if len(a.seededTraces) > 0 {
			// Seeded traces are stored [Oldest...Newest].
			// We want to prepend them to the list so they appear at the top (Newest First).
			// Iterating forwards and prepending achieves LIFO order in the final list.
			for _, t := range a.seededTraces {
				traces = append([]*Trace{t}, traces...)
			}
		}
		a.seededTracesMu.RUnlock()

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

		// Send seeded traces
		a.seededTracesMu.RLock()
		for _, t := range a.seededTraces {
			if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				logging.GetLogger().Error("failed to set write deadline", "error", err)
				break
			}
			if err := conn.WriteJSON(t); err != nil {
				logging.GetLogger().Error("failed to write seeded trace to websocket", "error", err)
				break
			}
		}
		a.seededTracesMu.RUnlock()

		seededSubCh := make(chan *Trace, 100)
		if a.seededTraceSubs == nil { a.seededTraceSubs = make(map[chan *Trace]struct{}) }; a.seededTraceSubsMu.Lock()
		a.seededTraceSubs[seededSubCh] = struct{}{}
		a.seededTraceSubsMu.Unlock()

		defer func() {
			if a.seededTraceSubs == nil { a.seededTraceSubs = make(map[chan *Trace]struct{}) }; a.seededTraceSubsMu.Lock()
			delete(a.seededTraceSubs, seededSubCh)
			a.seededTraceSubsMu.Unlock()
			close(seededSubCh)
		}()

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
			case trace, ok := <-seededSubCh:
				if !ok {
					return
				}
				if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
					logging.GetLogger().Error("failed to set write deadline", "error", err)
					return
				}
				if err := conn.WriteJSON(trace); err != nil {
					logging.GetLogger().Error("failed to write seeded trace to websocket", "error", err)
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

		trace := generateMockTrace()

		a.seededTracesMu.Lock()
		a.seededTraces = append(a.seededTraces, &trace)
		// Prevent memory leak: cap at 50 traces
		if len(a.seededTraces) > 50 {
			a.seededTraces = a.seededTraces[len(a.seededTraces)-50:]
		}
		a.seededTracesMu.Unlock()

		a.seededTraceSubsMu.RLock()
		for sub := range a.seededTraceSubs {
			select {
			case sub <- &trace:
			default:
				// If channel is full, skip to avoid blocking
			}
		}
		a.seededTraceSubsMu.RUnlock()

		logging.GetLogger().Info("Seeded debug trace", "id", trace.ID)

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "seeded", "id": trace.ID})
	}
}


func generateMockTrace() Trace {
	now := time.Now().UnixMilli()
	traceID := fmt.Sprintf("trace-seed-%d", rand.Intn(10000)) //nolint:gosec // Testing only
	return Trace{
		ID:            traceID,
		Timestamp:     time.Now().Format(time.RFC3339),
		TotalDuration: 1250,
		Status:        "success",
		Trigger:       "user",
		RootSpan: Span{
			ID:        "span-1",
			Name:      "orchestrator-task",
			Type:      "core",
			StartTime: now,
			EndTime:   now + 1250,
			Status:    "success",
			Input: map[string]any{
				"query":   "Analyze Q3 financial report",
				"context": "user-session-123",
			},
			Output: map[string]any{
				"summary":    "Revenue up 15%",
				"confidence": 0.98,
			},
			Children: []Span{
				{
					ID:        "span-2",
					Name:      "search-tool",
					Type:      "tool",
					StartTime: now + 50,
					EndTime:   now + 450,
					Status:    "success",
					Input: map[string]any{
						"query": "Q3 2024 financials",
					},
					Output: map[string]any{
						"results": []string{"report_q3.pdf", "data_q3.xlsx"},
					},
					Children: []Span{
						{
							ID:        "span-2-1",
							Name:      "google-search-api",
							ServiceName: "google",
							Type:      "service",
							StartTime: now + 100,
							EndTime:   now + 400,
							Status:    "success",
							Input: map[string]any{
								"q": "Q3 2024 financials site:sec.gov",
							},
							Output: map[string]any{
								"items": []map[string]any{
									{
										"title": "10-Q",
										"link":  "...",
									},
								},
							},
						},
					},
				},
				{
					ID:        "span-3",
					Name:      "data-analyzer",
					Type:      "tool",
					StartTime: now + 500,
					EndTime:   now + 1200,
					Status:    "success",
					Input: map[string]any{
						"files": []string{"data_q3.xlsx"},
					},
					Output: map[string]any{
						"analysis": "Growth detected",
						"metrics": map[string]any{
							"revenue": 1.15,
						},
					},
					Children: []Span{
						{
							ID:        "span-3-1",
							Name:      "python-interpreter",
							ServiceName: "local-python",
							Type:      "service",
							StartTime: now + 550,
							EndTime:   now + 1150,
							Status:    "success",
							Input: map[string]any{
								"code": "import pandas as pd\ndf = pd.read_excel('data_q3.xlsx')\nprint(df.revenue.sum())",
							},
							Output: map[string]any{
								"stdout": "115000000",
							},
						},
					},
				},
			},
		},
	}
}
