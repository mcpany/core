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

	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Initialize prepares the metrics system with a Prometheus sink.
// It sets up a global metrics collector that can be used throughout the application.
// The metrics are exposed on the /metrics endpoint.
var GlobalMetrics *metrics.Metrics

func Initialize() {
	// Create a Prometheus sink
	sink, err := prometheus.NewPrometheusSink()
	if err != nil {
		panic(err)
	}

	// Create a metrics configuration
	conf := metrics.DefaultConfig("mcpany")
	conf.EnableHostname = false

	// Initialize the metrics system
	m, err := metrics.New(conf, sink)
	if err != nil {
		panic(err)
	}
	GlobalMetrics = m
}

// Handler returns an http.Handler for the /metrics endpoint.
func Handler(h http.Handler) http.Handler {
	if h != nil {
		return h
	}
	return promhttp.Handler()
}

// SetGauge sets the value of a gauge.
func SetGauge(name string, val float32, labels ...string) {
	GlobalMetrics.SetGaugeWithLabels([]string{name}, val, []metrics.Label{
		{Name: "service_name", Value: labels[0]},
	})
}

// IncrCounter increments a counter.
func IncrCounter(name []string, val float32) {
	GlobalMetrics.IncrCounter(name, val)
}

// MeasureSince measures the time since a given start time and records it.
func MeasureSince(name []string, start time.Time) {
	GlobalMetrics.MeasureSince(name, start)
}
