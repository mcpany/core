// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewPrometheusSink creates a new Prometheus sink.
func NewPrometheusSink() (*prometheus.PrometheusSink, error) {
	return prometheus.NewPrometheusSink()
}

var initOnce sync.Once

// Initialize prepares the metrics system with a Prometheus sink.
// It sets up a global metrics collector that can be used throughout the application.
// The metrics are exposed on the /metrics endpoint.
func Initialize() error {
	var err error
	initOnce.Do(func() {
		// Create a Prometheus sink
		sink, err := NewPrometheusSink()
		if err != nil {
			return
		}

		// Create a metrics configuration
		conf := metrics.DefaultConfig("mcpany")
		conf.EnableHostname = false

		// Initialize the metrics system
		if _, err = metrics.NewGlobal(conf, sink); err != nil {
			return
		}
	})
	return err
}

// Handler returns an http.Handler for the /metrics endpoint.
func Handler() http.Handler {
	return promhttp.Handler()
}

// StartServer starts an HTTP server to expose the metrics.
func StartServer(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", Handler())
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}
	return server.ListenAndServe()
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
