package rrr_test

import (
	"context"
	"testing"
	"github.com/mcpany/core/server/pkg/rrr"
)

func TestRRRManager(t *testing.T) {
	manager := rrr.NewManager()
	if manager == nil {
		t.Fatal("expected Manager to be created")
	}

	ctx := context.Background()

	err := manager.Reclaim(ctx, "mission-1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
