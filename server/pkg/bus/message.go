// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

import (
	"context"
	"encoding/json"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Message defines the interface that all messages exchanged on the event bus must.
//
// Summary: defines the interface that all messages exchanged on the event bus must.
type Message interface {
	// CorrelationID returns the unique identifier used to correlate messages.
	//
	// Summary: returns the unique identifier used to correlate messages.
	//
	// Parameters:
	//   None.
	//
	// Returns:
	//   - string: The string.
	CorrelationID() string
	// SetCorrelationID sets the correlation identifier for the message.
	//
	// Summary: sets the correlation identifier for the message.
	//
	// Parameters:
	//   - id: string. The unique identifier.
	//
	// Returns:
	//   None.
	SetCorrelationID(id string)
}

// BaseMessage provides a default implementation of the Message interface. It.
//
// Summary: provides a default implementation of the Message interface. It.
type BaseMessage struct {
	CID string `json:"cid"`
}

// CorrelationID returns the correlation ID of the message. This ID is used to.
//
// Summary: returns the correlation ID of the message. This ID is used to.
//
// Parameters:
//   None.
//
// Returns:
//   - string: The string.
func (m *BaseMessage) CorrelationID() string {
	return m.CID
}

// SetCorrelationID sets the correlation ID for the message. This is typically.
//
// Summary: sets the correlation ID for the message. This is typically.
//
// Parameters:
//   - id: string. The id.
//
// Returns:
//   None.
func (m *BaseMessage) SetCorrelationID(id string) {
	m.CID = id
}

// ServiceRegistrationRequest is a message sent to the bus to request the.
//
// Summary: is a message sent to the bus to request the.
type ServiceRegistrationRequest struct {
	BaseMessage
	Context context.Context
	Config  *configv1.UpstreamServiceConfig
}

// ServiceRegistrationResult is a message published in response to a.
//
// Summary: is a message published in response to a.
type ServiceRegistrationResult struct {
	BaseMessage
	ServiceKey          string
	DiscoveredTools     []*configv1.ToolDefinition
	DiscoveredResources []*configv1.ResourceDefinition
	Error               error
}

// ToolExecutionRequest is a message sent to the bus to request the execution of.
//
// Summary: is a message sent to the bus to request the execution of.
type ToolExecutionRequest struct {
	BaseMessage
	Context    context.Context
	ToolName   string
	ToolInputs json.RawMessage
}

// ToolExecutionResult is a message published in response to a.
//
// Summary: is a message published in response to a.
type ToolExecutionResult struct {
	BaseMessage
	Result json.RawMessage
	Error  error
}

// ServiceListRequest is a message sent to the bus to request a list of all.
//
// Summary: is a message sent to the bus to request a list of all.
type ServiceListRequest struct {
	BaseMessage
}

// ServiceListResult is a message published in response to a.
//
// Summary: is a message published in response to a.
type ServiceListResult struct {
	BaseMessage
	Services []*configv1.UpstreamServiceConfig
	Error    error
}

// ServiceGetRequest is a message sent to the bus to request a specific service.
//
// Summary: is a message sent to the bus to request a specific service.
type ServiceGetRequest struct {
	BaseMessage
	ServiceName string
}

// ServiceGetResult is a message published in response to a ServiceGetRequest.
//
// Summary: is a message published in response to a ServiceGetRequest.
type ServiceGetResult struct {
	BaseMessage
	Service *configv1.UpstreamServiceConfig
	Error   error
}
