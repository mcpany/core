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

	armonmetrics "github.com/armon/go-metrics"
	"github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Label is an alias for armonmetrics.Label.
//
// Summary: Represents a Label.
type Label = armonmetrics.Label

// NewPrometheusSink creates a new Prometheus sink for metrics collection.
//
// Summary: Creates a Prometheus sink.
//
// Parameters: None.
//
// Returns:
//   - *prometheus.PrometheusSink: The initialized Prometheus sink.
//   - error: An error if the sink creation fails.
func NewPrometheusSink() (*prometheus.PrometheusSink, error) {
	return prometheus.NewPrometheusSink()
}

var initOnce sync.Once

// Initialize prepares the metrics system with a Prometheus sink.
//
// Summary: Initializes the system.
//
// Parameters: None.
//
// Returns:
//   - error: An error if the initialization fails.
func Initialize() error {
	var err error
	initOnce.Do(func() {
		sink, sinkErr := NewPrometheusSink()
		if sinkErr != nil {
			err = sinkErr
			return
		}
		conf := armonmetrics.DefaultConfig("mcpany")
		conf.EnableHostname = false
		if _, globalErr := armonmetrics.NewGlobal(conf, sink); globalErr != nil {
			err = globalErr
			return
		}
	})
	return err
}

// Handler returns an http.Handler for the /metrics endpoint.
//
// Summary: Retrieves the metrics HTTP handler.
//
// Parameters: None.
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
//   - addr (string): The address to listen on.
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

// SetGauge sets a gauge value.
//
// Summary: Sets a gauge.
//
// Parameters:
//   - name (string): Metric name.
//   - val (float32): Value.
//   - labels (...string): Labels.
func SetGauge(name string, val float32, labels ...string) {
	var metricLabels []armonmetrics.Label
	if len(labels) > 0 {
		metricLabels = []armonmetrics.Label{{Name: "service_name", Value: labels[0]}}
	}
	armonmetrics.SetGaugeWithLabels([]string{name}, val, metricLabels)
}

// IncrCounter increments a counter.
//
// Summary: Increments a counter.
//
// Parameters:
//   - name ([]string): Metric path.
//   - val (float32): Amount.
func IncrCounter(name []string, val float32) {
	armonmetrics.IncrCounter(name, val)
}

// IncrCounterWithLabels increments a labeled counter.
//
// Summary: Increments a labeled counter.
//
// Parameters:
//   - name ([]string): Metric path.
//   - val (float32): Amount.
//   - labels ([]armonmetrics.Label): Labels.
func IncrCounterWithLabels(name []string, val float32, labels []armonmetrics.Label) {
	armonmetrics.IncrCounterWithLabels(name, val, labels)
}

// MeasureSince records a latency.
//
// Summary: Records latency since start.
//
// Parameters:
//   - name ([]string): Metric path.
//   - start (time.Time): Start time.
func MeasureSince(name []string, start time.Time) {
	armonmetrics.MeasureSince(name, start)
}

// MeasureSinceWithLabels records a labeled latency.
//
// Summary: Records labeled latency since start.
//
// Parameters:
//   - name ([]string): Metric path.
//   - start (time.Time): Start time.
//   - labels ([]armonmetrics.Label): Labels.
func MeasureSinceWithLabels(name []string, start time.Time, labels []armonmetrics.Label) {
	armonmetrics.MeasureSinceWithLabels(name, start, labels)
}

// AddSample adds a sample.
//
// Summary: Adds a sample to a histogram.
//
// Parameters:
//   - name ([]string): Metric path.
//   - val (float32): Value.
func AddSample(name []string, val float32) {
	armonmetrics.AddSample(name, val)
}

// AddSampleWithLabels adds a labeled sample.
//
// Summary: Adds a labeled sample.
//
// Parameters:
//   - name ([]string): Metric path.
//   - val (float32): Value.
//   - labels ([]armonmetrics.Label): Labels.
func AddSampleWithLabels(name []string, val float32, labels []armonmetrics.Label) {
	armonmetrics.AddSampleWithLabels(name, val, labels)
}
