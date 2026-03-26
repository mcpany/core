// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// TestWeatherProto ...
// Summary: TestWeatherProto
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	req := &GetWeatherRequest{}
	assert.NotNil(t, req)
	assert.NotNil(t, req.ProtoReflect())

	b, err := proto.Marshal(req)
	assert.NoError(t, err)
	assert.NotNil(t, b)
}
