// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

import (
	"context"
	"encoding/json"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Message - Auto-generated documentation.
//
// Summary: Message defines the interface that all messages exchanged on the event bus must
//
// Methods:
//   - Various methods for Message.
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

// BaseMessage - Auto-generated documentation.
//
// Summary: BaseMessage provides a default implementation of the Message interface. It
//
// Fields:
//   - Various fields for BaseMessage.
type BaseMessage struct {
	CID string `json:"cid"`
}

// CorrelationID returns the correlation ID of the message. This ID is used to associate requests with their corresponding responses in asynchronous workflows.
//
// Parameters:
//   - None
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *BaseMessage) CorrelationID() string {
	return m.CID
}

// SetCorrelationID sets the correlation ID for the message. This is typically called by the message publisher to assign a unique ID to a request.
//
// Parameters:
//   - id (string): The id parameter.
//
// Returns:
//   - None
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (m *BaseMessage) SetCorrelationID(id string) {
	m.CID = id
}

// ServiceRegistrationRequest - Auto-generated documentation.
//
// Summary: ServiceRegistrationRequest is a message sent to the bus to request the
//
// Fields:
//   - Various fields for ServiceRegistrationRequest.
type ServiceRegistrationRequest struct {
	BaseMessage
	Context context.Context
	Config  *configv1.UpstreamServiceConfig
}

// ServiceRegistrationResult - Auto-generated documentation.
//
// Summary: ServiceRegistrationResult is a message published in response to a
//
// Fields:
//   - Various fields for ServiceRegistrationResult.
type ServiceRegistrationResult struct {
	BaseMessage
	ServiceKey          string
	DiscoveredTools     []*configv1.ToolDefinition
	DiscoveredResources []*configv1.ResourceDefinition
	Error               error
}

// ToolExecutionRequest - Auto-generated documentation.
//
// Summary: ToolExecutionRequest is a message sent to the bus to request the execution of
//
// Fields:
//   - Various fields for ToolExecutionRequest.
type ToolExecutionRequest struct {
	BaseMessage
	Context    context.Context
	ToolName   string
	ToolInputs json.RawMessage
}

// ToolExecutionResult - Auto-generated documentation.
//
// Summary: ToolExecutionResult is a message published in response to a
//
// Fields:
//   - Various fields for ToolExecutionResult.
type ToolExecutionResult struct {
	BaseMessage
	Result json.RawMessage
	Error  error
}

// ServiceListRequest - Auto-generated documentation.
//
// Summary: ServiceListRequest is a message sent to the bus to request a list of all
//
// Fields:
//   - Various fields for ServiceListRequest.
type ServiceListRequest struct {
	BaseMessage
}

// ServiceListResult - Auto-generated documentation.
//
// Summary: ServiceListResult is a message published in response to a
//
// Fields:
//   - Various fields for ServiceListResult.
type ServiceListResult struct {
	BaseMessage
	Services []*configv1.UpstreamServiceConfig
	Error    error
}

// ServiceGetRequest - Auto-generated documentation.
//
// Summary: ServiceGetRequest is a message sent to the bus to request a specific service.
//
// Fields:
//   - Various fields for ServiceGetRequest.
type ServiceGetRequest struct {
	BaseMessage
	ServiceName string
}

// ServiceGetResult - Auto-generated documentation.
//
// Summary: ServiceGetResult is a message published in response to a ServiceGetRequest.
//
// Fields:
//   - Various fields for ServiceGetResult.
type ServiceGetResult struct {
	BaseMessage
	Service *configv1.UpstreamServiceConfig
	Error   error
}
