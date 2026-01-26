package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetStringSlice_CommaInPath(t *testing.T) {
	viper.Reset()

	// Simulate an environment variable setting a path with a comma, escaped.
	// We use `\,` to indicate a literal comma.
	// Since we are in Go string literal, we need `\\,`.
	viper.Set("config-path", "/path/with\\,comma/config.yaml")

	paths := getStringSlice("config-path")

	// We expect the path to be preserved and the backslash removed.
	assert.Equal(t, []string{"/path/with,comma/config.yaml"}, paths, "Should preserve path with escaped comma")
}

func TestGetStringSlice_MultiplePaths(t *testing.T) {
	viper.Reset()
	// Simulate "path1, path2"
	viper.Set("config-path", "path1, path2")

	paths := getStringSlice("config-path")

	assert.Equal(t, []string{"path1", "path2"}, paths, "Should split multiple paths by comma")
}

func TestGetStringSlice_DoubleBackslash(t *testing.T) {
	viper.Reset()
	// Simulate "path\\with\\backslashes"
	// We want literal backslash. `\\` should become `\`.
	// In Go string: `path\\\\with\\\\backslashes` represents `path\\with\\backslashes` in env.
	viper.Set("config-path", "path\\\\with\\\\backslashes")

	paths := getStringSlice("config-path")

	assert.Equal(t, []string{"path\\with\\backslashes"}, paths, "Should treat double backslash as single backslash")
}

func TestGetStringSlice_Mixed(t *testing.T) {
	viper.Reset()
	// "path1, path2\\,with\\,comma, path3"
	// Expect: ["path1", "path2,with,comma", "path3"]
	viper.Set("config-path", "path1, path2\\,with\\,comma, path3")

	paths := getStringSlice("config-path")

	assert.Equal(t, []string{"path1", "path2,with,comma", "path3"}, paths, "Should handle mixed escaped and unescaped commas")
}
