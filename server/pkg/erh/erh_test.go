package erh_test

import (
	"context"
	"testing"
	"github.com/mcpany/core/server/pkg/erh"
)

func TestERHProvider(t *testing.T) {
	provider := erh.NewProvider()
	if provider == nil {
		t.Fatal("expected Provider to be created")
	}

	ctx := context.Background()

	err := provider.MandateSessionLocked(ctx, "schema-1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
