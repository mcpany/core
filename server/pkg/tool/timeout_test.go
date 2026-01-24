package tool

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

type MockTimeoutTool struct {
	tool *pb.Tool
}

func (m *MockTimeoutTool) Tool() *pb.Tool { return m.tool }
func (m *MockTimeoutTool) MCPTool() *mcp.Tool { return nil }
func (m *MockTimeoutTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		return "no deadline", nil
	}
	return time.Until(deadline), nil
}
func (m *MockTimeoutTool) GetCacheConfig() *configv1.CacheConfig { return nil }

func TestExecuteTool_Timeout(t *testing.T) {
	b, _ := bus.NewProvider(nil)
	tm := NewManager(b)

	timeout := 100 * time.Millisecond
	toolDef := &pb.Tool{
		Name:      proto.String("test-tool"),
		ServiceId: proto.String("s1"),
		Timeout:   durationpb.New(timeout),
	}

	mockTool := &MockTimeoutTool{tool: toolDef}
	if err := tm.AddTool(mockTool); err != nil {
		t.Fatalf("AddTool failed: %v", err)
	}

	tm.AddServiceInfo("s1", &ServiceInfo{HealthStatus: "healthy"})

	// Tool name exposed is "s1.test-tool"
	res, err := tm.ExecuteTool(context.Background(), &ExecutionRequest{ToolName: "s1.test-tool"})
	if err != nil {
		t.Fatalf("ExecuteTool failed: %v", err)
	}

	dur, ok := res.(time.Duration)
	if !ok {
		t.Fatalf("Expected duration result, got %T: %v", res, res)
	}

	if dur > timeout {
		t.Errorf("Remaining duration %v > timeout %v", dur, timeout)
	}
}
