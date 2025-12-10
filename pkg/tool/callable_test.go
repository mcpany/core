/*
 * Copyright 2024 Author(s) of MCP Any
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

package tool_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type mockCallable struct {
	err error
}

func (m *mockCallable) Call(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if m.err != nil {
		return nil, m.err
	}
	return "response", nil
}

func TestCallableTool_Execute(t *testing.T) {
	tests := []struct {
		name          string
		enableLogging bool
		expectLogs    bool
		callable      tool.Callable
	}{
		{
			name:          "logging enabled",
			enableLogging: true,
			expectLogs:    true,
			callable:      &mockCallable{},
		},
		{
			name:          "logging disabled",
			enableLogging: false,
			expectLogs:    false,
			callable:      &mockCallable{},
		},
		{
			name:          "logging enabled with error",
			enableLogging: true,
			expectLogs:    true,
			callable:      &mockCallable{err: assert.AnError},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logging.ForTestsOnlyResetLogger()
			var buf bytes.Buffer
			logging.Init(slog.LevelInfo, &buf)

			toolDef := configv1.ToolDefinition_builder{
				Name: proto.String("test-tool"),
			}.Build()
			serviceConfig := configv1.UpstreamServiceConfig_builder{
				Logging: configv1.LoggingConfig_builder{
					Enable: &tc.enableLogging,
				}.Build(),
			}.Build()
			callableTool, err := tool.NewCallableTool(toolDef, serviceConfig, tc.callable)
			require.NoError(t, err)

			_, err = callableTool.Execute(context.Background(), &tool.ExecutionRequest{})
			if tc.callable.(*mockCallable).err != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tc.expectLogs {
				assert.True(t, strings.Contains(buf.String(), "Executing tool"))
				if tc.callable.(*mockCallable).err != nil {
					assert.True(t, strings.Contains(buf.String(), "Tool execution failed"))
				} else {
					assert.True(t, strings.Contains(buf.String(), "Tool execution successful"))
				}
			} else {
				assert.False(t, strings.Contains(buf.String(), "Executing tool"))
				assert.False(t, strings.Contains(buf.String(), "Tool execution successful"))
			}
		})
	}
}
