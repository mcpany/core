package tool

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mockSession struct{}

func (m *mockSession) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	return nil, nil
}

func (m *mockSession) ListRoots(ctx context.Context) (*mcp.ListRootsResult, error) {
	return nil, nil
}

func TestContextWithSession(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := &mockSession{}

	ctx = NewContextWithSession(ctx, s)

	retrievedSession, ok := GetSession(ctx)
	if !ok {
		t.Errorf("expected Session to be present in context")
	}
	if retrievedSession != s {
		t.Errorf("expected Session %v, got %v", s, retrievedSession)
	}

	// Test missing Session
	_, ok = GetSession(context.Background())
	if ok {
		t.Errorf("expected Session to be absent in empty context")
	}
}

func TestContextWithSampler_BackwardCompatibility(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := &mockSession{}

	// Use deprecated method
	ctx = NewContextWithSampler(ctx, s)

	// Retrieve with deprecated method
	retrievedSampler, ok := GetSampler(ctx)
	if !ok {
		t.Errorf("expected Sampler to be present in context")
	}
	if retrievedSampler != s {
		t.Errorf("expected Sampler %v, got %v", s, retrievedSampler)
	}

	// Retrieve with new method
	retrievedSession, ok := GetSession(ctx)
	if !ok {
		t.Errorf("expected Session to be present in context")
	}
	if retrievedSession != s {
		t.Errorf("expected Session %v, got %v", s, retrievedSession)
	}
}
