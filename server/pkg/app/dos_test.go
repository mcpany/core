package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream/factory"
)

func TestHandleStatelessJSONRPC_DoS(t *testing.T) {
	// Setup dependencies
	busProvider, _ := bus.NewProvider(nil)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager)

	serviceRegistry := serviceregistry.New(
		upstreamFactory,
		toolManager,
		promptManager,
		resourceManager,
		authManager,
	)

	mcpSrv, err := mcpserver.NewServer(
		context.Background(),
		toolManager,
		promptManager,
		resourceManager,
		authManager,
		serviceRegistry,
		busProvider,
		false,
	)
	if err != nil {
		t.Fatalf("Failed to create mcp server: %v", err)
	}

	// Create a large body (> 5MB)
	// 5MB = 5 * 1024 * 1024 = 5,242,880 bytes
	// We'll use 6MB
	largeBody := strings.Repeat("a", 6*1024*1024)
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
		"params":  map[string]interface{}{"padding": largeBody},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	// Call the function with 5MB limit
	handled := handleStatelessJSONRPC(mcpSrv, w, req, 5<<20)

	if !handled {
		t.Errorf("Expected request to be handled")
	}

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status code 413, got %d", w.Code)
	}
}

func TestHandleStatelessJSONRPC_Normal(t *testing.T) {
	// Setup dependencies
	busProvider, _ := bus.NewProvider(nil)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager)

	serviceRegistry := serviceregistry.New(
		upstreamFactory,
		toolManager,
		promptManager,
		resourceManager,
		authManager,
	)

	mcpSrv, err := mcpserver.NewServer(
		context.Background(),
		toolManager,
		promptManager,
		resourceManager,
		authManager,
		serviceRegistry,
		busProvider,
		false,
	)
	if err != nil {
		t.Fatalf("Failed to create mcp server: %v", err)
	}

	// Normal body
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	// Call the function with 5MB limit
	handled := handleStatelessJSONRPC(mcpSrv, w, req, 5<<20)

	if !handled {
		t.Errorf("Expected request to be handled")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}
}
