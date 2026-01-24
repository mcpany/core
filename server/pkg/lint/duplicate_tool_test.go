package lint

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestCheckDuplicateTools(t *testing.T) {
	cfg := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("s1"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Tools: []*configv1.ToolDefinition{
							{Name: proto.String("tool1")},
						},
					},
				},
			},
			{
				Name: proto.String("s2"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Tools: []*configv1.ToolDefinition{
							{Name: proto.String("tool1")}, // Duplicate
						},
					},
				},
			},
		},
	}

	linter := NewLinter(cfg)
	results, _ := linter.Run(context.Background())
	found := false
	for _, r := range results {
		if r.Severity == Warning && strings.Contains(r.Message, "Duplicate tool name") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected duplicate tool warning")
	}
}
