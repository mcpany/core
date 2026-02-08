package config

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestReproduction_SilentFailure_BadConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create a malformed YAML file
	_ = afero.WriteFile(fs, "bad.yaml", []byte("upstream_services:\n  - name: test\n    http_service:\n      address: http://127.0.0.1\n  indentation_error"), 0644)

	// Use the lenient store, as used in server.go
	store := NewFileStoreWithSkipErrors(fs, []string{"bad.yaml"})

	// Load should NOT return an error, effectively swallowing the parse error
	cfg, err := store.Load(context.Background())

	// We expect NO error because skipErrors=true
	assert.NoError(t, err, "Load should not return error when skipErrors is true")

	// And config should be empty or partial
	assert.Empty(t, cfg.GetUpstreamServices(), "Config should be empty because the file failed to parse")
}

func TestReproduction_ClaudeConfig_HelpfulError(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create a file mimicking Claude Desktop config
	claudeConfig := `
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/Users/me/Desktop"]
    }
  }
}
`
	_ = afero.WriteFile(fs, "claude_config.json", []byte(claudeConfig), 0644)

	// 1. Test with Lenient Store (Current Behavior)
	storeLenient := NewFileStoreWithSkipErrors(fs, []string{"claude_config.json"})
	_, err := storeLenient.Load(context.Background())
	assert.NoError(t, err, "Lenient store should swallow the error")

	// 2. Test with Strict Store (Desired Behavior)
	storeStrict := NewFileStore(fs, []string{"claude_config.json"})
	_, errStrict := storeStrict.Load(context.Background())

	assert.Error(t, errStrict)
	// Verify the helpful message is present
	assert.True(t, strings.Contains(errStrict.Error(), "Did you mean \"upstream_services\"?"), "Error message should contain helpful hint")
}

func TestReproduction_Typo_UnknownField_HelpfulError(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Config with a typo: "target_address" instead of "address"
	typoConfig := `
upstream_services:
  - name: "my-service"
    http_service:
      target_address: "http://127.0.0.1:8080"
`
	_ = afero.WriteFile(fs, "typo_config.yaml", []byte(typoConfig), 0644)

	store := NewFileStore(fs, []string{"typo_config.yaml"})
	_, err := store.Load(context.Background())

	assert.Error(t, err)
	// Now we expect a suggestion.
	// It might suggest "address" or "http_address" depending on Levenshtein distance.
	// Both are valid suggestions relative to "target_address".
	hasSuggestion := strings.Contains(err.Error(), "Did you mean \"address\"?") || strings.Contains(err.Error(), "Did you mean \"http_address\"?")
	assert.True(t, hasSuggestion, "Error message should contain a helpful suggestion")
	t.Logf("Error: %v", err)
}

func TestReproduction_YamlSyntaxError_HelpfulMessage(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Config with indentation error
	// "calls" is indented too far, making it look like it's inside "address" value (which is impossible)
	// or just invalid structure for list.
	// Actually, let's use a TAB character which is forbidden in YAML.
	badYaml := "upstream_services:\n\t- name: test"
	_ = afero.WriteFile(fs, "tab_config.yaml", []byte(badYaml), 0644)

	store := NewFileStore(fs, []string{"tab_config.yaml"})
	_, err := store.Load(context.Background())

	assert.Error(t, err)
	t.Logf("Error: %v", err)
	// Check if the error is reasonably descriptive
	assert.True(t, strings.Contains(err.Error(), "found character that cannot start any token"), "Should mention tab error or token error")
	assert.True(t, strings.Contains(err.Error(), "Hint: YAML files cannot contain tabs"), "Should provide hint about tabs")
}
