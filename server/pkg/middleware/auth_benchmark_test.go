// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mockAuthenticator struct{}

// Authenticate ...
// Summary: Authenticate
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return ctx, nil
}

// BenchmarkAuthMiddleware_ServiceMethod ...
// Summary: BenchmarkAuthMiddleware_ServiceMethod
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	authManager := auth.NewManager()
	authManager.AddAuthenticator("myservice", &mockAuthenticator{})

	mw := middleware.AuthMiddleware(authManager)
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return nil, nil
	}
	handler := mw(next)

	httpReq, _ := http.NewRequest("POST", "/", nil)
	ctx := context.WithValue(context.Background(), middleware.HTTPRequestContextKey, httpReq)
	method := "myservice.mymethod"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = handler(ctx, method, nil)
	}
}
