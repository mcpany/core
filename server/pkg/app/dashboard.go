// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

const (
	trendUp      = "up"
	trendDown    = "down"
	trendNeutral = "neutral"
)

type dashboardStats struct {
	reqChangeLabel   string
	reqTrend         string
	latChangeLabel   string
	latTrend         string
	errChangeLabel   string
	errTrend         string
	tokenChangeLabel string
	tokenTrend       string
}

// Helper to determine trend.
func getTrend(change float64, inverse bool) string {
	if change > 0 {
		if inverse {
			return trendUp // Bad (e.g. latency/error) but visually "up"
		}
		return trendUp
	} else if change < 0 {
		if inverse {
			return trendDown // Good (e.g. latency/error) but visually "down"
		}
		return trendDown
	}
	return trendNeutral
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

		var estTokens int64
		stats := dashboardStats{
			reqChangeLabel:   "--",
			reqTrend:         trendNeutral,
			latChangeLabel:   "--",
			latTrend:         trendNeutral,
			errChangeLabel:   "--",
			errTrend:         trendNeutral,
			tokenChangeLabel: "--",
			tokenTrend:       trendNeutral,
		}

		if a.TopologyManager != nil {
			mStats := a.TopologyManager.GetStats(serviceID)
			totalRequests = mStats.TotalRequests
			avgLatency = mStats.AvgLatency
			errorRate = mStats.ErrorRate

			// Calculate throughput from history (last 60m)
			history := a.TopologyManager.GetTrafficHistory(serviceID)
			var totalInWindow int64
			var totalBytes int64

			for _, p := range history {
				totalInWindow += p.Total
				totalBytes += p.Bytes
			}
			if len(history) > 0 {
				// history is minutes. len(history) * 60 seconds
				throughput = float64(totalInWindow) / (float64(len(history)) * 60.0)
			}
			estTokens = totalBytes / 4

			// Calculate Trends (Compare 2nd half vs 1st half of the window)
			if len(history) >= 2 {
				mid := len(history) / 2
				prevWindow := history[:mid]
				currWindow := history[mid:]

				var prevReqs, currReqs int64
				var prevLatSum, currLatSum int64
				var prevErrs, currErrs int64
				var prevBytes, currBytes int64

				for _, p := range prevWindow {
					prevReqs += p.Total
					prevLatSum += p.Latency * p.Total // Latency is avg per request
					prevErrs += p.Errors
					prevBytes += p.Bytes
				}
				for _, p := range currWindow {
					currReqs += p.Total
					currLatSum += p.Latency * p.Total
					currErrs += p.Errors
					currBytes += p.Bytes
				}

				// Request Trend
				var reqChange float64
				if prevReqs > 0 {
					reqChange = float64(currReqs-prevReqs) / float64(prevReqs) * 100
				} else if currReqs > 0 {
					reqChange = 100
				}
				stats.reqChangeLabel = fmt.Sprintf("%+.1f%%", reqChange)
				stats.reqTrend = getTrend(reqChange, false)

				// Latency Trend
				var prevAvgLat, currAvgLat, latChange float64
				if prevReqs > 0 {
					prevAvgLat = float64(prevLatSum) / float64(prevReqs)
				}
				if currReqs > 0 {
					currAvgLat = float64(currLatSum) / float64(currReqs)
				}
				if prevAvgLat > 0 {
					latChange = (currAvgLat - prevAvgLat) / prevAvgLat * 100
				} else if currAvgLat > 0 {
					latChange = 100
				}
				stats.latChangeLabel = fmt.Sprintf("%+.1f%%", latChange)
				stats.latTrend = getTrend(latChange, true)

				// Error Rate Trend
				var prevErrRate, currErrRate, errChange float64
				if prevReqs > 0 {
					prevErrRate = float64(prevErrs) / float64(prevReqs)
				}
				if currReqs > 0 {
					currErrRate = float64(currErrs) / float64(currReqs)
				}
				if prevErrRate > 0 {
					errChange = (currErrRate - prevErrRate) / prevErrRate * 100
				} else if currErrRate > 0 {
					errChange = 100
				}
				stats.errChangeLabel = fmt.Sprintf("%+.1f%%", errChange)
				stats.errTrend = getTrend(errChange, true)

				// Token Usage Trend
				var tokenChange float64
				if prevBytes > 0 {
					tokenChange = float64(currBytes-prevBytes) / float64(prevBytes) * 100
				} else if currBytes > 0 {
					tokenChange = 100
				}
				stats.tokenChangeLabel = fmt.Sprintf("%+.1f%%", tokenChange)
				stats.tokenTrend = getTrend(tokenChange, false)
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

		promptCount := 0
		if a.PromptManager != nil {
			promptCount = len(a.PromptManager.ListPrompts())
		}

		resourceCount := 0
		if a.ResourceManager != nil {
			resourceCount = len(a.ResourceManager.ListResources())
		}

		metrics := []Metric{
			{
				Label:    "Total Requests",
				Value:    fmt.Sprintf("%d", totalRequests),
				Change:   stats.reqChangeLabel,
				Trend:    stats.reqTrend,
				Icon:     "Activity",
				SubLabel: "Since startup",
			},
			{
				Label:    "Avg Throughput",
				Value:    fmt.Sprintf("%.2f rps", throughput),
				Change:   stats.reqChangeLabel, // Correlates with requests
				Trend:    stats.reqTrend,
				Icon:     "Activity",
				SubLabel: "Last 60m",
			},
			{
				Label:    "Active Services",
				Value:    fmt.Sprintf("%d", serviceCount),
				Change:   "--",
				Trend:    trendNeutral,
				Icon:     "Server",
				SubLabel: "Configured",
			},
			{
				Label:    "Connected Tools",
				Value:    fmt.Sprintf("%d", toolCount),
				Change:   "--",
				Trend:    trendNeutral,
				Icon:     "Zap",
				SubLabel: "Available",
			},
			{
				Label:    "Est. Tokens",
				Value:    fmt.Sprintf("%d", estTokens),
				Change:   stats.tokenChangeLabel,
				Trend:    stats.tokenTrend,
				Icon:     "Hash",
				SubLabel: "Last 60m",
			},
			{
				Label:    "Prompts",
				Value:    fmt.Sprintf("%d", promptCount),
				Change:   "--",
				Trend:    trendNeutral,
				Icon:     "MessageSquare",
				SubLabel: "Templates",
			},
			{
				Label:    "Resources",
				Value:    fmt.Sprintf("%d", resourceCount),
				Change:   "--",
				Trend:    trendNeutral,
				Icon:     "FileText",
				SubLabel: "Available",
			},
			{
				Label:    "Avg Latency",
				Value:    fmt.Sprintf("%dms", avgLatency.Milliseconds()),
				Change:   stats.latChangeLabel,
				Trend:    stats.latTrend,
				Icon:     "Clock",
				SubLabel: "Global Avg",
			},
			{
				Label:    "Error Rate",
				Value:    fmt.Sprintf("%.2f%%", errorRate*100),
				Change:   stats.errChangeLabel,
				Trend:    stats.errTrend,
				Icon:     "AlertCircle",
				SubLabel: "Since startup",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(metrics)
	}
}
