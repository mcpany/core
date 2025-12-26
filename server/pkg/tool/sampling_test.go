// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mockSampler struct{}

func (m *mockSampler) CreateMessage(ctx context.Context, params *mcp.CreateMessageParams) (*mcp.CreateMessageResult, error) {
	return nil, nil
}

func TestContextWithSampler(t *testing.T) {
	ctx := context.Background()
	s := &mockSampler{}

	ctx = NewContextWithSampler(ctx, s)

	retrievedSampler, ok := GetSampler(ctx)
	if !ok {
		t.Errorf("expected Sampler to be present in context")
	}
	if retrievedSampler != s {
		t.Errorf("expected Sampler %v, got %v", s, retrievedSampler)
	}

	// Test missing Sampler
	_, ok = GetSampler(context.Background())
	if ok {
		t.Errorf("expected Sampler to be absent in empty context")
	}
}
