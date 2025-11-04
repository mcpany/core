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

package redis

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	bus_pb "github.com/mcpany/core/proto/bus"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRedisBus_Publish(t *testing.T) {
	client, mock := redismock.NewClientMock()
	bus := NewWithClient[string](client)

	msg, _ := json.Marshal("hello")
	mock.ExpectPublish("test", msg).SetVal(1)
	err := bus.Publish(context.Background(), "test", "hello")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisBus_Publish_MarshalError(t *testing.T) {
	client, _ := redismock.NewClientMock()
	bus := NewWithClient[chan int](client)

	err := bus.Publish(context.Background(), "test", make(chan int))
	assert.Error(t, err)
}

func TestRedisBus_Publish_RedisError(t *testing.T) {
	client, mock := redismock.NewClientMock()
	bus := NewWithClient[string](client)

	msg, _ := json.Marshal("hello")
	mock.ExpectPublish("test", msg).SetErr(errors.New("redis error"))
	err := bus.Publish(context.Background(), "test", "hello")
	assert.Error(t, err)
	assert.Equal(t, "redis error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisBus_New(t *testing.T) {
	redisBus := bus_pb.RedisBus_builder{
		Address:  proto.String("localhost:6379"),
		Password: proto.String("password"),
		Db:       proto.Int32(1),
	}.Build()

	bus := New[string](redisBus)
	assert.NotNil(t, bus)
	assert.NotNil(t, bus.client)
}

func TestRedisBus_Subscribe(t *testing.T) {
	redisBus := bus_pb.RedisBus_builder{
		Address: proto.String("localhost:6379"),
	}.Build()

	client := redis.NewClient(&redis.Options{
		Addr: redisBus.GetAddress(),
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}

	bus := New[string](redisBus)

	var wg sync.WaitGroup
	wg.Add(1)

	unsub := bus.Subscribe(context.Background(), "test", func(msg string) {
		assert.Equal(t, "hello", msg)
		wg.Done()
	})
	defer unsub()

	// Use a channel to signal when the subscriber is ready
	ready := make(chan struct{})
	go func() {
		close(ready)
	}()
	<-ready

	err := bus.Publish(context.Background(), "test", "hello")
	assert.NoError(t, err)

	wg.Wait()
}

func TestRedisBus_Subscribe_UnmarshalError(t *testing.T) {
	redisBus := bus_pb.RedisBus_builder{
		Address: proto.String("localhost:6379"),
	}.Build()

	client := redis.NewClient(&redis.Options{
		Addr: redisBus.GetAddress(),
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}

	bus := New[string](redisBus)
	unsub := bus.Subscribe(context.Background(), "test_unmarshal_error", func(msg string) {
		assert.Fail(t, "should not be called")
	})
	defer unsub()

	ready := make(chan struct{})
	go func() {
		close(ready)
	}()
	<-ready

	err := client.Publish(context.Background(), "test_unmarshal_error", "invalid json").Err()
	assert.NoError(t, err)

	// Give a moment for the message to be processed
	time.Sleep(100 * time.Millisecond)
}

func TestRedisBus_SubscribeOnce(t *testing.T) {
	redisBus := bus_pb.RedisBus_builder{
		Address: proto.String("localhost:6379"),
	}.Build()

	client := redis.NewClient(&redis.Options{
		Addr: redisBus.GetAddress(),
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}

	bus := New[string](redisBus)

	var wg sync.WaitGroup
	wg.Add(1)

	unsub := bus.SubscribeOnce(context.Background(), "test_once", func(msg string) {
		assert.Equal(t, "hello", msg)
		wg.Done()
	})
	defer unsub()

	ready := make(chan struct{})
	go func() {
		close(ready)
	}()
	<-ready

	err := bus.Publish(context.Background(), "test_once", "hello")
	assert.NoError(t, err)
	err = bus.Publish(context.Background(), "test_once", "hello")
	assert.NoError(t, err)

	wg.Wait()
}

func TestRedisBus_Unsubscribe(t *testing.T) {
	redisBus := bus_pb.RedisBus_builder{
		Address: proto.String("localhost:6379"),
	}.Build()

	client := redis.NewClient(&redis.Options{
		Addr: redisBus.GetAddress(),
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}

	bus := New[string](redisBus)

	unsub := bus.Subscribe(context.Background(), "test_unsub", func(msg string) {
		assert.Fail(t, "should not be called")
	})

	ready := make(chan struct{})
	go func() {
		close(ready)
	}()
	<-ready

	unsub()

	// Give the unsubscribe a moment to take effect
	time.Sleep(100 * time.Millisecond)

	err := bus.Publish(context.Background(), "test_unsub", "hello")
	assert.NoError(t, err)

	// Give a moment to see if the message is received
	time.Sleep(100 * time.Millisecond)
}

func TestRedisBus_SubscribeOnce_Unsubscribe(t *testing.T) {
	redisBus := bus_pb.RedisBus_builder{
		Address: proto.String("localhost:6379"),
	}.Build()

	client := redis.NewClient(&redis.Options{
		Addr: redisBus.GetAddress(),
	})
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		t.Skip("Redis is not available")
	}

	bus := New[string](redisBus)

	unsub := bus.SubscribeOnce(context.Background(), "test_once_unsub", func(msg string) {
		assert.Fail(t, "should not be called")
	})

	ready := make(chan struct{})
	go func() {
		close(ready)
	}()
	<-ready

	unsub()

	// Give the unsubscribe a moment to take effect
	time.Sleep(100 * time.Millisecond)

	err := bus.Publish(context.Background(), "test_once_unsub", "hello")
	assert.NoError(t, err)

	// Give a moment to see if the message is received
	time.Sleep(100 * time.Millisecond)
}
