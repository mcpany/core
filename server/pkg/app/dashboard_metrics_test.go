// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/topology"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestHandleDashboardMetrics_Trends(t *testing.T) {
	// Setup TopologyManager
	tm := topology.NewManager(nil, nil)
	defer tm.Close()

	// Seed TopologyManager with specific history to verify trends
	// We want to simulate:
	// Previous Window (10-5 mins ago): Low traffic
	// Current Window (5-0 mins ago): High traffic
	// This should result in an "up" trend for Requests.

	now := time.Now()
	points := []topology.TrafficPoint{}

	// Previous window: 5 points with 10 reqs each
	for i := 10; i > 5; i-- {
		ts := now.Add(time.Duration(-i) * time.Minute)
		points = append(points, topology.TrafficPoint{
			Time:    ts.Format("15:04"),
			Total:   10,
			Errors:  0,
			Latency: 50,
			Bytes:   100,
		})
	}

	// Current window: 5 points with 20 reqs each
	for i := 5; i > 0; i-- {
		ts := now.Add(time.Duration(-i) * time.Minute)
		points = append(points, topology.TrafficPoint{
			Time:    ts.Format("15:04"),
			Total:   20,
			Errors:  2,   // 10% error rate
			Latency: 100, // Higher latency
			Bytes:   200, // Higher bytes
		})
	}

	tm.SeedTrafficHistory(points)

	mockRegistry := new(MockServiceRegistryForDashboard)
	svc := configv1.UpstreamServiceConfig_builder{
		Id:   proto.String("test-service"),
		Name: proto.String("test-service"),
	}.Build()
	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{svc}, nil)

	app := &Application{
		ServiceRegistry: mockRegistry,
		TopologyManager: tm,
	}

	req := httptest.NewRequest(http.MethodGet, "/dashboard/metrics", nil)
	w := httptest.NewRecorder()

	handler := app.handleDashboardMetrics()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var metrics []Metric
	err := json.Unmarshal(w.Body.Bytes(), &metrics)
	assert.NoError(t, err)

	// Verify Metrics
	assert.NotEmpty(t, metrics)

	metricMap := make(map[string]Metric)
	for _, m := range metrics {
		metricMap[m.Label] = m
	}

	// Total Requests
	// Previous: 5 * 10 = 50
	// Current: 5 * 20 = 100
	// Change: (100 - 50) / 50 * 100 = +100%
	reqMetric := metricMap["Total Requests"]
	assert.Equal(t, "up", reqMetric.Trend)
	assert.Equal(t, "+100.0%", reqMetric.Change)

	// Est. Tokens
	// Previous: 5 * 100 = 500 bytes
	// Current: 5 * 200 = 1000 bytes
	// Change: +100%
	tokenMetric := metricMap["Est. Tokens"]
	assert.Equal(t, "up", tokenMetric.Trend)
	assert.Equal(t, "+100.0%", tokenMetric.Change)
	// Value should be total reqs * 250 = 150 * 250 = 37500
	assert.Equal(t, "37500", tokenMetric.Value)

	// Avg Latency
	// Previous Avg: 50ms
	// Current Avg: 100ms
	// Change: +100%
	latMetric := metricMap["Avg Latency"]
	assert.Equal(t, "up", latMetric.Trend) // Up is bad for latency, but Trend field is generic direction
	assert.Equal(t, "+100.0%", latMetric.Change)

	// Error Rate
	// Previous: 0 errors / 50 reqs = 0%
	// Current: 10 errors / 100 reqs = 10%
	// Change: +100% (from 0 to 0.1 is infinite increase, clamped to 100% in logic)
	errMetric := metricMap["Error Rate"]
	assert.Equal(t, "up", errMetric.Trend)
	assert.Equal(t, "+100.0%", errMetric.Change)
}
