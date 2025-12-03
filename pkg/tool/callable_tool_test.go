/*
 * Copyright 2024 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may
 * obtain a copy of the License at
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
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCallable is a mock implementation of the Callable interface.
type MockCallable struct {
	mock.Mock
}

func (m *MockCallable) Call(ctx context.Context, req *ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func TestCallableTool_Execute(t *testing.T) {
	mockCallable := new(MockCallable)
	toolDef := &configv1.ToolDefinition{}
	serviceConfig := &configv1.UpstreamServiceConfig{}

	tool, err := NewCallableTool(toolDef, serviceConfig, mockCallable)
	assert.NoError(t, err)

	req := &ExecutionRequest{}
	expectedResponse := "mocked response"
	mockCallable.On("Call", context.Background(), req).Return(expectedResponse, nil)

	resp, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, resp)
	mockCallable.AssertExpectations(t)
}

func TestCallableTool_Execute_Error(t *testing.T) {
	mockCallable := new(MockCallable)
	toolDef := &configv1.ToolDefinition{}
	serviceConfig := &configv1.UpstreamServiceConfig{}

	tool, err := NewCallableTool(toolDef, serviceConfig, mockCallable)
	assert.NoError(t, err)

	req := &ExecutionRequest{}
	expectedError := errors.New("mocked error")
	mockCallable.On("Call", context.Background(), req).Return(nil, expectedError)

	_, err = tool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockCallable.AssertExpectations(t)
}

func TestCallableTool_Callable(t *testing.T) {
	mockCallable := new(MockCallable)
	toolDef := &configv1.ToolDefinition{}
	serviceConfig := &configv1.UpstreamServiceConfig{}

	tool, err := NewCallableTool(toolDef, serviceConfig, mockCallable)
	assert.NoError(t, err)

	assert.Equal(t, mockCallable, tool.Callable())
}
