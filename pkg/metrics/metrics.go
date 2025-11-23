// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
func Initialize() error {
	// Create a Prometheus sink
	sink, err := prometheus.NewPrometheusSink()
	if err != nil {
		return err
	}

	// Create a metrics configuration
	conf := metrics.DefaultConfig("mcpany")
	conf.EnableHostname = false

	// Initialize the metrics system
	if _, err := metrics.NewGlobal(conf, sink); err != nil {
		return err
	}

	return nil
}

// Handler returns an http.Handler for the /metrics endpoint.
func Handler() http.Handler {
	return promhttp.Handler()
}

// SetGauge sets the value of a gauge.
func SetGauge(name string, val float32, labels ...string) {
	metrics.SetGaugeWithLabels([]string{name}, val, []metrics.Label{
		{Name: "service_name", Value: labels[0]},
	})
}

// IncrCounter increments a counter.
func IncrCounter(name []string, val float32) {
	metrics.IncrCounter(name, val)
}

// MeasureSince measures the time since a given start time and records it.
func MeasureSince(name []string, start time.Time) {
	metrics.MeasureSince(name, start)
}
