// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Metric matches the UI interface.
type Metric struct {
	Label    string `json:"label"`
	Value    string `json:"value"`
	Change   string `json:"change"`
	Trend    string `json:"trend"` // "up" | "down" | "neutral"
	Icon     string `json:"icon"`
	SubLabel string `json:"subLabel"`
}

func (a *Application) handleDashboardMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Gather stats
		var totalRequests int64
		var avgLatency time.Duration
		var errorRate float64

		if a.TopologyManager != nil {
			stats := a.TopologyManager.GetStats()
			totalRequests = stats.TotalRequests
			avgLatency = stats.AvgLatency
			errorRate = stats.ErrorRate
		}

		var serviceCount int
		if a.ServiceRegistry != nil {
			services, _ := a.ServiceRegistry.GetAllServices()
			serviceCount = len(services)
		}

		toolCount := 0
		if a.ToolManager != nil {
			toolCount = len(a.ToolManager.ListTools())
		}

		resourceCount := 0
		if a.ResourceManager != nil {
			resourceCount = len(a.ResourceManager.ListResources())
		}

		promptCount := 0
		if a.PromptManager != nil {
			promptCount = len(a.PromptManager.ListPrompts())
		}

		metrics := []Metric{
			{
				Label:    "Total Requests",
				Value:    fmt.Sprintf("%d", totalRequests),
				Change:   "--", // Need history for change
				Trend:    "neutral",
				Icon:     "Activity",
				SubLabel: "Since startup",
			},
			{
				Label:    "Active Services",
				Value:    fmt.Sprintf("%d", serviceCount),
				Change:   "--",
				Trend:    "neutral",
				Icon:     "Server",
				SubLabel: "Configured",
			},
			{
				Label:    "Connected Tools",
				Value:    fmt.Sprintf("%d", toolCount),
				Change:   "--",
				Trend:    "neutral",
				Icon:     "Zap",
				SubLabel: "Available",
			},
			{
				Label:    "Resources",
				Value:    fmt.Sprintf("%d", resourceCount),
				Change:   "--",
				Trend:    "neutral",
				Icon:     "Database",
				SubLabel: "Managed",
			},
			{
				Label:    "Prompts",
				Value:    fmt.Sprintf("%d", promptCount),
				Change:   "--",
				Trend:    "neutral",
				Icon:     "MessageSquare",
				SubLabel: "Templates",
			},
			{
				Label:    "Avg Latency",
				Value:    fmt.Sprintf("%dms", avgLatency.Milliseconds()),
				Change:   "--",
				Trend:    "neutral",
				Icon:     "Clock",
				SubLabel: "Global Avg",
			},
			{
				Label:    "Error Rate",
				Value:    fmt.Sprintf("%.2f%%", errorRate*100),
				Change:   "--",
				Trend:    "neutral", // Could set based on threshold
				Icon:     "AlertCircle",
				SubLabel: "Since startup",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(metrics)
	}
}
