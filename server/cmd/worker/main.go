// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package main implements the worker service.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	buspb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/worker"
)

func setup() (*worker.Worker, error) {
	busConfig := buspb.MessageBus_builder{}.Build()
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		redisBus := buspb.RedisBus_builder{}.Build()
		redisBus.SetAddress(redisAddr)
		busConfig.SetRedis(redisBus)
	} else {
		busConfig.SetInMemory(buspb.InMemoryBus_builder{}.Build())
	}
	return setupWithConfig(busConfig)
}

func setupWithConfig(busConfig *buspb.MessageBus) (*worker.Worker, error) {
	// Validate the bus configuration
	globalSettings := configv1.GlobalSettings_builder{}.Build()
	globalSettings.SetMessageBus(busConfig)

	cfgToValidate := configv1.McpAnyServerConfig_builder{}.Build()
	cfgToValidate.SetGlobalSettings(globalSettings)

	if validationErrors := config.Validate(context.Background(), cfgToValidate, config.Worker); len(validationErrors) > 0 {
		for _, e := range validationErrors {
			fmt.Printf("Config validation error: %v\n", e)
		}
		return nil, fmt.Errorf("config validation failed")
	}

	busProvider, err := bus.NewProvider(busConfig)
	if err != nil {
		return nil, err
	}

	workerCfg := &worker.Config{
		MaxWorkers:   10,
		MaxQueueSize: 100,
	}
	return worker.New(busProvider, workerCfg), nil
}

// main is the entry point for the worker service.
// It sets up the message bus and starts the worker to process tasks.
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
