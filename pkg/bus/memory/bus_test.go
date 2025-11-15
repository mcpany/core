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

package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	bus := New[any]()
	assert.NotNil(t, bus)
	assert.NotNil(t, bus.subscribers)
	assert.Equal(t, defaultPublishTimeout, bus.publishTimeout)
}

func TestPublishAndSubscribe(t *testing.T) {
	bus := New[string]()
	var wg sync.WaitGroup
	wg.Add(1)

	var receivedMsg string
	bus.Subscribe(context.Background(), "test-topic", func(msg string) {
		receivedMsg = msg
		wg.Done()
	})

	bus.Publish(context.Background(), "test-topic", "hello")
	wg.Wait()

	assert.Equal(t, "hello", receivedMsg)
}

func TestPublishToDifferentTopic(t *testing.T) {
	bus := New[string]()
	var received bool
	bus.Subscribe(context.Background(), "topic-a", func(msg string) {
		received = true
	})

	bus.Publish(context.Background(), "topic-b", "hello")
	time.Sleep(100 * time.Millisecond) // Allow time for potential message processing

	assert.False(t, received, "should not receive message from a different topic")
}

func TestSubscribeOnce(t *testing.T) {
	bus := New[string]()
	var wg sync.WaitGroup
	wg.Add(1)

	var receivedMsgs []string
	bus.SubscribeOnce(context.Background(), "test-topic", func(msg string) {
		receivedMsgs = append(receivedMsgs, msg)
		wg.Done()
	})

	bus.Publish(context.Background(), "test-topic", "hello")
	bus.Publish(context.Background(), "test-topic", "world")
	wg.Wait()

	assert.Len(t, receivedMsgs, 1)
	assert.Equal(t, "hello", receivedMsgs[0])
}

func TestUnsubscribe(t *testing.T) {
	bus := New[string]()
	var receivedMsgs []string
	unsubscribe := bus.Subscribe(context.Background(), "test-topic", func(msg string) {
		receivedMsgs = append(receivedMsgs, msg)
	})

	bus.Publish(context.Background(), "test-topic", "hello")
	time.Sleep(100 * time.Millisecond) // Allow time for message processing

	unsubscribe()

	bus.Publish(context.Background(), "test-topic", "world")
	time.Sleep(100 * time.Millisecond) // Allow time for message processing

	assert.Len(t, receivedMsgs, 1)
	assert.Equal(t, "hello", receivedMsgs[0])
}

func TestConcurrentPublishAndSubscribe(t *testing.T) {
	bus := New[string]()
	numSubscribers := 10
	numMessages := 100
	var wg sync.WaitGroup
	wg.Add(numSubscribers * numMessages)

	var receivedMsgs = make([][]string, numSubscribers)
	var readyWg sync.WaitGroup
	readyWg.Add(numSubscribers)

	for i := 0; i < numSubscribers; i++ {
		go func(i int) {
			bus.Subscribe(context.Background(), "test-topic", func(msg string) {
				receivedMsgs[i] = append(receivedMsgs[i], msg)
				wg.Done()
			})
			readyWg.Done()
		}(i)
	}

	readyWg.Wait()

	for i := 0; i < numMessages; i++ {
		bus.Publish(context.Background(), "test-topic", "hello")
	}

	wg.Wait()

	for i := 0; i < numSubscribers; i++ {
		assert.Len(t, receivedMsgs[i], numMessages)
	}
}
