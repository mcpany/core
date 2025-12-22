package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestConfigProto(t *testing.T) {
	cfg := &McpAnyServerConfig{}
	assert.NotNil(t, cfg)
	assert.Equal(t, "", cfg.String()) // Empty config should match empty string rep or similar?
	// Actually String() usually produces something. Just smoke test methods.
	assert.NotNil(t, cfg.ProtoReflect())

	// Marshaling
	b, err := proto.Marshal(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, b)
}
