// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

import (
	"context"
	"encoding/json"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Message defines the interface that all messages exchanged on the event bus must
// implement. It provides a standard way to manage correlation IDs for tracking
// requests and their corresponding responses.
//
// Summary: Interface for all bus-exchanged messages.
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
// includes a correlation ID field (`CID`) and can be embedded in other message
// structs to provide a common mechanism for message tracking.
//
// Summary: Base struct providing common message functionality.
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
//   - arg: The parameter.
//
// Returns:
//   - result: The result.
//
// Errors:
//   - err: The error if any.
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
//   - arg: The parameter.
//
// Returns:
//   - result: The result.
//
// Errors:
//   - err: The error if any.
//
// Side Effects:
//   - None.
func (m *BaseMessage) SetCorrelationID(id string) {
	m.CID = id
}

// ServiceRegistrationRequest is a message sent to the bus to request the
// registration of a new upstream service. It contains the service's
// configuration and the context for the request.
//
// Summary: Request message for upstream service registration.
type ServiceRegistrationRequest struct {
	BaseMessage
	Context context.Context
	Config  *configv1.UpstreamServiceConfig
}

// ServiceRegistrationResult is a message published in response to a
// ServiceRegistrationRequest. It contains the outcome of the registration
// process, including the generated service key, a list of any tools that were
// discovered, or an error if the registration failed.
//
// Summary: Result message for upstream service registration.
type ServiceRegistrationResult struct {
	BaseMessage
	ServiceKey          string
	DiscoveredTools     []*configv1.ToolDefinition
	DiscoveredResources []*configv1.ResourceDefinition
	Error               error
}

// ToolExecutionRequest is a message sent to the bus to request the execution of
// a specific tool on an upstream service. It includes the name of the tool and
// its inputs in raw JSON format.
//
// Summary: Request message for tool execution.
type ToolExecutionRequest struct {
	BaseMessage
	Context    context.Context
	ToolName   string
	ToolInputs json.RawMessage
}

// ToolExecutionResult is a message published in response to a
// ToolExecutionRequest. It contains the result of the tool execution, in raw
// JSON format, or an error if the execution failed.
//
// Summary: Result message for tool execution.
type ToolExecutionResult struct {
	BaseMessage
	Result json.RawMessage
	Error  error
}

// ServiceListRequest is a message sent to the bus to request a list of all
// registered services.
//
// Summary: Request message for listing registered services.
type ServiceListRequest struct {
	BaseMessage
}

// ServiceListResult is a message published in response to a
// ServiceListRequest. It contains a list of all registered services.
//
// Summary: Result message for listing registered services.
type ServiceListResult struct {
	BaseMessage
	Services []*configv1.UpstreamServiceConfig
	Error    error
}

// ServiceGetRequest is a message sent to the bus to request a specific service.
//
// Summary: Request message for retrieving a single service.
type ServiceGetRequest struct {
	BaseMessage
	ServiceName string
}

// ServiceGetResult is a message published in response to a ServiceGetRequest.
//
// Summary: Result message for retrieving a single service.
type ServiceGetResult struct {
	BaseMessage
	Service *configv1.UpstreamServiceConfig
	Error   error
}
