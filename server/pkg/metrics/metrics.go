// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package metrics provides utilities for collecting and exposing application metrics.
// NewPrometheusSink creates a new Prometheus sink for metrics collection.
//
// Summary: Creates a Prometheus sink.
//
// Parameters:
//   - None.
// Initialize prepares the metrics system with a Prometheus sink.
//
// Summary: Initializes the global metrics collector.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// It sets up a global metrics collector that can be used throughout the application.
// The metrics are exposed on the /metrics endpoint.
//
// Returns:
//   - error: An error if the initialization fails.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
package metrics

package metrics

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
// StartServer starts an HTTP server to expose the metrics.
//
// Summary: Starts the metrics server.
//
// Parameters:
//   - addr: string. The address to listen on (e.g., ":8080").
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Returns:
//   - error: An error if the server fails to start.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
// SetGauge sets the value of a gauge.
//
// Summary: Sets a gauge metric.
//
// Parameters:
//   - name: string. The name of the gauge.
//   - val: float32. The value to set.
//   - labels: ...string. A list of labels to apply to the gauge.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func SetGauge(name string, val float32, labels ...string) {
	var metricLabels []metrics.Label
// IncrCounter increments a counter.
//
// IncrCounterWithLabels increments a counter with labels.
//
// Summary: Increments a labeled counter metric.
//
// MeasureSince measures the time since a given start time and records it.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// MeasureSinceWithLabels measures the time since a given start time and records it with labels.
//
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Records labeled latency metric.
//
// AddSample adds a sample to a histogram/summary.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// AddSampleWithLabels adds a sample to a histogram/summary with labels.
//
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Adds a labeled sample to a metric.
//
// Parameters:
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - name: []string. The name of the metric (as a path).
//   - val: float32. The value to sample.
//   - labels: []metrics.Label. The labels to apply.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func AddSampleWithLabels(name []string, val float32, labels []metrics.Label) {
	metrics.AddSampleWithLabels(name, val, labels)
}
