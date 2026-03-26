// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mockSession struct{}

// CreateMessage ...
// Summary: CreateMessage
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}

// ListRoots ...
// Summary: ListRoots
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}

// TestContextWithSession ...
// Summary: TestContextWithSession
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestContextWithSampler_BackwardCompatibility ...
// Summary: TestContextWithSampler_BackwardCompatibility
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
