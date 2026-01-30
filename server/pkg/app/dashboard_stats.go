// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/topology"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricToolsCallTotal = "mcpany_tools_call_total"
	labelTool            = "tool"
	labelServiceID       = "service_id"
	labelStatus          = "status"
)

// ToolUsageStats represents usage statistics for a tool.
type ToolUsageStats struct {
	Name      string `json:"name"`
	ServiceID string `json:"serviceId"`
	Count     int64  `json:"count"`
}

// handleDashboardTopTools returns the top used tools based on Prometheus metrics.
func (a *Application) handleDashboardTopTools() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		gatherer := a.MetricsGatherer
		if gatherer == nil {
			gatherer = prometheus.DefaultGatherer
		}

		mfs, err := gatherer.Gather()
		if err != nil {
			logging.GetLogger().Error("failed to gather metrics", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		filterServiceID := r.URL.Query().Get("serviceId")
		toolCounts := make(map[string]*ToolUsageStats)

		for _, mf := range mfs {
			if mf.GetName() == metricToolsCallTotal {
				for _, m := range mf.GetMetric() {
					var toolName, serviceID string
					for _, label := range m.GetLabel() {
						if label.GetName() == labelTool {
							toolName = label.GetValue()
						}
						if label.GetName() == labelServiceID {
							serviceID = label.GetValue()
						}
					}

					if filterServiceID != "" && serviceID != filterServiceID {
						continue
					}

					if toolName != "" {
						key := toolName + "@" + serviceID
						if _, exists := toolCounts[key]; !exists {
							toolCounts[key] = &ToolUsageStats{
								Name:      toolName,
								ServiceID: serviceID,
								Count:     0,
							}
						}
						toolCounts[key].Count += int64(m.GetCounter().GetValue())
					}
				}
			}
		}

		// Convert map to slice
		var stats []ToolUsageStats
		for _, s := range toolCounts {
			stats = append(stats, *s)
		}

		// Sort by count descending
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].Count > stats[j].Count
		})

		// Take top 10
		if len(stats) > 10 {
			stats = stats[:10]
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats)
	}
}

// handleDashboardTraffic returns the traffic history for the dashboard chart.
func (a *Application) handleDashboardTraffic() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if a.TopologyManager == nil {
			http.Error(w, "Topology manager not initialized", http.StatusServiceUnavailable)
			return
		}

		serviceID := r.URL.Query().Get("serviceId")
		points := a.TopologyManager.GetTrafficHistory(serviceID)

		// Transform to simple JSON if needed, or just return as is.
		// UI expects [{time: "00:00", total: 123}, ...]
		// TrafficPoint struct matches this.

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(points)
	}
}

// handleDebugSeedTraffic seeds the traffic history.
func (a *Application) handleDebugSeedTraffic() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if a.TopologyManager == nil {
			http.Error(w, "Topology manager not initialized", http.StatusServiceUnavailable)
			return
		}

		var points []topology.TrafficPoint
		if err := json.NewDecoder(r.Body).Decode(&points); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		a.TopologyManager.SeedTrafficHistory(points)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}
}

// ToolFailureStats represents failure statistics for a tool.
type ToolFailureStats struct {
	Name        string  `json:"name"`
	ServiceID   string  `json:"serviceId"`
	FailureRate float64 `json:"failureRate"` // Percentage 0-100
	TotalCalls  int64   `json:"totalCalls"`
}

// handleDashboardToolFailures returns the tools with highest failure rates based on Prometheus metrics.
func (a *Application) handleDashboardToolFailures() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		gatherer := a.MetricsGatherer
		if gatherer == nil {
			gatherer = prometheus.DefaultGatherer
		}

		mfs, err := gatherer.Gather()
		if err != nil {
			logging.GetLogger().Error("failed to gather metrics", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		type aggregatedStats struct {
			Name      string
			ServiceID string
			Success   int64
			Error     int64
		}

		filterServiceID := r.URL.Query().Get("serviceId")
		toolStats := make(map[string]*aggregatedStats)

		for _, mf := range mfs {
			if mf.GetName() == metricToolsCallTotal {
				for _, m := range mf.GetMetric() {
					var toolName, serviceID, status string
					for _, label := range m.GetLabel() {
						if label.GetName() == labelTool {
							toolName = label.GetValue()
						}
						if label.GetName() == labelServiceID {
							serviceID = label.GetValue()
						}
						if label.GetName() == labelStatus {
							status = label.GetValue()
						}
					}

					if filterServiceID != "" && serviceID != filterServiceID {
						continue
					}

					if toolName != "" {
						key := toolName + "@" + serviceID
						if _, exists := toolStats[key]; !exists {
							toolStats[key] = &aggregatedStats{
								Name:      toolName,
								ServiceID: serviceID,
							}
						}
						count := int64(m.GetCounter().GetValue())
						if status == "error" {
							toolStats[key].Error += count
						} else {
							toolStats[key].Success += count
						}
					}
				}
			}
		}

		// Convert map to slice of ToolFailureStats
		var stats []ToolFailureStats
		for _, s := range toolStats {
			total := s.Success + s.Error
			if total == 0 {
				continue
			}
			rate := (float64(s.Error) / float64(total)) * 100.0
			stats = append(stats, ToolFailureStats{
				Name:        s.Name,
				ServiceID:   s.ServiceID,
				FailureRate: rate,
				TotalCalls:  total,
			})
		}

		// Sort by FailureRate descending
		sort.Slice(stats, func(i, j int) bool {
			return stats[i].FailureRate > stats[j].FailureRate
		})

		// Take top 5
		if len(stats) > 5 {
			stats = stats[:5]
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats)
	}
}

// ToolAnalytics represents detailed usage analytics for a tool.
type ToolAnalytics struct {
	Name        string  `json:"name"`
	ServiceID   string  `json:"serviceId"`
	TotalCalls  int64   `json:"totalCalls"`
	SuccessRate float64 `json:"successRate"`
}

// handleDashboardToolUsage returns detailed usage statistics for all tools.
func (a *Application) handleDashboardToolUsage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		gatherer := a.MetricsGatherer
		if gatherer == nil {
			gatherer = prometheus.DefaultGatherer
		}

		mfs, err := gatherer.Gather()
		if err != nil {
			logging.GetLogger().Error("failed to gather metrics", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		type aggregatedStats struct {
			Name      string
			ServiceID string
			Success   int64
			Error     int64
		}

		filterServiceID := r.URL.Query().Get("serviceId")
		toolStats := make(map[string]*aggregatedStats)

		for _, mf := range mfs {
			if mf.GetName() == metricToolsCallTotal {
				for _, m := range mf.GetMetric() {
					var toolName, serviceID, status string
					for _, label := range m.GetLabel() {
						if label.GetName() == labelTool {
							toolName = label.GetValue()
						}
						if label.GetName() == labelServiceID {
							serviceID = label.GetValue()
						}
						if label.GetName() == labelStatus {
							status = label.GetValue()
						}
					}

					if filterServiceID != "" && serviceID != filterServiceID {
						continue
					}

					if toolName != "" {
						key := toolName + "@" + serviceID
						if _, exists := toolStats[key]; !exists {
							toolStats[key] = &aggregatedStats{
								Name:      toolName,
								ServiceID: serviceID,
							}
						}
						count := int64(m.GetCounter().GetValue())
						if status == "error" {
							toolStats[key].Error += count
						} else {
							toolStats[key].Success += count
						}
					}
				}
			}
		}

		var analytics []ToolAnalytics
		for _, s := range toolStats {
			total := s.Success + s.Error
			rate := 0.0
			if total > 0 {
				rate = (float64(s.Success) / float64(total)) * 100.0
			}
			analytics = append(analytics, ToolAnalytics{
				Name:        s.Name,
				ServiceID:   s.ServiceID,
				TotalCalls:  total,
				SuccessRate: rate,
			})
		}

		// Sort by Name
		sort.Slice(analytics, func(i, j int) bool {
			return analytics[i].Name < analytics[j].Name
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(analytics)
	}
}
