package tool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
	"github.com/mcpany/core/server/pkg/pool"
)

func TestCheckUnquotedKeywords_ExtraCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		val      string
		keywords []string
		wantErr  bool
	}{
		{
			name:     "unquoted keyword",
			val:      "hello system world",
			keywords: []string{"system"},
			wantErr:  true,
		},
		{
			name:     "escaped keyword",
			val:      "hello \\system world",
			keywords: []string{"system"},
			wantErr:  false,
		},
		{
			name:     "single quoted keyword",
			val:      "hello 'system' world",
			keywords: []string{"system"},
			wantErr:  false,
		},
		{
			name:     "double quoted keyword",
			val:      "hello \"system\" world",
			keywords: []string{"system"},
			wantErr:  false,
		},
		{
			name:     "backtick quoted keyword",
			val:      "hello `system` world",
			keywords: []string{"system"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := checkUnquotedKeywords(tt.val, tt.keywords)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckAwkInjection_ExtraCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		val     string
		base    string
		wantErr bool
	}{
		{
			name:    "not awk",
			val:     "|",
			base:    "echo",
			wantErr: false,
		},
		{
			name:    "awk pipe",
			val:     "print | something",
			base:    "awk",
			wantErr: true,
		},
		{
			name:    "gawk redirect in",
			val:     "< file",
			base:    "gawk",
			wantErr: true,
		},
		{
			name:    "nawk redirect out",
			val:     "> file",
			base:    "nawk",
			wantErr: true,
		},
		{
			name:    "mawk getline",
			val:     "getline",
			base:    "mawk",
			wantErr: true,
		},
		{
			name:    "awk indirect func",
			val:     "@fun",
			base:    "awk",
			wantErr: true,
		},
		{
			name:    "awk include",
			val:     "@include",
			base:    "awk",
			wantErr: true,
		},
		{
			name:    "awk load",
			val:     "@load",
			base:    "awk",
			wantErr: true,
		},
		{
			name:    "awk append",
			val:     ">>",
			base:    "awk",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := checkAwkInjection(tt.val, tt.base)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckNodePerlPhpInjection_ExtraCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		val        string
		base       string
		quoteLevel int
		wantErr    bool
	}{
		{
			name:       "not node/perl/php",
			val:        "${injection}",
			base:       "python",
			quoteLevel: 1, // Double quotes
			wantErr:    false,
		},
		{
			name:       "node unquoted",
			val:        "${injection}",
			base:       "node",
			quoteLevel: 0,
			wantErr:    false,
		},
		{
			name:       "node backticks",
			val:        "${injection}",
			base:       "node",
			quoteLevel: 3, // backticks
			wantErr:    true,
		},
		{
			name:       "bun backticks",
			val:        "${injection}",
			base:       "bun",
			quoteLevel: 3,
			wantErr:    true,
		},
		{
			name:       "deno backticks",
			val:        "${injection}",
			base:       "deno",
			quoteLevel: 3,
			wantErr:    true,
		},
		{
			name:       "perl double quotes",
			val:        "${injection}",
			base:       "perl",
			quoteLevel: 1, // double quotes
			wantErr:    true,
		},
		{
			name:       "php double quotes",
			val:        "${injection}",
			base:       "php",
			quoteLevel: 1, // double quotes
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := checkNodePerlPhpInjection(tt.val, tt.base, tt.quoteLevel)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGRPCTool_Coverage(t *testing.T) {
	t.Parallel()

	toolProto := v1.Tool_builder{
		Name: proto.String("test-grpc-tool"),
	}.Build()

	methodDesc := new(MockMethodDescriptor)
	msgDesc := new(MockMessageDescriptor)
	msgDesc.On("FullName").Return(protoreflect.FullName("TestMessage"))
	methodDesc.On("Input").Return(msgDesc)
	methodDesc.On("FullName").Return(protoreflect.FullName("TestService.TestMethod"))

	poolManager := pool.NewManager()
	callDef := &configv1.GrpcCallDefinition{}
	resilienceCfg := &configv1.ResilienceConfig{}

	grpcTool := NewGRPCTool(toolProto, poolManager, "test-service", methodDesc, callDef, resilienceCfg)

	// Test IsStreaming
	assert.False(t, grpcTool.IsStreaming())

	// Test StreamExecute
	ch, err := grpcTool.StreamExecute(context.Background(), &ExecutionRequest{})
	assert.NoError(t, err)
	// it should execute and return an error (due to pool manager error, etc.)
	select {
	case res := <-ch:
		errRes, ok := res.(error)
		assert.True(t, ok)
		assert.Error(t, errRes)
	case <-time.After(time.Second):
		t.Fatal("StreamExecute did not return anything")
	}

	// Test MCPTool (it should at least initialize and return a non-nil object or gracefully handle if tool is incomplete)
	mcpTool := grpcTool.MCPTool()
	assert.NotNil(t, mcpTool)
}
