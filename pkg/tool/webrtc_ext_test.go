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

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebrtcTool_GetCacheConfig(t *testing.T) {
	toolDef := &v1.Tool{}
	cacheConfig := &configv1.CacheConfig{}
	callDef := &configv1.WebrtcCallDefinition{}
	callDef.SetCache(cacheConfig)
	wt, err := NewWebrtcTool(toolDef, nil, "service-key", nil, callDef)
	require.NoError(t, err)
	assert.Equal(t, cacheConfig, wt.GetCacheConfig())
}

func TestWebrtcTool_Execute_InvalidInputTemplate(t *testing.T) {
	toolDef := &v1.Tool{}
	callDef := &configv1.WebrtcCallDefinition{}
	inputTransformer := &configv1.InputTransformer{}
	inputTransformer.SetTemplate("{{ .invalid }}")
	callDef.SetInputTransformer(inputTransformer)
	wt, err := NewWebrtcTool(toolDef, nil, "", nil, callDef)
	require.NoError(t, err)

	req := &ExecutionRequest{
		ToolInputs: []byte(`{"message":"hello"}`),
	}

	_, err = wt.Execute(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to render input template")
}

func TestWebrtcTool_CloseMethod(t *testing.T) {
	wt, err := NewWebrtcTool(&v1.Tool{}, nil, "", nil, &configv1.WebrtcCallDefinition{})
	require.NoError(t, err)
	assert.NoError(t, wt.Close())
}
