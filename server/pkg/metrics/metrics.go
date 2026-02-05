// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package metrics provides utilities for collecting and exposing application metrics.
package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Label is an alias for metrics.Label. It represents a key-value pair for labeling metrics.
type Label = metrics.Label

// NewPrometheusSink creates a new Prometheus sink for metrics collection.
//
// Summary: Creates a new Prometheus sink for metrics collection.
//
// Returns:
//   - *prometheus.PrometheusSink: A pointer to a prometheus.PrometheusSink.
//   - error: An error if the sink creation fails.
func NewPrometheusSink() (*prometheus.PrometheusSink, error) {
	return prometheus.NewPrometheusSink()
}

var initOnce sync.Once

// Initialize prepares the metrics system with a Prometheus sink.
// It sets up a global metrics collector that can be used throughout the application.
// The metrics are exposed on the /metrics endpoint.
//
// Summary: Initializes the global metrics system.
//
// Returns:
//   - error: An error if the initialization fails.
func Initialize() error {
	var err error
	initOnce.Do(func() {
		// Create a Prometheus sink
		var sink *prometheus.PrometheusSink
		sink, err = NewPrometheusSink()
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
//
// Summary: Returns an HTTP handler for the metrics endpoint.
//
// Returns:
//   - http.Handler: An http.Handler that serves the Prometheus metrics.
func Handler() http.Handler {
	return promhttp.Handler()
}

// StartServer starts an HTTP server to expose the metrics.
//
// Summary: Starts an HTTP server to expose metrics.
//
// Parameters:
//   - addr: string. The address to listen on (e.g., ":8080").
//
// Returns:
//   - error: An error if the server fails to start.
func StartServer(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", Handler())

	var lc net.ListenConfig
	ln, err := lc.Listen(context.Background(), "tcp", addr)
	if err != nil {
		return err
	}

	if tcpAddr, ok := ln.Addr().(*net.TCPAddr); ok {
		// Log to stdout so E2E tests can parse the dynamically assigned port
		fmt.Printf("Metrics server listening on port %d\n", tcpAddr.Port)
	}

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
	}
	return server.Serve(ln)
}

// SetGauge sets the value of a gauge.
//
// Summary: Sets the value of a gauge metric.
//
// Parameters:
//   - name: string. The name of the gauge.
//   - val: float32. The value to set.
//   - labels: ...string. A list of labels to apply to the gauge.
func SetGauge(name string, val float32, labels ...string) {
	var metricLabels []metrics.Label
	if len(labels) > 0 {
		metricLabels = []metrics.Label{
			{Name: "service_name", Value: labels[0]},
		}
	}
	metrics.SetGaugeWithLabels([]string{name}, val, metricLabels)
}

// IncrCounter increments a counter.
//
// Summary: Increments a counter metric.
//
// Parameters:
//   - name: []string. The name of the counter (as a path).
//   - val: float32. The amount to increment.
func IncrCounter(name []string, val float32) {
	metrics.IncrCounter(name, val)
}

// IncrCounterWithLabels increments a counter with labels.
//
// Summary: Increments a counter metric with labels.
//
// Parameters:
//   - name: []string. The name of the counter (as a path).
//   - val: float32. The amount to increment.
//   - labels: []metrics.Label. The labels to apply.
func IncrCounterWithLabels(name []string, val float32, labels []metrics.Label) {
	metrics.IncrCounterWithLabels(name, val, labels)
}

// MeasureSince measures the time since a given start time and records it.
//
// Summary: Measures and records the duration since the start time.
//
// Parameters:
//   - name: []string. The name of the metric (as a path).
//   - start: time.Time. The start time.
func MeasureSince(name []string, start time.Time) {
	metrics.MeasureSince(name, start)
}

// MeasureSinceWithLabels measures the time since a given start time and records it with labels.
//
// Summary: Measures and records the duration since the start time with labels.
//
// Parameters:
//   - name: []string. The name of the metric (as a path).
//   - start: time.Time. The start time.
//   - labels: []metrics.Label. The labels to apply.
func MeasureSinceWithLabels(name []string, start time.Time, labels []metrics.Label) {
	metrics.MeasureSinceWithLabels(name, start, labels)
}

// AddSample adds a sample to a histogram/summary.
//
// Summary: Adds a sample to a histogram or summary metric.
//
// Parameters:
//   - name: []string. The name of the metric (as a path).
//   - val: float32. The value to sample.
func AddSample(name []string, val float32) {
	metrics.AddSample(name, val)
}

// AddSampleWithLabels adds a sample to a histogram/summary with labels.
//
// Summary: Adds a sample to a histogram or summary metric with labels.
//
// Parameters:
//   - name: []string. The name of the metric (as a path).
//   - val: float32. The value to sample.
//   - labels: []metrics.Label. The labels to apply.
func AddSampleWithLabels(name []string, val float32, labels []metrics.Label) {
	metrics.AddSampleWithLabels(name, val, labels)
}
