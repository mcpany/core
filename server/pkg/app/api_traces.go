package app

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
)

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
				"jsonrpc": "2.0",
				"id":      "req-123",
				"method":  "tools/call",
				"params": map[string]any{
					"name": "orchestrator-task",
					"arguments": map[string]any{
						"query": "Analyze Q3 financial report",
						"context": map[string]any{
							"user_id":  "usr_998",
							"session":  "user-session-123",
							"metadata": map[string]any{
								"device": "desktop",
								"locale": "en-US",
								"tags":   []string{"finance", "report", "q3"},
							},
						},
						"options": map[string]any{
							"deep_scan":   true,
							"max_results": 50,
						},
					},
				},
			},
			Output: map[string]any{
				"jsonrpc": "2.0",
				"id":      "req-123",
				"result": map[string]any{
					"content": []map[string]any{
						{
							"type": "text",
							"text": "Revenue up 15%",
						},
					},
					"isError": false,
					"summary": map[string]any{
						"confidence":          0.98,
						"sources":             []string{"report_q3.pdf", "data_q3.xlsx"},
						"computation_time_ms": 1250,
					},
				},
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
							ID:          "span-2-1",
							Name:        "google-search-api",
							ServiceName: "google",
							Type:        "service",
							StartTime:   now + 100,
							EndTime:     now + 400,
							Status:      "success",
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
					Status:    "error",
					ErrorMessage: "Failed to read data file: corrupted format",
					Input: map[string]any{
						"jsonrpc": "2.0",
						"id":      "req-124",
						"method":  "tools/call",
						"params": map[string]any{
							"name": "data-analyzer",
							"arguments": map[string]any{
								"files": []string{"data_q3.xlsx"},
								"config": map[string]any{
									"strict_mode": true,
								},
							},
						},
					},
					Output: map[string]any{
						"jsonrpc": "2.0",
						"id":      "req-124",
						"error": map[string]any{
							"code":    -32603,
							"message": "Internal error",
							"data": map[string]any{
								"details":   "Failed to read data file: corrupted format",
								"traceback": "Traceback (most recent call last):\n  File \"analyzer.py\", line 42, in process\n    raise ValueError('Corrupted format')",
							},
						},
					},
					Children: []Span{
						{
							ID:          "span-3-1",
							Name:        "python-interpreter",
							ServiceName: "local-python",
							Type:        "service",
							StartTime:   now + 550,
							EndTime:     now + 1150,
							Status:      "error",
							ErrorMessage: "ValueError: Corrupted format",
							Input: map[string]any{
								"jsonrpc": "2.0",
								"id":      "req-125",
								"method":  "tools/call",
								"params": map[string]any{
									"name": "python-interpreter",
									"arguments": map[string]any{
										"code": "import pandas as pd\ndf = pd.read_excel('data_q3.xlsx')\nprint(df.revenue.sum())",
									},
								},
							},
							Output: map[string]any{
								"jsonrpc": "2.0",
								"id":      "req-125",
								"error": map[string]any{
									"code":    1,
									"message": "Execution failed",
									"data": map[string]any{
										"stderr": "ValueError: Corrupted format",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
