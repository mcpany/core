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
	"testing"

	"github.com/mcpany/core/proto/bus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRedisBus_New(t *testing.T) {
	redisBus := bus.RedisBus_builder{
		Address:  proto.String("localhost:6379"),
		Password: proto.String("password"),
		Db:       proto.Int32(1),
	}.Build()

	bus := New[string](redisBus)
	assert.NotNil(t, bus)
	assert.NotNil(t, bus.client)
	options := bus.client.Options()
	assert.Equal(t, "localhost:6379", options.Addr)
	assert.Equal(t, "password", options.Password)
	assert.Equal(t, 1, options.DB)
}
