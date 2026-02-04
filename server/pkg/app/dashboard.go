// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mcpany/core/server/pkg/auth"
)

// Metric represents a single dashboard metric to be displayed in the UI.
// It contains the label, value, trend direction, and other visual metadata.
type Metric struct {
	// Label is the primary text description of the metric (e.g., "Total Requests").
	Label string `json:"label"`
	// Value is the current value of the metric to display (e.g., "1,234").
	Value string `json:"value"`
	// Change represents the change from a previous period (e.g., "+10%").
	Change string `json:"change"`
	// Trend indicates the direction of the change ("up", "down", or "neutral").
	Trend string `json:"trend"` // "up" | "down" | "neutral"
	// Icon is the name of the icon to display with the metric (e.g., "Activity").
	Icon string `json:"icon"`
	// SubLabel provides additional context or subtitle for the metric (e.g., "Since startup").
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
		var throughput float64

		serviceID := r.URL.Query().Get("serviceId")

		if a.TopologyManager != nil {
			stats := a.TopologyManager.GetStats(serviceID)
			totalRequests = stats.TotalRequests
			avgLatency = stats.AvgLatency
			errorRate = stats.ErrorRate

			// Calculate throughput from history (last 60m)
			history := a.TopologyManager.GetTrafficHistory(serviceID)
			var totalInWindow int64
			for _, p := range history {
				totalInWindow += p.Total
			}
			if len(history) > 0 {
				// history is minutes. len(history) * 60 seconds
				throughput = float64(totalInWindow) / (float64(len(history)) * 60.0)
			}
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
				Label:    "Avg Throughput",
				Value:    fmt.Sprintf("%.2f rps", throughput),
				Change:   "--",
				Trend:    "neutral",
				Icon:     "Activity",
				SubLabel: "Last 60m",
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

func (a *Application) handleDashboardLayout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := auth.UserFromContext(ctx)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		layout, err := a.Storage.GetDashboardLayout(ctx, userID)
		if err != nil {
			http.Error(w, "Failed to get dashboard layout", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if layout == "" {
			layout = "[]"
		}
		_, _ = w.Write([]byte(layout))

	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if err := a.Storage.SaveDashboardLayout(ctx, userID, string(body)); err != nil {
			http.Error(w, "Failed to save dashboard layout", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
