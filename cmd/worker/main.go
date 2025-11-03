/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/config"
	"github.com/mcpany/core/pkg/worker"
	buspb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

func setup() (*worker.Worker, error) {
	busConfig := &buspb.MessageBus{}
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		redisBus := &buspb.RedisBus{}
		redisBus.SetAddress(redisAddr)
		busConfig.SetRedis(redisBus)
	} else {
		busConfig.SetInMemory(&buspb.InMemoryBus{})
	}

	// Validate the bus configuration
	globalSettings := &configv1.GlobalSettings{}
	globalSettings.SetMessageBus(busConfig)

	cfgToValidate := &configv1.McpxServerConfig{}
	cfgToValidate.SetGlobalSettings(globalSettings)

	if validationErrors := config.Validate(cfgToValidate, config.Worker); len(validationErrors) > 0 {
		for _, e := range validationErrors {
			fmt.Printf("Config validation error: %v\n", e)
		}
		return nil, fmt.Errorf("config validation failed")
	}

	busProvider, err := bus.NewBusProvider(busConfig)
	if err != nil {
		return nil, err
	}

	workerCfg := &worker.Config{
		MaxWorkers:   10,
		MaxQueueSize: 100,
	}
	return worker.New(busProvider, workerCfg), nil
}

func main() {
	worker, err := setup()
	if err != nil {
		panic(err)
	}
	worker.Start(context.Background())

	// Wait for a signal to exit
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	worker.Stop()
}
