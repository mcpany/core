package util_test

import (
	"os"
	"testing"

	"github.com/mcpany/core/pkg/util"
	"github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestResolveSecret_PlainText(t *testing.T) {
	secret := &v1.SecretValue{}
	secret.SetPlainText("test-value")

	resolved, err := util.ResolveSecret(secret)
	assert.NoError(t, err)
	assert.Equal(t, "test-value", resolved)
}

func TestResolveSecret_EnvironmentVariable(t *testing.T) {
	os.Setenv("TEST_SECRET", "test-value")
	defer os.Unsetenv("TEST_SECRET")

	secret := &v1.SecretValue{}
	secret.SetEnvironmentVariable("TEST_SECRET")

	resolved, err := util.ResolveSecret(secret)
	assert.NoError(t, err)
	assert.Equal(t, "test-value", resolved)
}

func TestResolveSecret_FilePath(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "secret-")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString("test-value")
	assert.NoError(t, err)
	tmpfile.Close()

	secret := &v1.SecretValue{}
	secret.SetFilePath(tmpfile.Name())

	resolved, err := util.ResolveSecret(secret)
	assert.NoError(t, err)
	assert.Equal(t, "test-value", resolved)
}
