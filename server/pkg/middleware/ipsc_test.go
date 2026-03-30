package middleware

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestIPSCMiddleware(t *testing.T) {
	config := IPSCConfig{
		Enabled:               true,
		DefaultCorrectionLimit: 2,
	}

	ipsc := NewIPSCMiddleware(config)

	ctx := context.Background()

	// 1. First call - should succeed and budget should be initialized

	nextHandler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	_, err := ipsc.Execute(ctx, "tools/call", mockCallToolRequest("refine_data", nil), nextHandler)
	if err != nil {
		t.Fatalf("Expected success on first call, got error: %v", err)
	}

	// 2. Second call - should succeed, budget decremented
	_, err = ipsc.Execute(ctx, "tools/call", mockCallToolRequest("refine_data", nil), nextHandler)
	if err != nil {
		t.Fatalf("Expected success on second call, got error: %v", err)
	}

	// 3. Third call - should fail due to Cognitive Lock
	_, err = ipsc.Execute(ctx, "tools/call", mockCallToolRequest("refine_data", nil), nextHandler)
	if err == nil || err.Error() != "Cognitive Lock Detected: Correction Budget exhausted for intent session. Mandatory Intent Re-Verification required." {
		t.Fatalf("Expected Cognitive Lock error, got: %v", err)
	}

	// 4. Test Ghost Fragment Mutation
	ipsc.ResetBudget("global_intent_safe_tool")
	_, err = ipsc.Execute(ctx, "tools/call", mockCallToolRequest("safe_tool", map[string]interface{}{"payload": "__ghost_fragment__"}), nextHandler)
	if err == nil || err.Error() != "Continuous BSH Integrity Monitor Failed: Ghost Fragment Mutation detected in payload." {
		t.Fatalf("Expected Ghost Fragment Mutation error, got: %v", err)
	}
}

func mockCallToolRequest(name string, args map[string]interface{}) mcp.Request {
	req := &mcp.CallToolRequest{}
	// Set the fields via JSON unmarshal to bypass the complex nested SDK types
	data := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "1",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      name,
		},
	}
	if args != nil {
		bytes, _ := json.Marshal(args)
		data["params"].(map[string]interface{})["arguments"] = json.RawMessage(bytes)
	}

	bytes, _ := json.Marshal(data)
	_ = json.Unmarshal(bytes, req)

	return req
}
