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

// NewPrometheusSink - Auto-generated documentation.
//
// Summary: NewPrometheusSink creates a new Prometheus sink for metrics collection.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func NewPrometheusSink() (*prometheus.PrometheusSink, error) {
	return prometheus.NewPrometheusSink()
}

var initOnce sync.Once

// Initialize - Auto-generated documentation.
//
// Summary: Initialize prepares the metrics system with a Prometheus sink.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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

// Handler - Auto-generated documentation.
//
// Summary: Handler returns an http.Handler for the /metrics endpoint.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
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

// SetGauge - Auto-generated documentation.
//
// Summary: SetGauge sets the value of a gauge.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func SetGauge(name string, val float32, labels ...string) {
	var metricLabels []metrics.Label
	if len(labels) > 0 {
		metricLabels = []metrics.Label{
			{Name: "service_name", Value: labels[0]},
		}
	}
	metrics.SetGaugeWithLabels([]string{name}, val, metricLabels)
}

// IncrCounter - Auto-generated documentation.
//
// Summary: IncrCounter increments a counter.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func IncrCounter(name []string, val float32) {
	metrics.IncrCounter(name, val)
}

// IncrCounterWithLabels - Auto-generated documentation.
//
// Summary: IncrCounterWithLabels increments a counter with labels.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func IncrCounterWithLabels(name []string, val float32, labels []metrics.Label) {
	metrics.IncrCounterWithLabels(name, val, labels)
}

// MeasureSince - Auto-generated documentation.
//
// Summary: MeasureSince measures the time since a given start time and records it.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func MeasureSince(name []string, start time.Time) {
	metrics.MeasureSince(name, start)
}

// MeasureSinceWithLabels - Auto-generated documentation.
//
// Summary: MeasureSinceWithLabels measures the time since a given start time and records it with labels.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func MeasureSinceWithLabels(name []string, start time.Time, labels []metrics.Label) {
	metrics.MeasureSinceWithLabels(name, start, labels)
}

// AddSample - Auto-generated documentation.
//
// Summary: AddSample adds a sample to a histogram/summary.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func AddSample(name []string, val float32) {
	metrics.AddSample(name, val)
}

// AddSampleWithLabels - Auto-generated documentation.
//
// Summary: AddSampleWithLabels adds a sample to a histogram/summary with labels.
//
// Parameters:
//   - args: Variable arguments.
//
// Returns:
//   - result: The result of the operation.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - May modify internal state or perform external calls.
func AddSampleWithLabels(name []string, val float32, labels []metrics.Label) {
	metrics.AddSampleWithLabels(name, val, labels)
}
