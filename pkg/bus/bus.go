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

package bus

import (
	"fmt"
	"sync"

	"github.com/mcpany/core/pkg/bus/memory"
	"github.com/mcpany/core/pkg/bus/redis"
	"github.com/mcpany/core/pkg/busiface"
	"github.com/mcpany/core/proto/bus"
	redisclient "github.com/redis/go-redis/v9"
)

// BusProvider is a thread-safe container for managing multiple, type-safe bus
// instances, with each bus being dedicated to a specific topic. It ensures that
// for any given topic, there is only one bus instance, creating one on demand
// if it doesn't already exist.
//
// This allows different parts of the application to get a bus for a specific
// message type and topic without needing to manage the lifecycle of the bus
// instances themselves.
type BusProvider struct {
	buses       map[string]any
	mu          sync.RWMutex
	config      *bus.MessageBus
	redisClient *redisclient.Client
}

// NewBusProvider creates and returns a new BusProvider, which is used to manage
// multiple topic-based bus instances.
func NewBusProvider(config *bus.MessageBus) (*BusProvider, error) {
	provider := &BusProvider{
		buses:  make(map[string]any),
		config: config,
	}

	switch config.WhichBusType() {
	case bus.MessageBus_InMemory_case:
	case bus.MessageBus_Redis_case:
		redisConfig := config.GetRedis()
		client := redisclient.NewClient(&redisclient.Options{
			Addr:     redisConfig.GetAddress(),
			Password: redisConfig.GetPassword(),
			DB:       int(redisConfig.GetDb()),
		})
		provider.redisClient = client
	default:
		return nil, fmt.Errorf("unknown bus type")
	}

	return provider, nil
}

// GetBus retrieves or creates a bus for a specific topic and message type. If a
// bus for the given topic already exists, it is returned; otherwise, a new one
// is created and stored for future use.
//
// The type parameter T specifies the message type for the bus, ensuring
// type safety for each topic.
//
// Parameters:
//   - p: The BusProvider instance.
//   - topic: The name of the topic for which to get the bus.
//
// Returns a Bus instance for the specified message type and topic.
func GetBus[T any](p *BusProvider, topic string) busiface.Bus[T] {
	p.mu.Lock()
	defer p.mu.Unlock()

	if bus, ok := p.buses[topic]; ok {
		return bus.(busiface.Bus[T])
	}

	var newBus busiface.Bus[T]
	switch p.config.WhichBusType() {
	case bus.MessageBus_InMemory_case:
		newBus = memory.New[T]()
	case bus.MessageBus_Redis_case:
		newBus = redis.New[T](p.redisClient)
	}

	p.buses[topic] = newBus
	return newBus
}
