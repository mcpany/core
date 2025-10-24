/*
 * Copyright 2025 Author(s) of MCP-XY
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

package bus

import (
	"context"
	"encoding/json"

	configv1 "github.com/mcpxy/core/proto/config/v1"
)

// Message defines the interface that all messages exchanged on the event bus must
// implement. It provides a standard way to manage correlation IDs for tracking
// requests and their corresponding responses.
type Message interface {
	// CorrelationID returns the unique identifier used to correlate messages.
	CorrelationID() string
	// SetCorrelationID sets the correlation identifier for the message.
	SetCorrelationID(id string)
}

// BaseMessage provides a default implementation of the Message interface,
// containing a correlation ID field.
type BaseMessage struct {
	CID string `json:"cid"`
}

// CorrelationID returns the correlation ID of the message.
func (m *BaseMessage) CorrelationID() string {
	return m.CID
}

// SetCorrelationID sets the correlation ID for the message.
func (m *BaseMessage) SetCorrelationID(id string) {
	m.CID = id
}

// ServiceRegistrationRequest is a message sent to the bus to request the
// registration of a new upstream service.
type ServiceRegistrationRequest struct {
	BaseMessage
	Context context.Context
	Config  *configv1.UpstreamServiceConfig
}

// ServiceRegistrationResult is a message published in response to a
// ServiceRegistrationRequest. It contains the outcome of the registration,
// including the service key and any discovered tools, or an error if the
// registration failed.
type ServiceRegistrationResult struct {
	BaseMessage
	ServiceKey      string
	DiscoveredTools []*configv1.ToolDefinition
	Error           error
}

// ToolExecutionRequest is a message sent to the bus to request the execution of
// a specific tool on an upstream service.
type ToolExecutionRequest struct {
	BaseMessage
	Context    context.Context
	ToolName   string
	ToolInputs json.RawMessage
}

// ToolExecutionResult is a message published in response to a
// ToolExecutionRequest. It contains the result of the tool execution or an
// error if the execution failed.
type ToolExecutionResult struct {
	BaseMessage
	Result json.RawMessage
	Error  error
}
