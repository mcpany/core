// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
	return setupWithConfig(busConfig)
}

func setupWithConfig(busConfig *buspb.MessageBus) (*worker.Worker, error) {
	// Validate the bus configuration
	globalSettings := &configv1.GlobalSettings{}
	globalSettings.SetMessageBus(busConfig)

	cfgToValidate := &configv1.McpAnyServerConfig{}
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
