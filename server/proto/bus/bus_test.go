// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestBusProto(t *testing.T) {
	msg := &InMemoryBus{}
	assert.NotNil(t, msg)
	assert.NotNil(t, msg.ProtoReflect())

	b, err := proto.Marshal(msg)
	assert.NoError(t, err)
	assert.NotNil(t, b)
}
