package mcpserver

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

// The go-sdk/mcp package does not easily allow mocking ServerSession because it is a struct, not an interface.
// However, we can test the nil session checks.

func TestMCPSession_NilSession(t *testing.T) {
	s := NewMCPSession(nil)

	t.Run("CreateMessage with nil session", func(t *testing.T) {
		res, err := s.CreateMessage(context.Background(), &mcp.CreateMessageParams{})
		assert.Error(t, err)
		assert.Equal(t, "no active session available for sampling", err.Error())
		assert.Nil(t, res)
	})

	t.Run("ListRoots with nil session", func(t *testing.T) {
		res, err := s.ListRoots(context.Background())
		assert.Error(t, err)
		assert.Equal(t, "no active session available for roots inspection", err.Error())
		assert.Nil(t, res)
	})
}

func TestNewMCPSampler(t *testing.T) {
	s := NewMCPSampler(nil)
	assert.NotNil(t, s)
}
