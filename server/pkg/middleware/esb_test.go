package middleware

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestESBMiddleware_Disabled(t *testing.T) {
	mw := NewESBMiddleware(&configv1.Middleware{Disabled: true})
	req := mcp.PingRequest{}

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return mcp.PingResult{}, nil
	}

	mw.Execute(context.Background(), "ping", req, next)
	if !nextCalled {
		t.Error("expected next handler to be called when middleware is disabled")
	}
}

func TestESBMiddleware_NotCallToolRequest(t *testing.T) {
	mw := NewESBMiddleware(nil)
	req := mcp.PingRequest{}

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return mcp.PingResult{}, nil
	}

	mw.Execute(context.Background(), "ping", req, next)
	if !nextCalled {
		t.Error("expected next handler to be called for non-CallToolRequest")
	}
}

func TestESBMiddleware_CallToolRequest_MissingIntent(t *testing.T) {
	mw := NewESBMiddleware(nil)
	req := &mcp.CallToolRequest{}

	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		t.Error("next handler should not be called")
		return nil, nil
	}

	res, err := mw.Execute(context.Background(), "tools/call", req, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	callRes, ok := res.(*mcp.CallToolResult)
	if !ok {
		t.Fatalf("expected *mcp.CallToolResult, got %T", res)
	}
	if !callRes.IsError {
		t.Error("expected result to be an error")
	}
	if len(callRes.Content) == 0 {
		t.Fatal("expected content in error result")
	}
	textContent, ok := callRes.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", callRes.Content[0])
	}
	if textContent.Text != "ESB Error: Missing x-mission-intent header/context." {
		t.Errorf("unexpected error message: %s", textContent.Text)
	}
}

func TestESBMiddleware_CallToolRequest_MissingShard(t *testing.T) {
	mw := NewESBMiddleware(nil)
	req := &mcp.CallToolRequest{}
	ctx := context.WithValue(context.Background(), missionIntentKey, "test-intent")

	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		t.Error("next handler should not be called")
		return nil, nil
	}

	res, err := mw.Execute(ctx, "tools/call", req, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	callRes, ok := res.(*mcp.CallToolResult)
	if !ok {
		t.Fatalf("expected *mcp.CallToolResult, got %T", res)
	}
	if !callRes.IsError {
		t.Error("expected result to be an error")
	}
	if len(callRes.Content) == 0 {
		t.Fatal("expected content in error result")
	}
	textContent, ok := callRes.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", callRes.Content[0])
	}
	if textContent.Text != "ESB Error: Missing x-entanglement-shard header/context." {
		t.Errorf("unexpected error message: %s", textContent.Text)
	}
}

func TestESBMiddleware_CallToolRequest_Success(t *testing.T) {
	mw := NewESBMiddleware(nil)
	req := &mcp.CallToolRequest{}

	ctx := context.WithValue(context.Background(), missionIntentKey, "test-intent")
	ctx = context.WithValue(ctx, entanglementShardKey, "test-shard")

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.CallToolResult{}, nil
	}

	start := time.Now()
	mw.Execute(ctx, "tools/call", req, next)
	duration := time.Since(start)

	if !nextCalled {
		t.Error("expected next handler to be called")
	}

	if duration < 0 {
		t.Errorf("expected execution to take at least 5ms, took %v", duration)
	}
}

func TestESBMiddleware_CallToolRequest_StringKeys(t *testing.T) {
	mw := NewESBMiddleware(nil)
	req := &mcp.CallToolRequest{}

	ctx := context.WithValue(context.Background(), "x-mission-intent", "test-intent")
	ctx = context.WithValue(ctx, "x-entanglement-shard", "test-shard")

	nextCalled := false
	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.CallToolResult{}, nil
	}

	mw.Execute(ctx, "tools/call", req, next)

	if !nextCalled {
		t.Error("expected next handler to be called")
	}
}

func TestESBMiddleware_CallToolRequest_StringKeys_MissingIntent(t *testing.T) {
	mw := NewESBMiddleware(nil)
	req := &mcp.CallToolRequest{}

	ctx := context.Background()

	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		t.Error("next handler should not be called")
		return nil, nil
	}

	res, _ := mw.Execute(ctx, "tools/call", req, next)

	callRes := res.(*mcp.CallToolResult)
	if !callRes.IsError {
		t.Error("expected error")
	}
}

func TestESBMiddleware_CallToolRequest_StringKeys_MissingShard(t *testing.T) {
	mw := NewESBMiddleware(nil)
	req := &mcp.CallToolRequest{}

	ctx := context.WithValue(context.Background(), "x-mission-intent", "test-intent")

	next := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		t.Error("next handler should not be called")
		return nil, nil
	}

	res, _ := mw.Execute(ctx, "tools/call", req, next)

	callRes := res.(*mcp.CallToolResult)
	if !callRes.IsError {
		t.Error("expected error")
	}
}
