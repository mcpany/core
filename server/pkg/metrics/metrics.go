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

// NewPrometheusSink provides newprometheussink functionality.
//
// Summary: NewPrometheusSink.
//
// Parameters.
//   - ): The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func NewPrometheusSink() (*prometheus.PrometheusSink, error) {
	return prometheus.NewPrometheusSink()
}

var initOnce sync.Once

// Initialize provides initialize functionality.
//
// Summary: Initialize.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
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

// Handler provides handler functionality.
//
// Summary: Handler.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func Handler() http.Handler {
	return promhttp.Handler()
}

// StartServer provides startserver functionality.
//
// Summary: StartServer.
//
// Parameters.
//   - addr: The parameter.
//
// Returns.
//   - result: The result.
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

// SetGauge provides setgauge functionality.
//
// Summary: SetGauge.
//
// Parameters.
//   - name: The parameter.
//   - val: The parameter.
//   - labels: The parameter.
//
// Returns.
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

// IncrCounter provides incrcounter functionality.
//
// Summary: IncrCounter.
//
// Parameters.
//   - name: The parameter.
//   - val: The parameter.
//
// Returns.
//   - None.
func IncrCounter(name []string, val float32) {
	metrics.IncrCounter(name, val)
}

// IncrCounterWithLabels provides incrcounterwithlabels functionality.
//
// Summary: IncrCounterWithLabels.
//
// Parameters.
//   - name: The parameter.
//   - val: The parameter.
//   - labels: The parameter.
//
// Returns.
//   - None.
func IncrCounterWithLabels(name []string, val float32, labels []metrics.Label) {
	metrics.IncrCounterWithLabels(name, val, labels)
}

// MeasureSince provides measuresince functionality.
//
// Summary: MeasureSince.
//
// Parameters.
//   - name: The parameter.
//   - start: The parameter.
//
// Returns.
//   - None.
func MeasureSince(name []string, start time.Time) {
	metrics.MeasureSince(name, start)
}

// MeasureSinceWithLabels provides measuresincewithlabels functionality.
//
// Summary: MeasureSinceWithLabels.
//
// Parameters.
//   - name: The parameter.
//   - start: The parameter.
//   - labels: The parameter.
//
// Returns.
//   - None.
func MeasureSinceWithLabels(name []string, start time.Time, labels []metrics.Label) {
	metrics.MeasureSinceWithLabels(name, start, labels)
}

// AddSample provides addsample functionality.
//
// Summary: AddSample.
//
// Parameters.
//   - name: The parameter.
//   - val: The parameter.
//
// Returns.
//   - None.
func AddSample(name []string, val float32) {
	metrics.AddSample(name, val)
}

// AddSampleWithLabels provides addsamplewithlabels functionality.
//
// Summary: AddSampleWithLabels.
//
// Parameters.
//   - name: The parameter.
//   - val: The parameter.
//   - labels: The parameter.
//
// Returns.
//   - None.
func AddSampleWithLabels(name []string, val float32, labels []metrics.Label) {
	metrics.AddSampleWithLabels(name, val, labels)
}
