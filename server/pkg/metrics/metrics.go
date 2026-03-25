// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package metrics provides utilities for collecting and exposing application metrics.
//
// Summary: Metrics collection and exposition utilities.
package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	armonmetrics "github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Label is an alias for armonmetrics.Label. It represents a key-value pair for labeling metrics.
//
// Summary: Represents a metric label.
type Label = armonmetrics.Label

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
// Errors:
//   - Returns error if the prometheus sink cannot be created.
//
// Side Effects:
//   - None.
func NewPrometheusSink() (*prometheus.PrometheusSink, error) {
	return prometheus.NewPrometheusSink()
}

var (
	initOnce sync.Once
	initErr  error
)

// Initialize prepares the metrics system with a Prometheus sink.
//
// Summary: Initializes the global metrics collector.
//
// Parameters:
//   - None.
//
// Returns:
//   - error: An error if the initialization fails.
//
// Errors:
//   - Returns error if sink creation or global initialization fails.
//
// Side Effects:
//   - Initializes a global metrics sink.
func Initialize() error {
	initOnce.Do(func() {
		// Create a Prometheus sink
		var sink *prometheus.PrometheusSink
		sink, initErr = NewPrometheusSink()
		if initErr != nil {
			return
		}

		// Create a metrics configuration
		conf := armonmetrics.DefaultConfig("mcpany")
		conf.EnableHostname = false

		// Initialize the metrics system
		if _, initErr = armonmetrics.NewGlobal(conf, sink); initErr != nil {
			return
		}
	})
	return initErr
}

// Handler returns an http.Handler for the /metrics endpoint.
//
// Summary: Retrieves the metrics HTTP handler.
//
// Parameters:
//   - None.
//
// Returns:
//   - http.Handler: An http.Handler that serves the Prometheus metrics.
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
//
// Summary: Starts the metrics server.
//
// Parameters:
//   - addr (string): The address to listen on.
//
// Returns:
//   - error: An error if the server fails to start.
//
// Errors:
//   - Returns error if the listener cannot be started or the server fails.
//
// Side Effects:
//   - Starts a background HTTP server.
func StartServer(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", Handler())

	var lc net.ListenConfig
	ln, err := lc.Listen(context.Background(), "tcp", addr)
	if err != nil {
		return err
	}

	if tcpAddr, ok := ln.Addr().(*net.TCPAddr); ok {
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
// Errors:
//   - None.
//
// Side Effects:
//   - Updates a global metric.
func SetGauge(name string, val float32, labels ...string) {
	var metricLabels []Label
	if len(labels) > 0 {
		metricLabels = []Label{
			{Name: "service_name", Value: labels[0]},
		}
	}
	armonmetrics.SetGaugeWithLabels([]string{name}, val, metricLabels)
}

// IncrCounter increments a counter.
//
// Summary: Increments a counter metric.
//
// Parameters:
//   - name ([]string): The name of the counter.
//   - val (float32): The amount to increment.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Updates a global metric.
func IncrCounter(name []string, val float32) {
	armonmetrics.IncrCounter(name, val)
}

// IncrCounterWithLabels increments a counter with labels.
//
// Summary: Increments a labeled counter metric.
//
// Parameters:
//   - name ([]string): The name of the counter.
//   - val (float32): The amount to increment.
//   - labels ([]Label): The labels to apply.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Updates a global metric.
func IncrCounterWithLabels(name []string, val float32, labels []Label) {
	armonmetrics.IncrCounterWithLabels(name, val, labels)
}

// MeasureSince measures the time since a given start time.
//
// Summary: Records latency metric.
//
// Parameters:
//   - name ([]string): The name of the metric.
//   - start (time.Time): The start time.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Updates a global metric.
func MeasureSince(name []string, start time.Time) {
	armonmetrics.MeasureSince(name, start)
}

// MeasureSinceWithLabels measures the time since a given start time with labels.
//
// Summary: Records labeled latency metric.
//
// Parameters:
//   - name ([]string): The name of the metric.
//   - start (time.Time): The start time.
//   - labels ([]Label): The labels to apply.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Updates a global metric.
func MeasureSinceWithLabels(name []string, start time.Time, labels []Label) {
	armonmetrics.MeasureSinceWithLabels(name, start, labels)
}

// AddSample adds a sample to a histogram.
//
// Summary: Adds a sample to a metric.
//
// Parameters:
//   - name ([]string): The name of the metric.
//   - val (float32): The value to sample.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Updates a global metric.
func AddSample(name []string, val float32) {
	armonmetrics.AddSample(name, val)
}

// AddSampleWithLabels adds a sample to a histogram with labels.
//
// Summary: Adds a labeled sample to a metric.
//
// Parameters:
//   - name ([]string): The name of the metric.
//   - val (float32): The value to sample.
//   - labels ([]Label): The labels to apply.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Updates a global metric.
func AddSampleWithLabels(name []string, val float32, labels []Label) {
	armonmetrics.AddSampleWithLabels(name, val, labels)
}
