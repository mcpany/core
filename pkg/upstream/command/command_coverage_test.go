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

package command

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestCommandUpstream_Register_DisabledItems(t *testing.T) {
	u := NewCommandUpstream()
	tm := newMockToolManager()
	pm := prompt.NewPromptManager()
	rm := resource.NewResourceManager()

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-disabled-items")
	cmdService := &configv1.CommandLineUpstreamService{}
	cmdService.SetCommand("/bin/echo")

	// Disabled Tool
	toolDef := configv1.ToolDefinition_builder{
		Name:    proto.String("echo"),
		CallId:  proto.String("echo-call"),
		Disable: proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Id: proto.String("echo-call"),
	}.Build()
	calls := make(map[string]*configv1.CommandLineCallDefinition)
	calls["echo-call"] = callDef
	cmdService.SetCalls(calls)
	cmdService.SetTools([]*configv1.ToolDefinition{toolDef})

	// Disabled Prompt
	promptDef := configv1.PromptDefinition_builder{
		Name:    proto.String("test-prompt"),
		Disable: proto.Bool(true),
	}.Build()
	cmdService.SetPrompts([]*configv1.PromptDefinition{promptDef})

	serviceConfig.SetCommandLineService(cmdService)

	_, discoveredTools, _, err := u.Register(
		context.Background(),
		serviceConfig,
		tm,
		pm,
		rm,
		false,
	)
	require.NoError(t, err)
	assert.Len(t, discoveredTools, 0)
	assert.Len(t, tm.ListTools(), 0)

	// Check prompts
	// PM doesn't have list/len easily, so try to get
	_, ok := pm.GetPrompt("test-disabled-items.test-prompt")
	assert.False(t, ok)
}
