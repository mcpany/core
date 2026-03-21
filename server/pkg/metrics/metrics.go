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

	armometrics "github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Label is an alias for armometrics.Label. It represents a key-value pair for labeling metrics.
//
// Summary: Represents a Label.
type Label = armometrics.Label

// NewPrometheusSink creates a new Prometheus sink for metrics collection.
//
// Summary: Creates a Prometheus sink.
//
// Parameters:
//   - None.
//
// Returns:
//   - *prometheus.PrometheusSink: The initialized Prometheus sink.
//   - error: An error if the sink creation fails.
//
// Side Effects:
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
		conf := armometrics.DefaultConfig("mcpany")
		conf.EnableHostname = false

		// Initialize the metrics system
		if _, err = armometrics.NewGlobal(conf, sink); err != nil {
			return
		}
	})
	return err
}

// Handler returns an http.Handler for the /metrics endpoint.
//
// Summary: Retrieves the metrics HTTP handler.
//
// Returns:
//   - http.Handler: An http.Handler that serves the Prometheus metrics.
func Handler() http.Handler {
	return promhttp.Handler()
}

// StartServer starts an HTTP server to expose the metrics.
//
// Summary: Starts the metrics server.
//
// Parameters:
//   - addr (string): The address to listen on (e.g., ":8080").
//
// Returns:
//   - error: An error if the server fails to start.
//
// Side Effects:
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
// Summary: Sets a gauge metric.
//
// Parameters:
//   - name (string): The name of the gauge.
//   - val (float32): The value to set.
//   - labels (...string): A list of labels to apply to the gauge.
//
// Returns:
//   - None.
//
// Side Effects:
//   - Updates a global gauge metric.
func SetGauge(name string, val float32, labels ...string) {
	var metricLabels []armometrics.Label
	if len(labels) > 0 {
		metricLabels = []armometrics.Label{
			{Name: "service_name", Value: labels[0]},
		}
	}
	armometrics.SetGaugeWithLabels([]string{name}, val, metricLabels)
}

// IncrCounter increments a counter.
//
// Summary: Increments a counter metric.
//
// Parameters:
//   - name ([]string): The name of the counter (as a path).
//   - val (float32): The amount to increment.
//
// Returns:
//   - None.
//
// Side Effects:
//   - Increments a global counter metric.
func IncrCounter(name []string, val float32) {
	armometrics.IncrCounter(name, val)
}

// IncrCounterWithLabels increments a counter with labels.
//
// Summary: Increments a labeled counter metric.
//
// Parameters:
//   - name ([]string): The name of the counter (as a path).
//   - val (float32): The amount to increment.
//   - labels ([]armometrics.Label): The labels to apply.
//
// Returns:
//   - None.
//
// Side Effects:
//   - Increments a global labeled counter metric.
func IncrCounterWithLabels(name []string, val float32, labels []armometrics.Label) {
	armometrics.IncrCounterWithLabels(name, val, labels)
}

// MeasureSince measures the time since a given start time and records it.
//
// Summary: Records latency metric.
//
// Parameters:
//   - name ([]string): The name of the metric (as a path).
//   - start (time.Time): The start time.
//
// Returns:
//   - None.
//
// Side Effects:
//   - Records a latency sample in the global metrics collector.
func MeasureSince(name []string, start time.Time) {
	armometrics.MeasureSince(name, start)
}

// MeasureSinceWithLabels measures the time since a given start time and records it with labels.
//
// Summary: Records labeled latency metric.
//
// Parameters:
//   - name ([]string): The name of the metric (as a path).
//   - start (time.Time): The start time.
//   - labels ([]armometrics.Label): The labels to apply.
//
// Returns:
//   - None.
//
// Side Effects:
//   - Records a labeled latency sample in the global metrics collector.
func MeasureSinceWithLabels(name []string, start time.Time, labels []armometrics.Label) {
	armometrics.MeasureSinceWithLabels(name, start, labels)
}

// AddSample adds a sample to a histogram/summary.
//
// Summary: Adds a sample to a metric.
//
// Parameters:
//   - name ([]string): The name of the metric (as a path).
//   - val (float32): The value to sample.
//
// Returns:
//   - None.
//
// Side Effects:
//   - Records a sample in the global metrics collector.
func AddSample(name []string, val float32) {
	armometrics.AddSample(name, val)
}

// AddSampleWithLabels adds a sample to a histogram/summary with labels.
//
// Summary: Adds a labeled sample to a metric.
//
// Parameters:
//   - name ([]string): The name of the metric (as a path).
//   - val (float32): The value to sample.
//   - labels ([]armometrics.Label): The labels to apply.
//
// Returns:
//   - None.
//
// Side Effects:
//   - Records a labeled sample in the global metrics collector.
func AddSampleWithLabels(name []string, val float32, labels []armometrics.Label) {
	armometrics.AddSampleWithLabels(name, val, labels)
}
