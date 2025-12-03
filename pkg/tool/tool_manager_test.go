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

package tool_test

import (
	"testing"

	"github.com/mcpany/core/pkg/tool"
	"github.com/stretchr/testify/assert"
)

func TestNewToolManager(t *testing.T) {
	tm := tool.NewToolManager(nil)
	assert.NotNil(t, tm)
	assert.Empty(t, tm.ListTools())
}

func TestToolManager_AddAndGetServiceInfo(t *testing.T) {
	tm := tool.NewToolManager(nil)
	serviceInfo := &tool.ServiceInfo{
		Name: "test-service",
	}
	tm.AddServiceInfo("test-service", serviceInfo)
	retrievedServiceInfo, ok := tm.GetServiceInfo("test-service")
	assert.True(t, ok)
	assert.Equal(t, serviceInfo, retrievedServiceInfo)
}
