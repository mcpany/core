
package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetup_NoErrorOnValidRedis(t *testing.T) {
	// Set a valid REDIS_ADDR
	os.Setenv("REDIS_ADDR", "localhost:6379")
	defer os.Unsetenv("REDIS_ADDR")

	_, err := setup()
	assert.NoError(t, err, "setup() should not return an error with a valid REDIS_ADDR")
}
