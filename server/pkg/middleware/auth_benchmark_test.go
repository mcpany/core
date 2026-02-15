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

func (m *mockAuthenticator) Authenticate(ctx context.Context, r *http.Request) (context.Context, error) {
	return ctx, nil
}

func BenchmarkAuthMiddleware_ServiceMethod(b *testing.B) {
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
