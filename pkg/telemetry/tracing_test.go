// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestInitTracer(t *testing.T) {
	t.Run("Disabled", func(t *testing.T) {
		shutdown, err := InitTracer(context.Background(), nil)
		assert.NoError(t, err)
		assert.NotNil(t, shutdown)
		err = shutdown(context.Background())
		assert.NoError(t, err)

		shutdown, err = InitTracer(context.Background(), &configv1.TracingConfig{Enabled: proto.Bool(false)})
		assert.NoError(t, err)
		assert.NotNil(t, shutdown)
		err = shutdown(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Enabled", func(t *testing.T) {
		shutdown, err := InitTracer(context.Background(), &configv1.TracingConfig{
			Enabled:  proto.Bool(true),
			Endpoint: proto.String("localhost:4317"),
			Insecure: proto.Bool(true),
		})
		assert.NoError(t, err)
		assert.NotNil(t, shutdown)
		_ = shutdown(context.Background())
	})
}
