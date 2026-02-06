// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ProviderType represents the type of LLM provider.
type ProviderType string

const (
	// ProviderOpenAI represents OpenAI.
	ProviderOpenAI ProviderType = "openai"
	// ProviderClaude represents Anthropic Claude.
	ProviderClaude ProviderType = "claude"
	// ProviderGemini represents Google Gemini.
	ProviderGemini ProviderType = "gemini"
)

// ProviderConfig holds configuration for an LLM provider.
type ProviderConfig struct {
	Type    ProviderType `json:"type"`
	APIKey  string       `json:"apiKey"`
	BaseURL string       `json:"baseUrl,omitempty"`
	Model   string       `json:"model,omitempty"` // Default model to use
}

// ProviderStore manages storage of provider configurations.
type ProviderStore struct {
	mu       sync.RWMutex
	filePath string
	configs  map[ProviderType]ProviderConfig
}

// NewProviderStore creates a new ProviderStore.
func NewProviderStore(dataDir string) (*ProviderStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	store := &ProviderStore{
		filePath: filepath.Join(dataDir, "llm_providers.json"),
		configs:  make(map[ProviderType]ProviderConfig),
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

// SaveConfig saves a provider configuration.
func (s *ProviderStore) SaveConfig(config ProviderConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.configs[config.Type] = config
	return s.save()
}

// GetConfig returns a provider configuration.
func (s *ProviderStore) GetConfig(pType ProviderType) (ProviderConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, ok := s.configs[pType]
	return config, ok
}

// ListConfigs returns all provider configurations (with masked keys).
func (s *ProviderStore) ListConfigs() []ProviderConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	configs := make([]ProviderConfig, 0, len(s.configs))
	for _, c := range s.configs {
		// Mask key for safety when listing
		masked := c
		if len(masked.APIKey) > 4 {
			masked.APIKey = "..." + masked.APIKey[len(masked.APIKey)-4:]
		} else {
			masked.APIKey = "****"
		}
		configs = append(configs, masked)
	}
	return configs
}

func (s *ProviderStore) load() error {
	data, err := os.ReadFile(s.filePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read provider config file: %w", err)
	}

	return json.Unmarshal(data, &s.configs)
}

func (s *ProviderStore) save() error {
	data, err := json.MarshalIndent(s.configs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal provider configs: %w", err)
	}

	return os.WriteFile(s.filePath, data, 0600) // Secure permissions
}
