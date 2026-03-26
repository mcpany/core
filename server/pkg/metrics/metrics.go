// Copyright 2025 Author(s) of MCP Any.
// SPDX-License-Identifier: Apache-2.0.

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

// Label represents a key-value pair for labeling metrics.
//
// Summary: Represents a metrics label.
type Label = metrics.Label

// NewPrometheusSink creates a new Prometheus sink for metrics collection.
//
// Summary: Creates a new Prometheus sink.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - *prometheus.PrometheusSink: The initialized Prometheus sink.
//   - error: An error if the sink creation fails.
//
// Errors: - None.
//   - Returns an error if the sink creation fails.
//
// Side Effects: - None.
//   - None.
func NewPrometheusSink() (*prometheus.PrometheusSink, error) {
	return prometheus.NewPrometheusSink()
}

var initOnce sync.Once

// Initialize prepares the metrics system with a Prometheus sink.
//
// Summary: Initializes the global metrics collector.
//
// It sets up a global metrics collector that can be used throughout the application.
// The metrics are exposed on the /metrics endpoint.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - error: An error if the initialization fails.
//
// Errors: - None.
//   - Returns an error if the sink creation or global metrics initialization fails.
//
// Side Effects: - None.
//   - Sets a global metrics collector.
func Initialize() error {
	var err error
	initOnce.Do(func() {
		// Create a Prometheus sink.
		var sink *prometheus.PrometheusSink
		sink, err = NewPrometheusSink()
		if err != nil {
			return
		}

		// Create a metrics configuration.
		conf := metrics.DefaultConfig("mcpany")
		conf.EnableHostname = false

		// Initialize the metrics system.
		if _, err = metrics.NewGlobal(conf, sink); err != nil {
			return
		}
	})
	return err
}

// Handler returns an http.Handler for the /metrics endpoint.
//
// Summary: Retrieves the metrics HTTP handler.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - http.Handler: An http.Handler that serves the Prometheus metrics.
//
// Side Effects: - None.
//   - None.
func Handler() http.Handler {
	return promhttp.Handler()
}

// StartServer starts an HTTP server to expose the metrics.
//
// Summary: Starts the metrics server.
//
// Parameters: - None.
//   - addr: string. The address to listen on (e.g., ":8080").
//
// Returns: - None.
//   - error: An error if the server fails to start.
//
// Errors: - None.
//   - Returns an error if the server fails to start.
//
// Side Effects: - None.
//   - Starts an HTTP server.
func StartServer(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", Handler())

	var lc net.ListenConfig
	ln, err := lc.Listen(context.Background(), "tcp", addr)
	if err != nil {
		return err
	}

	if tcpAddr, ok := ln.Addr().(*net.TCPAddr); ok {
		// Log to stdout so E2E tests can parse the dynamically assigned port.
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

// SetGauge setGauge set gauge.
//
// Summary: SetGauge set gauge.
//
// Parameters:
//   - name (string): The name.
//   - val (float32): The val.
//   - labels (....string): The labels.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func SetGauge(name string, val float32, labels ....string) {
	var metricLabels []metrics.Label
	if len(labels) > 0 {
		metricLabels = []metrics.Label{
			{Name: "service_name", Value: labels[0]},
		}
	}
	metrics.SetGaugeWithLabels([]string{name}, val, metricLabels)
}

// IncrCounter incrCounter incr counter.
//
// Summary: IncrCounter incr counter.
//
// Parameters:
//   - name ([]string): The name.
//   - val (float32): The val.
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

// IncrCounterWithLabels incrCounterWithLabels incr counter with labels.
//
// Summary: IncrCounterWithLabels incr counter with labels.
//
// Parameters:
//   - name ([]string): The name.
//   - val (float32): The val.
//   - labels ([]metrics.Label): The labels.
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

// MeasureSince measureSince measure since.
//
// Summary: MeasureSince measure since.
//
// Parameters:
//   - name ([]string): The name.
//   - start (time.Time): The start.
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

// MeasureSinceWithLabels measureSinceWithLabels measure since with labels.
//
// Summary: MeasureSinceWithLabels measure since with labels.
//
// Parameters:
//   - name ([]string): The name.
//   - start (time.Time): The start.
//   - labels ([]metrics.Label): The labels.
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

// AddSample addSample add sample.
//
// Summary: AddSample add sample.
//
// Parameters:
//   - name ([]string): The name.
//   - val (float32): The val.
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

// AddSampleWithLabels addSampleWithLabels add sample with labels.
//
// Summary: AddSampleWithLabels add sample with labels.
//
// Parameters:
//   - name ([]string): The name.
//   - val (float32): The val.
//   - labels ([]metrics.Label): The labels.
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
