// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/prometheus/client_golang/prometheus"
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

		mfs, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			logging.GetLogger().Error("failed to gather metrics", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		toolCounts := make(map[string]*ToolUsageStats)

		for _, mf := range mfs {
			if mf.GetName() == "mcpany_tools_call_total" {
				for _, m := range mf.GetMetric() {
					var toolName, serviceID string
					for _, label := range m.GetLabel() {
						if label.GetName() == "tool" {
							toolName = label.GetValue()
						}
						if label.GetName() == "service_id" {
							serviceID = label.GetValue()
						}
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
