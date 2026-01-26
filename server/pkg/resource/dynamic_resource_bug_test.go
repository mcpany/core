package resource

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestDynamicResource_Read_Bug_Primitives(t *testing.T) {
	def := configv1.ResourceDefinition_builder{
		Uri: proto.String("test-uri"),
	}.Build()

	tests := []struct {
		name     string
		retVal   interface{}
		expected string
	}{
		{"int", 123, "123"},
		{"float", 12.34, "12.34"},
		{"bool", true, "true"},
		{"slice", []interface{}{"a", 1}, `["a",1]`},
		// nil marshal to "null"
		{"nil", nil, "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTool := new(MockTool)
			mockTool.On("Execute", mock.Anything, mock.Anything).Return(tt.retVal, nil)
			mockTool.On("Tool").Return(v1.Tool_builder{ServiceId: proto.String("test-service")}.Build())
			dr, _ := NewDynamicResource(def, mockTool)
			result, err := dr.Read(context.Background())

			// Currently this fails
			require.NoError(t, err)
			require.Len(t, result.Contents, 1)
			assert.Equal(t, tt.expected, result.Contents[0].Text)
		})
	}
}
