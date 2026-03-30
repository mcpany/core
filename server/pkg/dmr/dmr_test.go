package dmr_test

import (
	"context"
	"testing"
	"github.com/mcpany/core/server/pkg/dmr"
)

func TestDMRHub(t *testing.T) {
	hub := dmr.NewHub()
	if hub == nil {
		t.Fatal("expected Hub to be created")
	}

	ctx := context.Background()

	err := hub.Heartbeat(ctx, "node-1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	err = hub.Reshard(ctx, "state")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
