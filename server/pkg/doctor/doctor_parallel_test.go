package doctor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestRunChecksParallel(t *testing.T) {
	// Enable loopback for testing
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

	// Start a test server that sleeps for 200ms
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Configure 5 services pointing to this server
	services := make([]*configv1.UpstreamServiceConfig, 5)
	for i := 0; i < 5; i++ {
		services[i] = &configv1.UpstreamServiceConfig{
			Name: proto.String("slow-service"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String(ts.URL),
				},
			},
		}
	}

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: services,
	}

	start := time.Now()
	results := RunChecks(context.Background(), config)
	duration := time.Since(start)

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}

	// Verify all passed
	for _, res := range results {
		if res.Status != StatusOk {
			t.Errorf("Expected status OK for %s, got %s: %s", res.ServiceName, res.Status, res.Message)
		}
	}

	// 5 * 200ms = 1s. Parallel should be around 200ms + overhead.
	// We allow some buffer, but if it takes > 800ms, it's definitely serial.
	if duration > 800*time.Millisecond {
		t.Errorf("RunChecks took %v, expected parallel execution (< 800ms)", duration)
	} else {
		t.Logf("RunChecks took %v (parallel confirmed)", duration)
	}
}
