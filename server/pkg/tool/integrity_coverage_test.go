package tool

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestIntegrityCoverage(t *testing.T) {
	t.Run("VerifyConfigIntegrity", func(t *testing.T) {
		// Currently a placeholder returning nil
		err := VerifyConfigIntegrity(&configv1.ToolDefinition{})
		assert.NoError(t, err)
	})

	t.Run("CalculateConfigHash", func(t *testing.T) {
		toolDef := &configv1.ToolDefinition{
			Name: proto.String("test"),
		}
		hash, err := CalculateConfigHash(toolDef)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)

		// Ensure determinism
		hash2, err := CalculateConfigHash(toolDef)
		require.NoError(t, err)
		assert.Equal(t, hash, hash2)
	})

	t.Run("VerifyIntegrity Unsupported Algorithm", func(t *testing.T) {
		tool := &v1.Tool{
			Integrity: &v1.ToolIntegrity{
				Algorithm: proto.String("md5"),
				Hash:      proto.String("123"),
			},
		}
		err := VerifyIntegrity(tool)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported integrity algorithm")
	})
}
