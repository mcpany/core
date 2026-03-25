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
// Summary: Represents a metrics label for categorizing observations.
type Label = metrics.Label

// NewPrometheusSink creates a new Prometheus sink for metrics collection.
// Summary: Initializes a Prometheus-specific metrics sink.
// Parameters:
//   - None.
//
// Returns:
//   - *prometheus.PrometheusSink: The initialized Prometheus metrics sink.
//   - error: An error if the sink cannot be created.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewPrometheusSink() (*prometheus.PrometheusSink, error) {
	return prometheus.NewPrometheusSink()
}

var initOnce sync.Once

// Initialize prepares the metrics system with a Prometheus sink.
// Summary: Sets up the global metrics infrastructure, including a Prometheus sink and default configuration.
// Returns:
//   - error: An error if the global metrics system fails to initialize.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
// Summary: Retrieves the standard Prometheus HTTP handler for metrics scraping.
// Returns:
//   - http.Handler: An HTTP handler configured to serve Prometheus metrics.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func Handler() http.Handler {
	return promhttp.Handler()
}

// StartServer starts an HTTP server to expose the metrics.
// Summary: Launches a dedicated HTTP server for serving application metrics.
// Parameters:
//   - addr (string): The network address to listen on (e.g., ":8080").
//
// Returns:
//   - error: An error if the server fails to bind or start.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
// Summary: Records the current value of a gauge metric.
// Parameters:
//   - name (string): The name of the gauge.
//   - val (float32): The current value to be recorded.
//   - labels (...string): Optional service name to be used as a label.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
// Summary: Increments a counter metric by a specified value.
// Parameters:
//   - name ([]string): The hierarchical name of the counter.
//   - val (float32): The amount by which to increment the counter.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func IncrCounter(name []string, val float32) {
	metrics.IncrCounter(name, val)
}

// IncrCounterWithLabels increments a counter with labels.
// Summary: Increments a labeled counter metric by a specified value.
// Parameters:
//   - name ([]string): The hierarchical name of the counter.
//   - val (float32): The amount by which to increment the counter.
//   - labels ([]metrics.Label): The labels to apply to the observation.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func IncrCounterWithLabels(name []string, val float32, labels []metrics.Label) {
	metrics.IncrCounterWithLabels(name, val, labels)
}

// MeasureSince measures the time since a given start time and records it.
// Summary: Records the latency of an operation since the provided start time.
// Parameters:
//   - name ([]string): The hierarchical name of the latency metric.
//   - start (time.Time): The timestamp representing the beginning of the operation.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func MeasureSince(name []string, start time.Time) {
	metrics.MeasureSince(name, start)
}

// MeasureSinceWithLabels measures the time since a given start time and records it with labels.
// Summary: Records the latency of a labeled operation since the provided start time.
// Parameters:
//   - name ([]string): The hierarchical name of the latency metric.
//   - start (time.Time): The timestamp representing the beginning of the operation.
//   - labels ([]metrics.Label): The labels to apply to the latency observation.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func MeasureSinceWithLabels(name []string, start time.Time, labels []metrics.Label) {
	metrics.MeasureSinceWithLabels(name, start, labels)
}

// AddSample adds a sample to a histogram/summary.
// Summary: Records a single value as a sample for a histogram metric.
// Parameters:
//   - name ([]string): The hierarchical name of the histogram.
//   - val (float32): The value to be sampled.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func AddSample(name []string, val float32) {
	metrics.AddSample(name, val)
}

// AddSampleWithLabels adds a sample to a histogram/summary with labels.
// Summary: Records a single labeled value as a sample for a histogram metric.
// Parameters:
//   - name ([]string): The hierarchical name of the histogram.
//   - val (float32): The value to be sampled.
//   - labels ([]metrics.Label): The labels to apply to the histogram sample.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func AddSampleWithLabels(name []string, val float32, labels []metrics.Label) {
	metrics.AddSampleWithLabels(name, val, labels)
}
