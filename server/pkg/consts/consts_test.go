package consts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsts(t *testing.T) {
	assert.NotEmpty(t, HeaderMcpSessionID)
}
