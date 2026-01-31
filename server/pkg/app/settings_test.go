// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"sync"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestNewGlobalSettingsManager(t *testing.T) {
	t.Run("Initialize with values", func(t *testing.T) {
		apiKey := "test-key"
		ips := []string{"127.0.0.1"}
		origins := []string{"http://localhost:3000"}

		m := NewGlobalSettingsManager(apiKey, ips, origins)

		assert.Equal(t, apiKey, m.GetAPIKey())
		assert.Equal(t, ips, m.GetAllowedIPs())
		assert.Equal(t, origins, m.GetAllowedOrigins())
	})

	t.Run("Initialize with nil origins", func(t *testing.T) {
		m := NewGlobalSettingsManager("key", nil, nil)

		assert.Equal(t, "key", m.GetAPIKey())
		assert.Nil(t, m.GetAllowedIPs())
		assert.Equal(t, []string{}, m.GetAllowedOrigins())
	})
}

func TestGlobalSettingsManager_Update(t *testing.T) {
	t.Run("API Key Priority - Explicit overrides Config", func(t *testing.T) {
		m := NewGlobalSettingsManager("initial", nil, nil)
		settings := configv1.GlobalSettings_builder{
			ApiKey: proto.String("config-key"),
		}.Build()

		// Explicit key provided
		m.Update(settings, "explicit-key")
		assert.Equal(t, "explicit-key", m.GetAPIKey())
	})

	t.Run("API Key Priority - Config used when Explicit is empty", func(t *testing.T) {
		m := NewGlobalSettingsManager("initial", nil, nil)
		settings := configv1.GlobalSettings_builder{
			ApiKey: proto.String("config-key"),
		}.Build()

		// No explicit key
		m.Update(settings, "")
		assert.Equal(t, "config-key", m.GetAPIKey())
	})

	t.Run("Allowed IPs Update", func(t *testing.T) {
		m := NewGlobalSettingsManager("", nil, nil)
		settings := configv1.GlobalSettings_builder{
			AllowedIps: []string{"10.0.0.1", "10.0.0.2"},
		}.Build()

		m.Update(settings, "")
		assert.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, m.GetAllowedIPs())
	})

	t.Run("Allowed Origins Update - Normal", func(t *testing.T) {
		m := NewGlobalSettingsManager("", nil, nil)
		settings := configv1.GlobalSettings_builder{
			AllowedOrigins: []string{"https://example.com"},
			LogLevel:       configv1.GlobalSettings_LOG_LEVEL_INFO.Enum(),
		}.Build()

		m.Update(settings, "")
		assert.Equal(t, []string{"https://example.com"}, m.GetAllowedOrigins())
	})

	t.Run("Allowed Origins Update - Debug Defaults", func(t *testing.T) {
		m := NewGlobalSettingsManager("", nil, nil)
		// Empty origins + Debug level
		settings := configv1.GlobalSettings_builder{
			AllowedOrigins: []string{},
			LogLevel:       configv1.GlobalSettings_LOG_LEVEL_DEBUG.Enum(),
		}.Build()

		m.Update(settings, "")
		assert.Equal(t, []string{"*"}, m.GetAllowedOrigins())
	})

	t.Run("Allowed Origins Update - Info No Defaults", func(t *testing.T) {
		m := NewGlobalSettingsManager("", nil, nil)
		// Empty origins + Info level
		settings := configv1.GlobalSettings_builder{
			AllowedOrigins: []string{},
			LogLevel:       configv1.GlobalSettings_LOG_LEVEL_INFO.Enum(),
		}.Build()

		m.Update(settings, "")
		assert.Empty(t, m.GetAllowedOrigins())
		assert.NotEqual(t, []string{"*"}, m.GetAllowedOrigins())
	})

	t.Run("Nil Settings", func(t *testing.T) {
		m := NewGlobalSettingsManager("initial", []string{"1.1.1.1"}, []string{"*"})

		// Update with nil settings, should rely on explicit key and set others to nil/defaults
		m.Update(nil, "new-key")

		assert.Equal(t, "new-key", m.GetAPIKey())
		assert.Nil(t, m.GetAllowedIPs())
		assert.Nil(t, m.GetAllowedOrigins())
	})
}

func TestGlobalSettingsManager_Concurrency(t *testing.T) {
	m := NewGlobalSettingsManager("initial", nil, nil)
	settings := configv1.GlobalSettings_builder{
		ApiKey:         proto.String("config-key"),
		AllowedIps:     []string{"127.0.0.1"},
		AllowedOrigins: []string{"*"},
	}.Build()

	var wg sync.WaitGroup
	iterations := 1000

	// Writer goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			// Toggle between explicit and config key
			key := ""
			if i%2 == 0 {
				key = "explicit"
			}
			m.Update(settings, key)
		}
	}()

	// Reader goroutines
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = m.GetAPIKey()
				_ = m.GetAllowedIPs()
				_ = m.GetAllowedOrigins()
			}
		}()
	}

	wg.Wait()
}
