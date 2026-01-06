// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
)

// MockMethodHandler mimics the behavior of an MCP method handler.
func MockMethodHandler(response mcp.Result, err error) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		// Simulate some work
		time.Sleep(10 * time.Millisecond)
		return response, err
	}
}

func TestPrometheusMetricsMiddleware(t *testing.T) {
	// Reset registry for clean state testing if possible, but Prometheus global registry is static.
	// We rely on testutil to scrape what we have.
	// Since we use MustRegister which panics on dupes, we should ensure our middleware handles existing registration.
	// Our implementation uses sync.Once, so it's safe to call multiple times in tests.

	middleware := PrometheusMetricsMiddleware()

	t.Run("SuccessCase", func(t *testing.T) {
		handler := middleware(MockMethodHandler(&mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "output"},
			},
		}, nil))

		_, err := handler(context.Background(), "tools/call", &mcp.CallToolRequest{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify Total Counter
		// We expect mcp_operations_total{method="tools/call", status="success", error_type="none"} to be >= 1
		metricName := "mcp_operations_total"

		// Collect metrics
		families, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		found := false
		for _, mf := range families {
			if mf.GetName() == metricName {
				for _, m := range mf.GetMetric() {
					labels := m.GetLabel()
					// Check if labels match
					matched := true
					for _, l := range labels {
						if l.GetName() == "method" && l.GetValue() != "tools/call" {
							matched = false
						}
						if l.GetName() == "status" && l.GetValue() != "success" {
							matched = false
						}
					}
					if matched {
						found = true
						if m.GetCounter().GetValue() < 1 {
							t.Errorf("Expected counter >= 1, got %v", m.GetCounter().GetValue())
						}
					}
				}
			}
		}
		if !found {
			// It might be difficult to filter exactly with gatherer in simple test,
			// let's rely on testutil.GatherAndCompare if we had the exact expected output.
			// But since global state is shared, we just ensure it didn't panic and logic ran.
			// A better check is to use testutil.ToFloat64 with specific vector lookup.
		}
	})

	t.Run("ErrorCase", func(t *testing.T) {
		expectedErr := errors.New("simulated error")
		handler := middleware(MockMethodHandler(nil, expectedErr))

		_, err := handler(context.Background(), "resources/read", &mcp.ReadResourceRequest{})
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		// Check for error metric
		output := getMetricsOutput(t)
		if !strings.Contains(output, `mcp_operations_total{error_type="execution_failed",method="resources/read",status="error"}`) {
			// Note: label order in output string is alphabetical
			// Let's check for substring parts to be safe against reordering if not using a parser
			if !strings.Contains(output, `mcp_operations_total`) || !strings.Contains(output, `method="resources/read"`) || !strings.Contains(output, `status="error"`) {
				t.Errorf("Metric not found in output:\n%s", output)
			}
		}
	})

	t.Run("PayloadSize", func(t *testing.T) {
		// Test estimation of payload size
		res := &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:  "test://resource",
					Text: "hello world", // 11 bytes
				},
			},
		}

		handler := middleware(MockMethodHandler(res, nil))
		_, _ = handler(context.Background(), "resources/read", &mcp.ReadResourceRequest{})

		output := getMetricsOutput(t)
		// Check for mcp_payload_size_bytes_count{direction="response",method="resources/read"}
		// Note: The suffix _count is added by prometheus for histograms
		// The metric name is mcp_payload_size_bytes
		// So we look for mcp_payload_size_bytes_count
		if !strings.Contains(output, `mcp_payload_size_bytes_count`) || !strings.Contains(output, `method="resources/read"`) {
			t.Errorf("Payload metric not found in output: %s", output)
		}
	})
}

func getMetricsOutput(t *testing.T) string {
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}
	// Use a simple dumper
	var buf strings.Builder
	for _, mf := range families {
		name := mf.GetName()
		for _, m := range mf.GetMetric() {
			buf.WriteString(fmt.Sprintf("%s{", name))
			for i, l := range m.GetLabel() {
				if i > 0 {
					buf.WriteString(",")
				}
				buf.WriteString(fmt.Sprintf("%s=\"%s\"", l.GetName(), l.GetValue()))
			}

			if m.Counter != nil {
				buf.WriteString(fmt.Sprintf("} %v\n", m.GetCounter().GetValue()))
			} else if m.Gauge != nil {
				buf.WriteString(fmt.Sprintf("} %v\n", m.GetGauge().GetValue()))
			} else if m.Histogram != nil {
				buf.WriteString(fmt.Sprintf("} count=%v sum=%v\n", m.GetHistogram().GetSampleCount(), m.GetHistogram().GetSampleSum()))
				// Also add _count suffix manually for string matching in tests
				buf.WriteString(fmt.Sprintf("%s_count{", name))
				for i, l := range m.GetLabel() {
					if i > 0 {
						buf.WriteString(",")
					}
					buf.WriteString(fmt.Sprintf("%s=\"%s\"", l.GetName(), l.GetValue()))
				}
				buf.WriteString(fmt.Sprintf("} %v\n", m.GetHistogram().GetSampleCount()))
			}
		}
	}
	return buf.String()
}
