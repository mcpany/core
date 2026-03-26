// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"sync"
	"testing"

	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestNewGlobalSettingsManager(t *testing.T) {
	apiKey := "test-key"
	allowedIPs := []string{"127.0.0.1"}
	allowedOrigins := []string{"example.com"}

	m := NewGlobalSettingsManager(apiKey, allowedIPs, allowedOrigins)

	assert.Equal(t, apiKey, m.GetAPIKey())
	assert.Equal(t, allowedIPs, m.GetAllowedIPs())
	assert.Equal(t, allowedOrigins, m.GetAllowedOrigins())
}

func TestNewGlobalSettingsManager_Defaults(t *testing.T) {
	m := NewGlobalSettingsManager("", nil, nil)
	assert.Empty(t, m.GetAPIKey())
	assert.Nil(t, m.GetAllowedIPs())                   // It stores what we give it. nil -> nil
	assert.Equal(t, []string{}, m.GetAllowedOrigins()) // logic says if allowedOrigins == nil, make it []string{}
}

func TestGlobalSettingsManager_Update(t *testing.T) {
	m := NewGlobalSettingsManager("initial-key", nil, nil)

	// Case 1: Explicit API Key overrides Config
	settings := config_v1.GlobalSettings_builder{
		ApiKey: proto.String("config-key"),
	}.Build()
	m.Update(settings, "explicit-key")
	assert.Equal(t, "explicit-key", m.GetAPIKey())

	// Case 2: Config API Key used if explicit is empty
	m.Update(settings, "")
	assert.Equal(t, "config-key", m.GetAPIKey())

	// Case 3: Allowed IPs update
	settings = config_v1.GlobalSettings_builder{
		AllowedIps: []string{"10.0.0.1"},
	}.Build()
	m.Update(settings, "")
	assert.Equal(t, []string{"10.0.0.1"}, m.GetAllowedIPs())

	// Case 4: Allowed Origins update
	settings = config_v1.GlobalSettings_builder{
		AllowedOrigins: []string{"foo.com"},
	}.Build()
	m.Update(settings, "")
	assert.Equal(t, []string{"foo.com"}, m.GetAllowedOrigins())

	// Case 5: Debug Mode Default Origins
	debugLevel := config_v1.GlobalSettings_LOG_LEVEL_DEBUG
	settings = config_v1.GlobalSettings_builder{
		LogLevel: &debugLevel,
	}.Build()
	m.Update(settings, "")
	assert.Equal(t, []string{"*"}, m.GetAllowedOrigins())

	// Case 6: Non-Debug Mode with empty origins
	infoLevel := config_v1.GlobalSettings_LOG_LEVEL_INFO
	settings = config_v1.GlobalSettings_builder{
		LogLevel: &infoLevel,
	}.Build()
	m.Update(settings, "")
	assert.Empty(t, m.GetAllowedOrigins())
}

func TestGlobalSettingsManager_Concurrency(t *testing.T) {
	m := NewGlobalSettingsManager("initial", nil, nil)
	var wg sync.WaitGroup

	// Writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			settings := config_v1.GlobalSettings_builder{
				ApiKey:     proto.String("updated-key"),
				AllowedIps: []string{"1.1.1.1"},
			}.Build()
			m.Update(settings, "")
		}
	}()

	// Readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				m.GetAPIKey()
				m.GetAllowedIPs()
				m.GetAllowedOrigins()
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, "updated-key", m.GetAPIKey())
}
