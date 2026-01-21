// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestContextAwareSuggestions(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Case 1: Typo in GlobalSettings. 'adress' is NOT valid, and NOT close to any valid field in GlobalSettings.
	// Should return error but NO suggestion of 'address' (which belongs to http_service).
	yamlContent1 := `
global_settings:
  adress: "something"
`
	_ = afero.WriteFile(fs, "config_bad_global.yaml", []byte(yamlContent1), 0644)

	store1 := config.NewFileStore(fs, []string{"config_bad_global.yaml"})
	_, err1 := store1.Load(context.Background())

	assert.Error(t, err1)
	if err1 != nil {
		assert.Contains(t, err1.Error(), `unknown field "global_settings.adress" in GlobalSettings`)
		assert.NotContains(t, err1.Error(), `Did you mean "address"?`) // Should NOT suggest address
	}

	// Case 2: Typo in GlobalSettings. 'log_lvl' IS close to 'log_level'.
	// Should return error AND suggestion.
	yamlContent2 := `
global_settings:
  log_lvl: "debug"
`
	_ = afero.WriteFile(fs, "config_typo_global.yaml", []byte(yamlContent2), 0644)
	store2 := config.NewFileStore(fs, []string{"config_typo_global.yaml"})
	_, err2 := store2.Load(context.Background())

	assert.Error(t, err2)
	if err2 != nil {
		assert.Contains(t, err2.Error(), `unknown field "global_settings.log_lvl" in GlobalSettings`)
		assert.Contains(t, err2.Error(), `Did you mean "log_level"?`)
	}

	// Case 3: Typo in HttpService. 'adress' -> 'address'.
	// Should suggest 'address'.
	yamlContent3 := `
upstream_services:
  - name: "weather"
    http_service:
      adress: "https://api.weather.com"
`
	_ = afero.WriteFile(fs, "config_typo_http.yaml", []byte(yamlContent3), 0644)
	store3 := config.NewFileStore(fs, []string{"config_typo_http.yaml"})
	_, err3 := store3.Load(context.Background())

	assert.Error(t, err3)
	if err3 != nil {
		assert.Contains(t, err3.Error(), `unknown field "upstream_services[0].http_service.adress" in HttpUpstreamService`)
		assert.Contains(t, err3.Error(), `Did you mean "address"?`)
	}

	// Case 4: Typo in JSON.
	jsonContent := `{
		"global_settings": {
			"log_lvl": "debug"
		}
	}`
	_ = afero.WriteFile(fs, "config_typo.json", []byte(jsonContent), 0644)
	store4 := config.NewFileStore(fs, []string{"config_typo.json"})
	_, err4 := store4.Load(context.Background())

	assert.Error(t, err4)
	if err4 != nil {
		assert.Contains(t, err4.Error(), `unknown field "global_settings.log_lvl" in GlobalSettings`)
		assert.Contains(t, err4.Error(), `Did you mean "log_level"?`)
	}
}
