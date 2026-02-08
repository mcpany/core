package config

import (
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func ptr[T any](v T) *T {
	return &v
}

func TestSettings_Getters(t *testing.T) {
	s := &Settings{
		proto: func() *configv1.GlobalSettings {
			return configv1.GlobalSettings_builder{
				DbDsn:    proto.String("postgres://user:pass@127.0.0.1:5432/db"),
				DbDriver: proto.String("postgres"),
				Dlp: configv1.DLPConfig_builder{
					Enabled: proto.Bool(true),
				}.Build(),
			}.Build()
		}(),
	}

	assert.Equal(t, "postgres://user:pass@127.0.0.1:5432/db", s.GetDbDsn())
	assert.Equal(t, "postgres", s.GetDbDriver())
	assert.NotNil(t, s.GetDlp())
	assert.True(t, s.GetDlp().GetEnabled())
}

func TestSettings_Load_DbSettings(t *testing.T) {
	viper.Reset()
	fs := afero.NewMemMapFs()
	cmd := &cobra.Command{}

	viper.Set("db-dsn", "sqlite://file.db")
	viper.Set("db-driver", "sqlite3")

	settings := &Settings{
		proto: configv1.GlobalSettings_builder{}.Build(),
	}
	err := settings.Load(cmd, fs)
	require.NoError(t, err)

	assert.Equal(t, "sqlite://file.db", settings.GetDbDsn())
	assert.Equal(t, "sqlite3", settings.GetDbDriver())
}

func TestGetStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		setup    func()
		expected []string
	}{
		{
			name: "env var comma separated",
			key:  "test-env-slice",
			setup: func() {
				// Simulating env var by setting string in viper
				viper.Set("test-env-slice", "a,b, c")
			},
			expected: []string{"a", "b", "c"},
		},
		{
			name: "env var single value",
			key:  "test-env-single",
			setup: func() {
				viper.Set("test-env-single", "single")
			},
			expected: []string{"single"},
		},
		{
			name: "viper string slice",
			key:  "test-viper-slice",
			setup: func() {
				viper.Set("test-viper-slice", []string{"x", "y", "z"})
			},
			expected: []string{"x", "y", "z"},
		},
		{
			name: "viper string slice with commas",
			key:  "test-viper-slice-commas",
			setup: func() {
				viper.Set("test-viper-slice-commas", []string{"1, 2", "3"})
			},
			expected: []string{"1", "2", "3"},
		},
		{
			name: "empty",
			key:  "test-empty",
			setup: func() {
				viper.Set("test-empty", "")
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			tt.setup()
			res := getStringSlice(tt.key)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestSettings_Load_StringSliceEnv(t *testing.T) {
	// Integration test for Load using getStringSlice logic via viper
	viper.Reset()
	fs := afero.NewMemMapFs()
	cmd := &cobra.Command{}

	// Simulate Env vars
	viper.Set("config-path", "/path/1.textproto, /path/2.textproto")
	viper.Set("profiles", "p1, p2")

	// Create dummy config file to avoid Load error
	// We need to provide a service so that LoadServices doesn't fail with "empty config"
	content := `
upstream_services: {
	name: "dummy-service"
	http_service: {
		address: "http://example.com"
	}
}
`
	err := afero.WriteFile(fs, "/path/1.textproto", []byte(content), 0o644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, "/path/2.textproto", []byte(content), 0o644)
	require.NoError(t, err)

	// Also we need to ensure logging doesn't panic or fail.
	tmpLog, err := os.CreateTemp("", "app.log")
	require.NoError(t, err)
	defer os.Remove(tmpLog.Name())
	viper.Set("logfile", tmpLog.Name())

	middlewares := []*configv1.Middleware{
		func() *configv1.Middleware {
			return configv1.Middleware_builder{
				Name: proto.String("test-middleware"),
			}.Build()
		}(),
	}

	settings := &Settings{
		dbPath: "/path/to/db.sqlite",
		proto: func() *configv1.GlobalSettings {
			return configv1.GlobalSettings_builder{
				Middlewares: middlewares,
			}.Build()
		}(),
	}

	err = settings.Load(cmd, fs)
	require.NoError(t, err)

	assert.Equal(t, []string{"/path/1.textproto", "/path/2.textproto"}, settings.ConfigPaths())
	assert.Equal(t, []string{"p1", "p2"}, settings.Profiles())
}
