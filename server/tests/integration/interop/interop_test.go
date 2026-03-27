package interop_test

import (
	"context"
	"github.com/mcpany/core/src/interop"
	"testing"
)

// TestInteropIntegration verifies the interop hub using the actual implementations
// without database seeding since the interop logic is purely in-memory routing.
// Any upstream frameworks added in the future with API connectivity should
// mock those connections or use local test endpoints here.
func TestInteropIntegration(t *testing.T) {
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewOpenClawAdapter())

	ctx := context.Background()
	task := &interop.Task{
		ID:        "int-1",
		Framework: "OpenClaw",
		Intent:    "adaptive_reasoning",
		Payload:   map[string]string{"foo": "bar"},
	}

	res, err := hub.RouteTask(ctx, task)
	if err != nil {
		t.Fatalf("Integration task failed: %v", err)
	}

	if res.Status != "success" {
		t.Errorf("Expected success, got %s", res.Status)
	}
}
