package middleware

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSMMMiddleware_PassesValidRequest(t *testing.T) {
	// Setup SMM with a reasonable threshold.
	smm := NewSMMMiddleware(50.0)

	// Create a valid tool call request.
	argsBytes, _ := json.Marshal(map[string]interface{}{"query": "hello world"})
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "test",
			Arguments: argsBytes,
		},
	}

	// The next handler simply returns a success response.
	nextCalled := false
	next := func(ctx context.Context, method string, r mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.CallToolResult{Content: []mcp.Content{}}, nil
	}

	// Execute SMM.
	res, err := smm.Execute(context.Background(), "tools/call", req, next)

	// Verify the next handler was called and no error was returned.
	if !nextCalled {
		t.Error("expected next handler to be called")
	}
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if res == nil {
		t.Error("expected a result, got nil")
	}
}

func TestSMMMiddleware_BlocksShadowingAttack(t *testing.T) {
	// Setup SMM with a low threshold for the test.
	smm := NewSMMMiddleware(50.0)

	// Create a malicious tool call request with high entropy / mimicry payload.
	argsBytes, _ := json.Marshal(map[string]interface{}{"query": "SHADOW_MIMIC_PAYLOAD injected context"})
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "test",
			Arguments: argsBytes,
		},
	}

	// The next handler should NOT be called.
	nextCalled := false
	next := func(ctx context.Context, method string, r mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.CallToolResult{Content: []mcp.Content{}}, nil
	}

	// Execute SMM.
	res, err := smm.Execute(context.Background(), "tools/call", req, next)

	// Verify the request was blocked with a security exception.
	if nextCalled {
		t.Error("expected next handler NOT to be called")
	}
	if err == nil {
		t.Fatal("expected an error response, got nil")
	}
	if res != nil {
		t.Errorf("expected nil result, got %v", res)
	}
	if !strings.Contains(err.Error(), "SMM Security Exception") {
		t.Errorf("expected SMM Security Exception message, got %v", err.Error())
	}
}

func TestSMMMiddleware_OtherMethods(t *testing.T) {
	smm := NewSMMMiddleware(50.0)

	// Not tools/call
	req := &mcp.InitializeRequest{}

	nextCalled := false
	next := func(ctx context.Context, method string, r mcp.Request) (mcp.Result, error) {
		nextCalled = true
		return &mcp.InitializeResult{}, nil
	}

	_, err := smm.Execute(context.Background(), "initialize", req, next)
	if err != nil {
		t.Fatal(err)
	}
	if !nextCalled {
		t.Error("expected next handler to be called for non tools/call request")
	}
}
