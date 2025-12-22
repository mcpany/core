package appconsts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppConsts(t *testing.T) {
	assert.NotEmpty(t, Name)
	assert.NotEmpty(t, Version)
}
