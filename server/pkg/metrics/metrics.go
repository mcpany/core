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

// NewPrometheusSink creates a new Prometheus sink for metrics collection. Summary: Creates a Prometheus sink. Returns: - *prometheus.PrometheusSink: The initialized Prometheus sink. - error: An error if the sink creation fails.
//
// Summary: NewPrometheusSink creates a new Prometheus sink for metrics collection. Summary: Creates a Prometheus sink. Returns: - *prometheus.PrometheusSink: The initialized Prometheus sink. - error: An error if the sink creation fails.
//
// Parameters:
//   - None.
//
// Returns:
//   - (*prometheus.PrometheusSink): The resulting prometheus.PrometheusSink object containing the requested data.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
func NewPrometheusSink() (*prometheus.PrometheusSink, error) {
	return prometheus.NewPrometheusSink()
}

var initOnce sync.Once

// Initialize prepares the metrics system with a Prometheus sink. Summary: Initializes the global metrics collector. It sets up a global metrics collector that can be used throughout the application. The metrics are exposed on the /metrics endpoint. Returns: - error: An error if the initialization fails.
//
// Summary: Initialize prepares the metrics system with a Prometheus sink. Summary: Initializes the global metrics collector. It sets up a global metrics collector that can be used throughout the application. The metrics are exposed on the /metrics endpoint. Returns: - error: An error if the initialization fails.
//
// Parameters:
//   - None.
//
// Returns:
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
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

// Handler returns an http.Handler for the /metrics endpoint. Summary: Retrieves the metrics HTTP handler. Returns: - http.Handler: An http.Handler that serves the Prometheus metrics.
//
// Summary: Handler returns an http.Handler for the /metrics endpoint. Summary: Retrieves the metrics HTTP handler. Returns: - http.Handler: An http.Handler that serves the Prometheus metrics.
//
// Parameters:
//   - None.
//
// Returns:
//   - (http.Handler): The resulting http.Handler object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func Handler() http.Handler {
	return promhttp.Handler()
}

// StartServer starts an HTTP server to expose the metrics. Summary: Starts the metrics server. Parameters: - addr: string. The address to listen on (e.g., ":8080"). Returns: - error: An error if the server fails to start.
//
// Summary: StartServer starts an HTTP server to expose the metrics. Summary: Starts the metrics server. Parameters: - addr: string. The address to listen on (e.g., ":8080"). Returns: - error: An error if the server fails to start.
//
// Parameters:
//   - addr (string): The addr parameter used in the operation.
//
// Returns:
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - Modifies global state, writes to the database, or establishes network connections.
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

// SetGauge sets the value of a gauge. Summary: Sets a gauge metric. Parameters: - name: string. The name of the gauge. - val: float32. The value to set. - labels: ...string. A list of labels to apply to the gauge.
//
// Summary: SetGauge sets the value of a gauge. Summary: Sets a gauge metric. Parameters: - name: string. The name of the gauge. - val: float32. The value to set. - labels: ...string. A list of labels to apply to the gauge.
//
// Parameters:
//   - name (string): The name parameter used in the operation.
//   - val (float32): The val parameter used in the operation.
//   - labels (...string): The labels parameter used in the operation.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Modifies global state, writes to the database, or establishes network connections.
func SetGauge(name string, val float32, labels ...string) {
	var metricLabels []metrics.Label
	if len(labels) > 0 {
		metricLabels = []metrics.Label{
			{Name: "service_name", Value: labels[0]},
		}
	}
	metrics.SetGaugeWithLabels([]string{name}, val, metricLabels)
}

// IncrCounter increments a counter. Summary: Increments a counter metric. Parameters: - name: []string. The name of the counter (as a path). - val: float32. The amount to increment.
//
// Summary: IncrCounter increments a counter. Summary: Increments a counter metric. Parameters: - name: []string. The name of the counter (as a path). - val: float32. The amount to increment.
//
// Parameters:
//   - name ([]string): The name parameter used in the operation.
//   - val (float32): The val parameter used in the operation.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func IncrCounter(name []string, val float32) {
	metrics.IncrCounter(name, val)
}

// IncrCounterWithLabels increments a counter with labels. Summary: Increments a labeled counter metric. Parameters: - name: []string. The name of the counter (as a path). - val: float32. The amount to increment. - labels: []metrics.Label. The labels to apply.
//
// Summary: IncrCounterWithLabels increments a counter with labels. Summary: Increments a labeled counter metric. Parameters: - name: []string. The name of the counter (as a path). - val: float32. The amount to increment. - labels: []metrics.Label. The labels to apply.
//
// Parameters:
//   - name ([]string): The name parameter used in the operation.
//   - val (float32): The val parameter used in the operation.
//   - labels ([]metrics.Label): The labels parameter used in the operation.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func IncrCounterWithLabels(name []string, val float32, labels []metrics.Label) {
	metrics.IncrCounterWithLabels(name, val, labels)
}

// MeasureSince measures the time since a given start time and records it. Summary: Records latency metric. Parameters: - name: []string. The name of the metric (as a path). - start: time.Time. The start time.
//
// Summary: MeasureSince measures the time since a given start time and records it. Summary: Records latency metric. Parameters: - name: []string. The name of the metric (as a path). - start: time.Time. The start time.
//
// Parameters:
//   - name ([]string): The name parameter used in the operation.
//   - start (time.Time): The start parameter used in the operation.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func MeasureSince(name []string, start time.Time) {
	metrics.MeasureSince(name, start)
}

// MeasureSinceWithLabels measures the time since a given start time and records it with labels. Summary: Records labeled latency metric. Parameters: - name: []string. The name of the metric (as a path). - start: time.Time. The start time. - labels: []metrics.Label. The labels to apply.
//
// Summary: MeasureSinceWithLabels measures the time since a given start time and records it with labels. Summary: Records labeled latency metric. Parameters: - name: []string. The name of the metric (as a path). - start: time.Time. The start time. - labels: []metrics.Label. The labels to apply.
//
// Parameters:
//   - name ([]string): The name parameter used in the operation.
//   - start (time.Time): The start parameter used in the operation.
//   - labels ([]metrics.Label): The labels parameter used in the operation.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func MeasureSinceWithLabels(name []string, start time.Time, labels []metrics.Label) {
	metrics.MeasureSinceWithLabels(name, start, labels)
}

// AddSample adds a sample to a histogram/summary. Summary: Adds a sample to a metric. Parameters: - name: []string. The name of the metric (as a path). - val: float32. The value to sample.
//
// Summary: AddSample adds a sample to a histogram/summary. Summary: Adds a sample to a metric. Parameters: - name: []string. The name of the metric (as a path). - val: float32. The value to sample.
//
// Parameters:
//   - name ([]string): The name parameter used in the operation.
//   - val (float32): The val parameter used in the operation.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func AddSample(name []string, val float32) {
	metrics.AddSample(name, val)
}

// AddSampleWithLabels adds a sample to a histogram/summary with labels. Summary: Adds a labeled sample to a metric. Parameters: - name: []string. The name of the metric (as a path). - val: float32. The value to sample. - labels: []metrics.Label. The labels to apply.
//
// Summary: AddSampleWithLabels adds a sample to a histogram/summary with labels. Summary: Adds a labeled sample to a metric. Parameters: - name: []string. The name of the metric (as a path). - val: float32. The value to sample. - labels: []metrics.Label. The labels to apply.
//
// Parameters:
//   - name ([]string): The name parameter used in the operation.
//   - val (float32): The val parameter used in the operation.
//   - labels ([]metrics.Label): The labels parameter used in the operation.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func AddSampleWithLabels(name []string, val float32, labels []metrics.Label) {
	metrics.AddSampleWithLabels(name, val, labels)
}
