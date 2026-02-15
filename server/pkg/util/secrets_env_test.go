package util

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestResolveSecret_EnvVar_Empty(t *testing.T) {
	key := "TEST_EMPTY_SECRET_BUG"
	// Set the environment variable to an empty string.
	// This is a valid state (the variable exists, but its value is empty).
	os.Setenv(key, "")
	defer os.Unsetenv(key)

	secret := configv1.SecretValue_builder{
		EnvironmentVariable: proto.String(key),
	}.Build()

	// We expect ResolveSecret to respect that the variable IS set (albeit empty).
	// Currently, it likely returns an error claiming "is not set".
	val, err := ResolveSecret(context.Background(), secret)

	assert.NoError(t, err, "Should not return error for empty env var if it is explicitly set")
	assert.Equal(t, "", val, "Should return empty string for empty env var")
}
