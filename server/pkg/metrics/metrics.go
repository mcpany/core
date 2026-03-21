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

// Summary: Label is an alias for metrics.Label. It represents a key-value pair for labeling metrics. Represents a Label.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Label = metrics.Label

// Summary: NewPrometheusSink creates a new Prometheus sink for metrics collection. Creates a Prometheus sink.
//
// Parameters:
//   - None.
//
// Returns:
//   - *prometheus.PrometheusSink: The resulting *prometheus.PrometheusSink.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func NewPrometheusSink() (*prometheus.PrometheusSink, error) {
	return prometheus.NewPrometheusSink()
}

var initOnce sync.Once

// Summary: Initialize prepares the metrics system with a Prometheus sink. Initializes the global metrics collector. It sets up a global metrics collector that can be used throughout the application. The metrics are exposed on the /metrics endpoint.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
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

// Summary: Handler returns an http.Handler for the /metrics endpoint. Retrieves the metrics HTTP handler.
//
// Parameters:
//   - None.
//
// Returns:
//   - http.Handler: The resulting http.Handler.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func Handler() http.Handler {
	return promhttp.Handler()
}

// Summary: StartServer starts an HTTP server to expose the metrics. Starts the metrics server.
//
// Parameters:
//   - addr (string): The addr parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
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

// Summary: SetGauge sets the value of a gauge. Sets a gauge metric.
//
// Parameters:
//   - name (string): The name parameter.
//   - val (float32): The val parameter.
//   - labels (...string): The labels parameter.
//
// Returns:
//   - None.
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

// Summary: IncrCounter increments a counter. Increments a counter metric.
//
// Parameters:
//   - name ([]string): The name parameter.
//   - val (float32): The val parameter.
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

// Summary: IncrCounterWithLabels increments a counter with labels. Increments a labeled counter metric.
//
// Parameters:
//   - name ([]string): The name parameter.
//   - val (float32): The val parameter.
//   - labels ([]metrics.Label): The labels parameter.
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

// Summary: MeasureSince measures the time since a given start time and records it. Records latency metric.
//
// Parameters:
//   - name ([]string): The name parameter.
//   - start (time.Time): The start parameter.
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

// Summary: MeasureSinceWithLabels measures the time since a given start time and records it with labels. Records labeled latency metric.
//
// Parameters:
//   - name ([]string): The name parameter.
//   - start (time.Time): The start parameter.
//   - labels ([]metrics.Label): The labels parameter.
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

// Summary: AddSample adds a sample to a histogram/summary. Adds a sample to a metric.
//
// Parameters:
//   - name ([]string): The name parameter.
//   - val (float32): The val parameter.
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

// Summary: AddSampleWithLabels adds a sample to a histogram/summary with labels. Adds a labeled sample to a metric.
//
// Parameters:
//   - name ([]string): The name parameter.
//   - val (float32): The val parameter.
//   - labels ([]metrics.Label): The labels parameter.
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
