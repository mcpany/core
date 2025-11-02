/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metrics

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/hashicorp/go-metrics/prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var globalMetrics *metrics.Metrics

// Init initializes the metrics system with a Prometheus sink.
// It sets up a global metrics instance that can be used throughout the application.
// The metrics are exposed on the /metrics endpoint.
func Init() (http.Handler, error) {
	promSink, err := prometheus.NewPrometheusSink()
	if err != nil {
		return nil, err
	}

	metricsConfig := metrics.DefaultConfig("mcpany")
	met, err := metrics.New(metricsConfig, promSink)
	if err != nil {
		return nil, err
	}
	globalMetrics = met

	return promhttp.HandlerFor(prom.DefaultGatherer, promhttp.HandlerOpts{}), nil
}

// Global returns the global metrics instance.
func Global() *metrics.Metrics {
	return globalMetrics
}

// MeasureSince is a helper function to measure the duration of an operation
// and record it as a sample.
func MeasureSince(key []string, start time.Time) {
	if globalMetrics == nil {
		return
	}
	globalMetrics.MeasureSince(key, start)
}

// SetGauge sets the value of a gauge.
func SetGauge(key []string, val float32) {
	if globalMetrics == nil {
		return
	}
	globalMetrics.SetGauge(key, val)
}

// IncrCounter increments a counter by a given value.
func IncrCounter(key []string, val float32) {
	if globalMetrics == nil {
		return
	}
	globalMetrics.IncrCounter(key, val)
}

// AddSample adds a sample to a metric.
func AddSample(key []string, val float32) {
	if globalMetrics == nil {
		return
	}
	globalMetrics.AddSample(key, val)
}

// TestOnlyReset resets the global metrics instance and the default prometheus
// registry. This is intended for use in tests only.
func TestOnlyReset() {
	globalMetrics = nil
	r := prom.NewRegistry()
	prom.DefaultRegisterer = r
	prom.DefaultGatherer = r
}
