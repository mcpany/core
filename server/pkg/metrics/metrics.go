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
// Label is an alias for armonmetrics.Label.
//
// Summary: Represents a Label.
type Label = armonmetrics.Label

// NewPrometheusSink creates a new Prometheus sink.
//
// Summary: Creates a Prometheus sink.
//
// Parameters: None.
//
// Returns:
//   - *prometheus.PrometheusSink: The sink.
//   - error: Error if any.
// NewPrometheusSink creates a new Prometheus sink.
//
// Summary: creates a new Prometheus sink.
// Returns: *prometheus.PrometheusSink, error
func NewPrometheusSink() (*prometheus.PrometheusSink, error) {
	return prometheus.NewPrometheusSink()
}

var initOnce sync.Once

// Initialize prepares the metrics system.
//
// Summary: Initializes the system.
//
// Parameters: None.
//
// Returns:
//   - error: Error if any.
// Initialize prepares the metrics system.
//
// Summary: prepares the metrics system.
// Returns: error
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

// Handler returns the metrics handler.
//
// Summary: Returns the metrics HTTP handler.
//
// Parameters: None.
//
// Returns:
//   - http.Handler: The handler.
// Handler returns the metrics HTTP handler.
//
// Summary: returns the metrics HTTP handler.
// Returns: http.Handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// StartServer starts the HTTP server.
//
// Summary: Starts the metrics server.
//
// Parameters:
//   - addr (string): The address to listen on.
//
// Returns:
//   - error: Error if any.
// StartServer starts the HTTP server.
//
// Summary: starts the HTTP server.
// Parameters: addr: string
// Returns: error
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

// SetGauge sets a gauge.
//
// Summary: Sets a gauge value.
//
// Parameters:
//   - name (string): Metric name.
//   - val (float32): Value.
//   - labels (...string): Labels.
// SetGauge sets a gauge value.
//
// Summary: sets a gauge value.
// Parameters: name: string, val: float32, labels: ...string
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
//   - val (float32): Increment amount.
// IncrCounter increments a counter.
//
// Summary: increments a counter.
// Parameters: name: []string, val: float32
func IncrCounter(name []string, val float32) {
	armonmetrics.IncrCounter(name, val)
}

// IncrCounterWithLabels increments a labeled counter.
//
// Summary: Increments a labeled counter.
//
// Parameters:
//   - name ([]string): Metric path.
//   - val (float32): Increment amount.
//   - labels ([]armonmetrics.Label): Labels.
// IncrCounterWithLabels increments a counter with labels.
//
// Summary: increments a counter with labels.
// Parameters: name: []string, val: float32, labels: []armonmetrics.Label
func IncrCounterWithLabels(name []string, val float32, labels []armonmetrics.Label) {
	armonmetrics.IncrCounterWithLabels(name, val, labels)
}

// MeasureSince measures duration.
//
// Summary: Records latency.
//
// Parameters:
//   - name ([]string): Metric path.
//   - start (time.Time): Start time.
// MeasureSince measures duration since start.
//
// Summary: measures duration since start.
// Parameters: name: []string, start: time.Time
func MeasureSince(name []string, start time.Time) {
	armonmetrics.MeasureSince(name, start)
}

// MeasureSinceWithLabels measures labeled duration.
//
// Summary: Records labeled latency.
//
// Parameters:
//   - name ([]string): Metric path.
//   - start (time.Time): Start time.
//   - labels ([]armonmetrics.Label): Labels.
// MeasureSinceWithLabels measures duration with labels.
//
// Summary: measures duration with labels.
// Parameters: name: []string, start: time.Time, labels: []armonmetrics.Label
func MeasureSinceWithLabels(name []string, start time.Time, labels []armonmetrics.Label) {
	armonmetrics.MeasureSinceWithLabels(name, start, labels)
}

// AddSample adds a sample.
//
// Summary: Adds a sample to a histogram.
//
// Parameters:
//   - name ([]string): Metric path.
//   - val (float32): Sample value.
// AddSample adds a sample.
//
// Summary: adds a sample.
// Parameters: name: []string, val: float32
func AddSample(name []string, val float32) {
	armonmetrics.AddSample(name, val)
}

// AddSampleWithLabels adds a labeled sample.
//
// Summary: Adds a labeled sample.
//
// Parameters:
//   - name ([]string): Metric path.
//   - val (float32): Sample value.
//   - labels ([]armonmetrics.Label): Labels.
// AddSampleWithLabels adds a sample with labels.
//
// Summary: adds a sample with labels.
// Parameters: name: []string, val: float32, labels: []armonmetrics.Label
func AddSampleWithLabels(name []string, val float32, labels []armonmetrics.Label) {
	armonmetrics.AddSampleWithLabels(name, val, labels)
}
