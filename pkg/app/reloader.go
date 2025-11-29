// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"time"

	"github.com/spf13/afero"
	"google.golang.org/protobuf/proto"

	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/config"
	"github.com/mcpany/core/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

type ConfigReloader struct {
	fs            afero.Fs
	configPaths   []string
	interval      time.Duration
	bus           *bus.BusProvider
	currentConfig *configv1.McpAnyServerConfig
}

func NewConfigReloader(
	fs afero.Fs,
	configPaths []string,
	interval time.Duration,
	bus *bus.BusProvider,
	initialConfig *configv1.McpAnyServerConfig,
) *ConfigReloader {
	return &ConfigReloader{
		fs:            fs,
		configPaths:   configPaths,
		interval:      interval,
		bus:           bus,
		currentConfig: initialConfig,
	}
}

func (r *ConfigReloader) Start(ctx context.Context) {
	if r.interval <= 0 {
		return
	}
	log := logging.GetLogger().With("component", "ConfigReloader")
	log.Info("Starting config reloader", "interval", r.interval)

	go func() {
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.checkAndReload(ctx)
			}
		}
	}()
}

func (r *ConfigReloader) checkAndReload(ctx context.Context) {
	log := logging.GetLogger().With("component", "ConfigReloader")
	store := config.NewFileStore(r.fs, r.configPaths)
	newConfig, err := config.LoadServices(store, "server")
	if err != nil {
		log.Error("Failed to load config during reload check", "error", err)
		return
	}

	if proto.Equal(r.currentConfig, newConfig) {
		return
	}

	log.Info("Configuration change detected, reloading...")

	oldServices := make(map[string]*configv1.UpstreamServiceConfig)
	if r.currentConfig != nil {
		for _, s := range r.currentConfig.GetUpstreamServices() {
			oldServices[s.GetName()] = s
		}
	}

	newServices := make(map[string]*configv1.UpstreamServiceConfig)
	if newConfig != nil {
		for _, s := range newConfig.GetUpstreamServices() {
			newServices[s.GetName()] = s
		}
	}

	registrationBus := bus.GetBus[*bus.ServiceRegistrationRequest](
		r.bus,
		bus.ServiceRegistrationRequestTopic,
	)

	// Check for Added or Modified services
	for name, s := range newServices {
		oldS, exists := oldServices[name]
		if !exists || !proto.Equal(oldS, s) {
			if !exists {
				log.Info("Detected new service", "service", name)
			} else {
				log.Info("Detected modified service", "service", name)
			}
			regReq := &bus.ServiceRegistrationRequest{Config: s}
			if err := registrationBus.Publish(ctx, "request", regReq); err != nil {
				log.Error("Failed to publish registration request", "error", err)
			}
		}
	}

	// Check for Removed services
	for name, s := range oldServices {
		if _, exists := newServices[name]; !exists {
			log.Info("Detected removed service", "service", name)
			sCopy := proto.Clone(s).(*configv1.UpstreamServiceConfig)
			sCopy.SetDisable(true)
			regReq := &bus.ServiceRegistrationRequest{Config: sCopy}
			registrationBus.Publish(ctx, "request", regReq)
		}
	}

	r.currentConfig = newConfig
	log.Info("Configuration reload complete.")
}
