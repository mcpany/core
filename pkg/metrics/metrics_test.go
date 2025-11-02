// Copyright 2024 Author(s)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMetrics(t *testing.T) (*metrics.Metrics, http.Handler) {
	t.Helper()
	reg := prom.NewRegistry()
	sink, err := prometheus.NewPrometheusSinkFrom(prometheus.PrometheusOpts{
		Registerer: reg,
	})
	require.NoError(t, err)

	conf := metrics.DefaultConfig("mcpany")
	conf.EnableHostname = false

	m, err := metrics.New(conf, sink)
	require.NoError(t, err)

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	return m, handler
}

func TestInitialize(t *testing.T) {
	// Initialize the metrics system
	Initialize()

	// Verify that the handler is not nil
	assert.NotNil(t, Handler(nil))
}

func TestHandler(t *testing.T) {
	// Get the handler
	handler := Handler(nil)

	// Verify that the handler is not nil
	assert.NotNil(t, handler)
}

func TestSetGauge(t *testing.T) {
	// Initialize the metrics system
	m, handler := newTestMetrics(t)
	originalMetrics := GetGlobalMetrics()
	SetGlobalMetrics(m)
	defer func() {
		SetGlobalMetrics(originalMetrics)
	}()

	// Set a gauge
	SetGauge("test_gauge", 123.45, "test_service")

	// Create a test recorder
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	handler.ServeHTTP(rr, req)

	// Check the output
	body := rr.Body.String()
	// Find the line containing the metric
	var line string
	for _, l := range strings.Split(body, "\n") {
		if strings.HasPrefix(l, "mcpany_test_gauge") {
			line = l
			break
		}
	}
	require.NotEmpty(t, line, "metric not found")

	// Parse the value
	var val float32
	_, err := fmt.Sscanf(line, "mcpany_test_gauge{service_name=\"test_service\"} %f", &val)
	require.NoError(t, err)

	// Check the value with a tolerance
	assert.InDelta(t, 123.45, val, 0.001)
}

func TestIncrCounter(t *testing.T) {
	m, handler := newTestMetrics(t)
	originalMetrics := GetGlobalMetrics()
	SetGlobalMetrics(m)
	defer func() {
		SetGlobalMetrics(originalMetrics)
	}()
	// Increment a counter
	IncrCounter([]string{"test_counter"}, 1)

	// Create a test recorder
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	handler.ServeHTTP(rr, req)

	// Check the output
	body := rr.Body.String()
	assert.Contains(t, body, "mcpany_test_counter 1")
}

func TestMeasureSince(t *testing.T) {
	m, handler := newTestMetrics(t)
	originalMetrics := GetGlobalMetrics()
	SetGlobalMetrics(m)
	defer func() {
		SetGlobalMetrics(originalMetrics)
	}()
	// Record a measurement
	start := time.Now()
	time.Sleep(10 * time.Millisecond)
	MeasureSince([]string{"test_measurement"}, start)

	// Create a test recorder
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	handler.ServeHTTP(rr, req)

	// Check the output
	body := rr.Body.String()
	assert.True(t, strings.Contains(body, "mcpany_test_measurement_sum"))
	assert.True(t, strings.Contains(body, "mcpany_test_measurement_count"))
}
