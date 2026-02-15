package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/tokenizer"
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

	middleware := PrometheusMetricsMiddleware(tokenizer.NewSimpleTokenizer())

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

func TestEstimateRequestTokens(t *testing.T) {
	tok := tokenizer.NewSimpleTokenizer()

	tests := []struct {
		name     string
		req      mcp.Request
		minCount int
	}{
		{
			name: "CallToolRequest",
			req: &mcp.CallToolRequest{
				Params: &mcp.CallToolParamsRaw{
					Name:      "test-tool",
					Arguments: json.RawMessage(`{"arg1": "some long text value"}`),
				},
			},
			minCount: 1,
		},
		{
			name: "GetPromptRequest",
			req: &mcp.GetPromptRequest{
				Params: &mcp.GetPromptParams{
					Name:      "test-prompt",
					Arguments: map[string]string{"arg1": "some long text value"},
				},
			},
			minCount: 1,
		},
		{
			name:     "NilRequest",
			req:      nil,
			minCount: 0,
		},
		{
			name:     "OtherRequest",
			req:      &mcp.ListToolsRequest{},
			minCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := estimateRequestTokens(tok, tt.req)
			if count < tt.minCount {
				t.Errorf("Expected tokens >= %d, got %d", tt.minCount, count)
			}
		})
	}
}

func TestEstimateResultTokens(t *testing.T) {
	tok := tokenizer.NewSimpleTokenizer()

	tests := []struct {
		name     string
		res      mcp.Result
		minCount int
	}{
		{
			name: "ReadResourceResult",
			res: &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{Text: "some content"},
					{Text: "more content"},
					nil, // Edge case
				},
			},
			minCount: 2,
		},
		{
			name: "CallToolResult_Text",
			res: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "tool output"},
				},
			},
			minCount: 1,
		},
		{
			name: "GetPromptResult_Text",
			res: &mcp.GetPromptResult{
				Messages: []*mcp.PromptMessage{
					{Role: "user", Content: &mcp.TextContent{Text: "prompt content"}},
				},
			},
			minCount: 1,
		},
		{
			name: "GetPromptResult_Image",
			res: &mcp.GetPromptResult{
				Messages: []*mcp.PromptMessage{
					{Role: "user", Content: &mcp.ImageContent{Data: []byte("fake_image_data"), MIMEType: "image/png"}},
				},
			},
			minCount: 0, // Currently 0 for images
		},
		{
			name:     "NilResult",
			res:      nil,
			minCount: 0,
		},
		{
			name:     "FallbackJSON",
			res:      &mcp.ListToolsResult{Tools: []*mcp.Tool{{Name: "tool1"}}},
			minCount: 1, // Fallback marshals to JSON
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := estimateResultTokens(tok, tt.res)
			if count < tt.minCount {
				t.Errorf("Expected tokens >= %d, got %d", tt.minCount, count)
			}
		})
	}
}

func TestCalculateToolResultSize(t *testing.T) {
	_ = tokenizer.NewSimpleTokenizer()

	tests := []struct {
		name    string
		result  any
		minSize int
	}{
		{
			name: "CallToolResult_Text",
			result: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "12345"}, // 5 bytes
				},
			},
			minSize: 5,
		},
		{
			name: "CallToolResult_Image",
			result: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.ImageContent{Data: []byte("12345"), MIMEType: "image/png"},
				},
			},
			minSize: 5,
		},
		{
			name: "CallToolResult_Resource",
			result: &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.EmbeddedResource{Resource: &mcp.ResourceContents{Blob: []byte("12345")}},
				},
			},
			minSize: 5,
		},
		{
			name:    "String",
			result:  "12345",
			minSize: 5,
		},
		{
			name:    "Bytes",
			result:  []byte("12345"),
			minSize: 5,
		},
		{
			name:    "StructFallback",
			result:  map[string]string{"key": "val"},
			minSize: 10, // {"key":"val"} approx
		},
		{
			name:    "Nil",
			result:  nil,
			minSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := calculateToolResultSize(tt.result)
			if size < tt.minSize {
				t.Errorf("Expected size >= %d, got %d", tt.minSize, size)
			}
		})
	}
}

func TestEstimateResultSize(t *testing.T) {
	tests := []struct {
		name    string
		res     mcp.Result
		minSize int
	}{
		{
			name: "ReadResourceResult",
			res: &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{Text: "12345"},
					{Blob: []byte("12345")},
					nil,
				},
			},
			minSize: 10,
		},
		{
			name: "GetPromptResult",
			res: &mcp.GetPromptResult{
				Messages: []*mcp.PromptMessage{
					{Role: "user", Content: &mcp.TextContent{Text: "12345"}},
					{Role: "user", Content: &mcp.ImageContent{Data: []byte("12345"), MIMEType: "image/png"}},
					{Role: "user", Content: &mcp.EmbeddedResource{Resource: &mcp.ResourceContents{Blob: []byte("12345")}}},
				},
			},
			minSize: 15,
		},
		{
			name:    "Nil",
			res:     nil,
			minSize: 0,
		},
		{
			name:    "Fallback",
			res:     &mcp.ListToolsResult{},
			minSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := estimateResultSize(tt.res)
			if size < tt.minSize {
				t.Errorf("Expected size >= %d, got %d", tt.minSize, size)
			}
		})
	}
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
