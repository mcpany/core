// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkHandleDashboardTopTools(b *testing.B) {
	// Define counters
	toolsCallTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcpany_tools_call_total",
			Help: "Total number of tool calls",
		},
		[]string{"tool", "service_id", "status", "error_type"},
	)

	registry := prometheus.NewRegistry()
	_ = registry.Register(toolsCallTotal)

	// Populate with 10,000 distinct tools
	for i := 0; i < 10000; i++ {
		toolName := fmt.Sprintf("bench_tool_%d", i)
		count := float64(i % 100) // Random-ish count
		toolsCallTotal.WithLabelValues(toolName, "bench_service", "success", "").Add(count)
	}

	app := &Application{
		MetricsGatherer: registry,
		statsCache: make(map[string]statsCacheEntry),
	}

	handler := app.handleDashboardTopTools()
	req, _ := http.NewRequest("GET", "/dashboard/top-tools", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear cache to force calculation
		app.statsCache = make(map[string]statsCacheEntry)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}
