package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
    "github.com/spf13/afero"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "google.golang.org/protobuf/proto"
)

func TestSettingsCoverage(t *testing.T) {
	s := GlobalSettings()

    // SetMiddlewares
    mw := []*configv1.Middleware{{Name: proto.String("mw1")}}
    s.SetMiddlewares(mw)
    assert.Equal(t, mw, s.Middlewares())

    // SetDlp
    dlp := &configv1.DLPConfig{Enabled: proto.Bool(true)}
    s.SetDlp(dlp)
    assert.Equal(t, dlp, s.GetDlp())

    // GetOidc
    oidc := &configv1.OIDCConfig{Issuer: proto.String("issuer")}
    s.proto.Oidc = oidc
    assert.Equal(t, oidc, s.GetOidc())

    // GetProfileDefinitions
    pd := []*configv1.ProfileDefinition{{Name: proto.String("pd1")}}
    s.proto.ProfileDefinitions = pd
    assert.Equal(t, pd, s.GetProfileDefinitions())

    // GithubAPIURL
    url := "https://api.github.com"
    s.proto.GithubApiUrl = proto.String(url)
    assert.Equal(t, url, s.GithubAPIURL())
}

func TestSettingsLoadCoverage(t *testing.T) {
    // Test loading from config file to cover s.proto.SetMcpListenAddress and Middleware logic in Load

    content := `
global_settings: {
    mcp_listen_address: ":12345"
    middlewares: [
        { name: "test-mw", disabled: false }
    ]
}
`
    tmpDir := t.TempDir()
    fs := afero.NewOsFs()
    f, _ := fs.Create(tmpDir + "/config.textproto")
    f.WriteString(content)
    f.Close()

    viper.Set("config-path", tmpDir + "/config.textproto")

    cmd := &cobra.Command{}
    s := &Settings{proto: &configv1.GlobalSettings{}}

    err := s.Load(cmd, fs)
    assert.NoError(t, err)

    assert.Equal(t, ":12345", s.MCPListenAddress())
    assert.Len(t, s.Middlewares(), 1)
}
