// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/topology"
	"github.com/prometheus/client_golang/prometheus"
)

// getStatsCache returns cached data if valid.
func (a *Application) getStatsCache(key string) (any, bool) {
	a.statsCacheMu.RLock()
	defer a.statsCacheMu.RUnlock()
	entry, ok := a.statsCache[key]
	if !ok || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry.Data, true
}

// setStatsCache sets the cache with 5s TTL.
func (a *Application) setStatsCache(key string, data any) {
	a.statsCacheMu.Lock()
	defer a.statsCacheMu.Unlock()

	// ⚡ Bolt Security: Prevent DoS by bounding cache size.
	if len(a.statsCache) >= 100 {
		// ⚡ Bolt Optimization: Randomized Eviction
		// Instead of clearing the entire cache (which causes a thundering herd),
		// we evict a random subset (approx 25%) to make room.
		// Randomized Selection from Top 5 High-Impact Targets

		toEvict := len(a.statsCache) / 4
		for k := range a.statsCache {
			delete(a.statsCache, k)
			toEvict--
			if toEvict <= 0 {
				break
			}
		}
	}

	a.statsCache[key] = statsCacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(5 * time.Second),
	}
}

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

		filterServiceID := r.URL.Query().Get("serviceId")
		cacheKey := "dashboard_top_tools:" + filterServiceID

		// ⚡ Bolt Optimization: Check cache
		if data, ok := a.getStatsCache(cacheKey); ok {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(data)
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

		// ⚡ Bolt Optimization: Update cache
		a.setStatsCache(cacheKey, stats)

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

		filterServiceID := r.URL.Query().Get("serviceId")
		cacheKey := "dashboard_tool_failures:" + filterServiceID

		// ⚡ Bolt Optimization: Check cache
		if data, ok := a.getStatsCache(cacheKey); ok {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(data)
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

		// ⚡ Bolt Optimization: Update cache
		a.setStatsCache(cacheKey, stats)

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

		filterServiceID := r.URL.Query().Get("serviceId")
		cacheKey := "dashboard_tool_usage:" + filterServiceID

		// ⚡ Bolt Optimization: Check cache
		if data, ok := a.getStatsCache(cacheKey); ok {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(data)
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

		// ⚡ Bolt Optimization: Update cache
		a.setStatsCache(cacheKey, analytics)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(analytics)
	}
}

// ServiceHealthResponse represents the response for the health dashboard.
type ServiceHealthResponse struct {
	Services []ServiceHealth                 `json:"services"`
	History  map[string][]health.HistoryPoint `json:"history"`
}

// ServiceHealth represents the health status of a service.
type ServiceHealth struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Latency string `json:"latency"`
	Uptime  string `json:"uptime"`
	Message string `json:"message,omitempty"`
}

// handleDashboardHealth returns the health status and history for all services.
func (a *Application) handleDashboardHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if a.ServiceRegistry == nil {
			http.Error(w, "Service Registry not initialized", http.StatusServiceUnavailable)
			return
		}

		services, err := a.ServiceRegistry.GetAllServices()
		if err != nil {
			logging.GetLogger().Error("failed to list services", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		history := health.GetHealthHistory()
		var serviceHealths []ServiceHealth

		const (
			statusInactive  = "inactive"
			statusUnhealthy = "unhealthy"
			statusUnknown   = "unknown"
			statusHealthy   = "healthy"
			statusDegraded  = "degraded"
		)

		for _, svc := range services {
			name := svc.GetName()
			hPoints := history[name]
			var status string

			switch {
			case len(hPoints) > 0:
				status = hPoints[len(hPoints)-1].Status
			case svc.GetDisable():
				status = statusInactive
			default:
				// If no history but enabled, assume pending or unknown?
				// Or check if we have an error in registry
				if _, ok := a.ServiceRegistry.GetServiceError(svc.GetId()); ok {
					status = statusUnhealthy
				} else {
					status = statusUnknown
				}
			}

			// Map health package status to UI status
			// health package uses "up", "down"?
			// health.go uses health.AvailabilityStatus which is "up", "down".
			// UI expects "healthy", "unhealthy", "degraded", "inactive".
			var uiStatus string
			switch status {
			case "up", "UP":
				uiStatus = statusHealthy
			case "down", "DOWN":
				uiStatus = statusUnhealthy
			case statusInactive:
				uiStatus = statusInactive
			default:
				uiStatus = statusUnknown
			}

			// Get error message if any
			var msg string
			if errMsg, ok := a.ServiceRegistry.GetServiceError(svc.GetId()); ok {
				msg = errMsg
				if uiStatus == statusHealthy {
					uiStatus = statusDegraded // If up but has error (maybe partial?)
				}
			}

			// Calculate real stats
			latencyStr := "0ms"
			if a.TopologyManager != nil {
				avgLat, _ := a.TopologyManager.GetRecentServiceStats(svc.GetId(), 15*time.Minute)
				if avgLat > 0 {
					latencyStr = avgLat.String()
				}
			}

			uptimeStr := calculateUptime(hPoints, 24*time.Hour)

			serviceHealths = append(serviceHealths, ServiceHealth{
				ID:      svc.GetId(),
				Name:    name,
				Status:  uiStatus,
				Latency: latencyStr,
				Uptime:  uptimeStr,
				Message: msg,
			})
		}

		// Remap history to be keyed by Service ID for the frontend
		historyByID := make(map[string][]health.HistoryPoint)
		for _, svc := range services {
			if h, ok := history[svc.GetName()]; ok {
				historyByID[svc.GetId()] = h
			}
		}

		resp := ServiceHealthResponse{
			Services: serviceHealths,
			History:  historyByID,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func calculateUptime(points []health.HistoryPoint, window time.Duration) string {
	if len(points) == 0 {
		return "100%" // Default to 100% if no history (optimistic)
	}

	now := time.Now().UnixMilli()
	startWindow := now - window.Milliseconds()

	// Filter points within window, plus the one immediately before (to know start state)
	var relevantPoints []health.HistoryPoint

	// Find the index of the first point inside the window
	firstInside := -1
	for i, p := range points {
		if p.Timestamp >= startWindow {
			firstInside = i
			break
		}
	}

	if firstInside == -1 {
		// All points are older than window. Status is whatever the last point was.
		lastStatus := points[len(points)-1].Status
		if isHealthy(lastStatus) {
			return "100%"
		}
		return "0%"
	}

	// Add the point immediately before window start if it exists, to establish initial state
	if firstInside > 0 {
		relevantPoints = append(relevantPoints, points[firstInside-1])
	} else {
		// If first point is inside window, we assume previous state was same as first.
		relevantPoints = append(relevantPoints, health.HistoryPoint{Timestamp: startWindow, Status: points[firstInside].Status})
	}

	relevantPoints = append(relevantPoints, points[firstInside:]...)

	var upTimeMillis int64

	for i := 0; i < len(relevantPoints); i++ {
		current := relevantPoints[i]

		// Determine end of this segment
		var endMillis int64
		if i == len(relevantPoints)-1 {
			endMillis = now
		} else {
			endMillis = relevantPoints[i+1].Timestamp
		}

		// Clip start to window start
		startMillis := current.Timestamp
		if startMillis < startWindow {
			startMillis = startWindow
		}

		if endMillis > now {
			endMillis = now
		}

		if endMillis > startMillis {
			if isHealthy(current.Status) {
				upTimeMillis += (endMillis - startMillis)
			}
		}
	}

	uptimePercent := (float64(upTimeMillis) / float64(window.Milliseconds())) * 100.0
	if uptimePercent > 100 {
		uptimePercent = 100
	}

	if uptimePercent == 100 {
		return "100%"
	}
	if uptimePercent == 0 {
		return "0%"
	}
	return fmt.Sprintf("%.1f%%", uptimePercent)
}

func isHealthy(status string) bool {
	return status == "up" || status == "UP" || status == "healthy" || status == "HEALTHY"
}
