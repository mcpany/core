// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

import (
	"context"
	"encoding/json"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Summary: Defines the interface that all messages exchanged on the event bus must.
type Message interface {
	// CorrelationID returns the unique identifier used to correlate messages.
	//
	// Returns the result.
	CorrelationID() string
	// SetCorrelationID sets the correlation identifier for the message.
	//
	// id is the unique identifier.
	SetCorrelationID(id string)
}

// Summary: Provides a default implementation of the Message interface. It.
type BaseMessage struct {
	CID string `json:"cid"`
}

// Summary: Returns the correlation ID of the message. This ID is used to.
func (m *BaseMessage) CorrelationID() string {
	return m.CID
}

// Summary: Sets the correlation ID for the message. This is typically.
func (m *BaseMessage) SetCorrelationID(id string) {
	m.CID = id
}

// Summary: Is a message sent to the bus to request the.
type ServiceRegistrationRequest struct {
	BaseMessage
	Context context.Context
	Config  *configv1.UpstreamServiceConfig
}

// ServiceRegistrationResult is a message published in response to a
// ServiceRegistrationRequest. It contains the outcome of the registration
// process, including the generated service key, a list of any tools that were
// discovered, or an error if the registration failed.
type ServiceRegistrationResult struct {
	BaseMessage
	ServiceKey          string
	DiscoveredTools     []*configv1.ToolDefinition
	DiscoveredResources []*configv1.ResourceDefinition
	Error               error
}

// Summary: Is a message sent to the bus to request the execution of.
type ToolExecutionRequest struct {
	BaseMessage
	Context    context.Context
	ToolName   string
	ToolInputs json.RawMessage
}

// Summary: Is a message published in response to a.
type ToolExecutionResult struct {
	BaseMessage
	Result json.RawMessage
	Error  error
}

// Summary: Is a message sent to the bus to request a list of all.
type ServiceListRequest struct {
	BaseMessage
}

// Summary: Is a message published in response to a.
type ServiceListResult struct {
	BaseMessage
	Services []*configv1.UpstreamServiceConfig
	Error    error
}

// ServiceGetRequest is a message sent to the bus to request a specific service.
//
// Summary: Is a message sent to the bus to request a specific service.
type ServiceGetRequest struct {
	BaseMessage
	ServiceName string
}

// ServiceGetResult is a message published in response to a ServiceGetRequest.
//
// Summary: Is a message published in response to a ServiceGetRequest.
type ServiceGetResult struct {
	BaseMessage
	Service *configv1.UpstreamServiceConfig
	Error   error
}
