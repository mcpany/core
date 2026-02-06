package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUpstream_Uniqueness(t *testing.T) {
	u1 := NewUpstream(nil)
	u2 := NewUpstream(nil)

	// Check if the upstream instances are distinct pointers
	assert.NotSame(t, u1, u2, "NewUpstream should return distinct instances")

	upstream1, ok1 := u1.(*Upstream)
	upstream2, ok2 := u2.(*Upstream)

	assert.True(t, ok1)
	assert.True(t, ok2)

	// Check if sessionRegistry pointers are distinct
	assert.NotSame(t, upstream1.sessionRegistry, upstream2.sessionRegistry, "SessionRegistry should be distinct per upstream")
}
