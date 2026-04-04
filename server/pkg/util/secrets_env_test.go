package util

import (
	"context"
	"testing"
	"os"

	"github.com/stretchr/testify/assert"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestEnvVar_NotSet_Coverage(t *testing.T) {
	t.Run("Environment Variable Not Set", func(t *testing.T) {
		t.Setenv("MCPANY_ALLOW_RESTRICTED_ENV_VARS", "true")
		os.Unsetenv("SOME_RANDOM_MISSING_VAR")

		secret := configv1.SecretValue_builder{
			EnvironmentVariable: proto.String("SOME_RANDOM_MISSING_VAR"),
		}.Build()

		_, err := ResolveSecret(context.Background(), secret)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "environment variable \"SOME_RANDOM_MISSING_VAR\" is not set")
		}
	})
}
