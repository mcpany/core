// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

import (
	"context"
	"encoding/json"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Message defines the interface that all messages exchanged on the event bus must
//
// Summary: Defines the interface that all messages exchanged on the event bus must
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

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

// BaseMessage provides a default implementation of the Message interface. It
//
// Summary: Provides a default implementation of the Message interface. It
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

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
//
// Summary: Executes CorrelationID operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
//   - None.
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
//
// Summary: Updates SetCorrelationID operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
//   - None.
func (m *BaseMessage) SetCorrelationID(id string) {
	m.CID = id
}

// ServiceRegistrationRequest is a message sent to the bus to request the
//
// Summary: Is a message sent to the bus to request the
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

type ServiceRegistrationRequest struct {
	BaseMessage
	Context context.Context
	Config  *configv1.UpstreamServiceConfig
}

// ServiceRegistrationResult is a message published in response to a
//
// Summary: Is a message published in response to a
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

type ServiceRegistrationResult struct {
	BaseMessage
	ServiceKey          string
	DiscoveredTools     []*configv1.ToolDefinition
	DiscoveredResources []*configv1.ResourceDefinition
	Error               error
}

// ToolExecutionRequest is a message sent to the bus to request the execution of
//
// Summary: Is a message sent to the bus to request the execution of
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

type ToolExecutionRequest struct {
	BaseMessage
	Context    context.Context
	ToolName   string
	ToolInputs json.RawMessage
}

// ToolExecutionResult is a message published in response to a
//
// Summary: Is a message published in response to a
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

type ToolExecutionResult struct {
	BaseMessage
	Result json.RawMessage
	Error  error
}

// ServiceListRequest is a message sent to the bus to request a list of all
//
// Summary: Is a message sent to the bus to request a list of all
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

type ServiceListRequest struct {
	BaseMessage
}

// ServiceListResult is a message published in response to a
//
// Summary: Is a message published in response to a
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

type ServiceListResult struct {
	BaseMessage
	Services []*configv1.UpstreamServiceConfig
	Error    error
}

// ServiceGetRequest is a message sent to the bus to request a specific service.
//
// Summary: Is a message sent to the bus to request a specific service.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

type ServiceGetRequest struct {
	BaseMessage
	ServiceName string
}

// ServiceGetResult is a message published in response to a ServiceGetRequest.
//
// Summary: Is a message published in response to a ServiceGetRequest.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

type ServiceGetResult struct {
	BaseMessage
	Service *configv1.UpstreamServiceConfig
	Error   error
}
