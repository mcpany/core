package prompt

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestMCPServerProvider(t *testing.T) {
	server := &mcp.Server{}
	provider := NewMCPServerProvider(server)

	assert.Equal(t, server, provider.Server())
}
