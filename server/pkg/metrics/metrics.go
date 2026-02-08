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
//
// Summary: is an alias for metrics.Label. It represents a key-value pair for labeling metrics.
type Label = metrics.Label

// NewPrometheusSink creates a new Prometheus sink for metrics collection.
//
// Summary: creates a new Prometheus sink for metrics collection.
//
// Parameters:
//   None.
//
// Returns:
//   - *prometheus.PrometheusSink: The *prometheus.PrometheusSink.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func NewPrometheusSink() (*prometheus.PrometheusSink, error) {
	return prometheus.NewPrometheusSink()
}

var initOnce sync.Once

// Initialize prepares the metrics system with a Prometheus sink.
//
// Summary: prepares the metrics system with a Prometheus sink.
//
// Parameters:
//   None.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: returns an http.Handler for the /metrics endpoint.
//
// Parameters:
//   None.
//
// Returns:
//   - http.Handler: The http.Handler.
func Handler() http.Handler {
	return promhttp.Handler()
}

// StartServer starts an HTTP server to expose the metrics.
//
// Summary: starts an HTTP server to expose the metrics.
//
// Parameters:
//   - addr: string. The addr.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
// Summary: sets the value of a gauge.
//
// Parameters:
//   - name: string. The name.
//   - val: float32. The val.
//   - labels: ...string. The labels.
//
// Returns:
//   None.
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
// Summary: increments a counter.
//
// Parameters:
//   - name: []string. The name.
//   - val: float32. The val.
//
// Returns:
//   None.
func IncrCounter(name []string, val float32) {
	metrics.IncrCounter(name, val)
}

// IncrCounterWithLabels increments a counter with labels.
//
// Summary: increments a counter with labels.
//
// Parameters:
//   - name: []string. The name.
//   - val: float32. The val.
//   - labels: []metrics.Label. The labels.
//
// Returns:
//   None.
func IncrCounterWithLabels(name []string, val float32, labels []metrics.Label) {
	metrics.IncrCounterWithLabels(name, val, labels)
}

// MeasureSince measures the time since a given start time and records it.
//
// Summary: measures the time since a given start time and records it.
//
// Parameters:
//   - name: []string. The name.
//   - start: time.Time. The start.
//
// Returns:
//   None.
func MeasureSince(name []string, start time.Time) {
	metrics.MeasureSince(name, start)
}

// MeasureSinceWithLabels measures the time since a given start time and records it with labels.
//
// Summary: measures the time since a given start time and records it with labels.
//
// Parameters:
//   - name: []string. The name.
//   - start: time.Time. The start.
//   - labels: []metrics.Label. The labels.
//
// Returns:
//   None.
func MeasureSinceWithLabels(name []string, start time.Time, labels []metrics.Label) {
	metrics.MeasureSinceWithLabels(name, start, labels)
}

// AddSample adds a sample to a histogram/summary.
//
// Summary: adds a sample to a histogram/summary.
//
// Parameters:
//   - name: []string. The name.
//   - val: float32. The val.
//
// Returns:
//   None.
func AddSample(name []string, val float32) {
	metrics.AddSample(name, val)
}

// AddSampleWithLabels adds a sample to a histogram/summary with labels.
//
// Summary: adds a sample to a histogram/summary with labels.
//
// Parameters:
//   - name: []string. The name.
//   - val: float32. The val.
//   - labels: []metrics.Label. The labels.
//
// Returns:
//   None.
func AddSampleWithLabels(name []string, val float32, labels []metrics.Label) {
	metrics.AddSampleWithLabels(name, val, labels)
}
