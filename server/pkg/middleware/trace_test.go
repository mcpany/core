package middleware

import (
	"context"
	"testing"
)

func TestTraceContext(t *testing.T) {
	ctx := context.Background()

	ctx = WithTraceContext(ctx, "trace-123", "span-456", "parent-789")

	if GetTraceID(ctx) != "trace-123" {
		t.Errorf("expected trace ID 'trace-123', got %v", GetTraceID(ctx))
	}

	if GetSpanID(ctx) != "span-456" {
		t.Errorf("expected span ID 'span-456', got %v", GetSpanID(ctx))
	}

	if GetParentID(ctx) != "parent-789" {
		t.Errorf("expected parent ID 'parent-789', got %v", GetParentID(ctx))
	}

	// Test missing values
	emptyCtx := context.Background()

	if GetTraceID(emptyCtx) != "" {
		t.Errorf("expected empty trace ID, got %v", GetTraceID(emptyCtx))
	}

	if GetSpanID(emptyCtx) != "" {
		t.Errorf("expected empty span ID, got %v", GetSpanID(emptyCtx))
	}

	if GetParentID(emptyCtx) != "" {
		t.Errorf("expected empty parent ID, got %v", GetParentID(emptyCtx))
	}

	// Test missing parent ID
	ctxNoParent := WithTraceContext(context.Background(), "trace-123", "span-456", "")
	if GetParentID(ctxNoParent) != "" {
		t.Errorf("expected empty parent ID, got %v", GetParentID(ctxNoParent))
	}
}
