// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/topology"
)

// Metric represents a single dashboard metric to be displayed in the UI.
//
// Summary: Data structure for dashboard metrics.
//
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

// ToolUsageStats represents usage statistics for a tool.
type ToolUsageStats struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// ToolFailureStats represents failure statistics for a tool.
type ToolFailureStats struct {
	Name        string  `json:"name"`
	FailureRate float64 `json:"failureRate"` // Percentage 0-100
}

// ToolAnalytics represents detailed analytics for a tool.
type ToolAnalytics struct {
	Name        string  `json:"name"`
	TotalCalls  int64   `json:"totalCalls"`
	SuccessRate float64 `json:"successRate"`
}

const statusDegraded = "degraded"

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

func (a *Application) handleDashboardTraffic() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		serviceID := r.URL.Query().Get("serviceId")
		var history []topology.TrafficPoint
		if a.TopologyManager != nil {
			history = a.TopologyManager.GetTrafficHistory(serviceID)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(history)
	}
}

func (a *Application) handleDashboardTopTools() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		toolStats := a.aggregateToolStats()

		// Convert map to slice
		var stats []ToolUsageStats
		for name, data := range toolStats {
			stats = append(stats, ToolUsageStats{
				Name:  name,
				Count: data.Total,
			})
		}

		// Sort by count desc
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].Count > stats[j].Count
		})

		// Limit to top 5
		if len(stats) > 5 {
			stats = stats[:5]
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats)
	}
}

func (a *Application) handleDashboardToolFailures() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		toolStats := a.aggregateToolStats()

		var stats []ToolFailureStats
		for name, data := range toolStats {
			if data.Total == 0 {
				continue
			}
			failRate := float64(data.Errors) / float64(data.Total) * 100.0
			if failRate > 0 {
				stats = append(stats, ToolFailureStats{
					Name:        name,
					FailureRate: failRate,
				})
			}
		}

		// Sort by failure rate desc
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].FailureRate > stats[j].FailureRate
		})

		// Limit to top 5
		if len(stats) > 5 {
			stats = stats[:5]
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats)
	}
}

func (a *Application) handleDashboardToolUsage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		toolStats := a.aggregateToolStats()

		var stats []ToolAnalytics
		for name, data := range toolStats {
			successRate := 0.0
			if data.Total > 0 {
				successRate = float64(data.Total-data.Errors) / float64(data.Total) * 100.0
			}
			stats = append(stats, ToolAnalytics{
				Name:        name,
				TotalCalls:  data.Total,
				SuccessRate: successRate,
			})
		}

		// Sort by name
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].Name < stats[j].Name
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats)
	}
}

func (a *Application) handleDashboardHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		status := "ok"

		// Check config health
		if a.lastReloadErr != nil {
			status = statusDegraded
		}

		// Check services health
		if a.ServiceRegistry != nil {
			services, _ := a.ServiceRegistry.GetAllServices()
			for _, svc := range services {
				if _, ok := a.ServiceRegistry.GetServiceError(svc.GetName()); ok {
					status = statusDegraded
					break
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": status,
		})
	}
}

type aggregatedToolData struct {
	Total  int64
	Errors int64
}

// aggregateToolStats helper to gather metrics from Prometheus.
func (a *Application) aggregateToolStats() map[string]*aggregatedToolData {
	stats := make(map[string]*aggregatedToolData)

	mfs, err := a.MetricsGatherer.Gather()
	if err != nil {
		logging.GetLogger().Error("failed to gather metrics", "error", err)
		return stats
	}

	for _, mf := range mfs {
		if mf.GetName() == "mcpany_tools_call_total" {
			for _, m := range mf.GetMetric() {
				var toolName string
				var status string

				for _, label := range m.GetLabel() {
					if label.GetName() == "tool" {
						toolName = label.GetValue()
					}
					if label.GetName() == "status" {
						status = label.GetValue()
					}
				}

				if toolName == "" {
					continue
				}

				if _, ok := stats[toolName]; !ok {
					stats[toolName] = &aggregatedToolData{}
				}

				val := int64(m.GetCounter().GetValue())
				stats[toolName].Total += val
				if status == "error" {
					stats[toolName].Errors += val
				}
			}
		}
	}
	return stats
}

// handleDebugSeedTraffic handles POST /api/v1/debug/seed_traffic.
func (a *Application) handleDebugSeedTraffic() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var points []topology.TrafficPoint
		if err := json.NewDecoder(r.Body).Decode(&points); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		if a.TopologyManager != nil {
			a.TopologyManager.SeedTrafficHistory(points)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}
}
