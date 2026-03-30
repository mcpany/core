package haca_test

import (
	"context"
	"testing"
	"github.com/mcpany/core/server/pkg/haca"
)

func TestHACAProvider(t *testing.T) {
	provider := haca.NewProvider()
	if provider == nil {
		t.Fatal("expected Provider to be created")
	}

	ctx := context.Background()

	err := provider.AttributeTokenUsage(ctx, 100)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	err = provider.VerifyLineage(ctx, "sub-process-1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
