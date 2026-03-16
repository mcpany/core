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
// Summary: Message defines the interface that all messages exchanged on the event bus must
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
// Summary: BaseMessage provides a default implementation of the Message interface. It
type BaseMessage struct {
	CID string `json:"cid"`
}

// CorrelationID returns the correlation ID of the message. This ID is used to associate requests with their corresponding responses in asynchronous workflows.
//
// Summary: CorrelationID returns the correlation ID of the message. This ID is used to associate requests with their corresponding responses in asynchronous workflows.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting text.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Side Effects:
//   - May modify internal state or perform external network calls.
func (m *BaseMessage) CorrelationID() string {
	return m.CID
// SetCorrelationID sets the correlation ID for the message. This is typically called by the message publisher to assign a unique ID to a request.
//
// Summary: SetCorrelationID sets the correlation ID for the message. This is typically called by the message publisher to assign a unique ID to a request.
//
// Parameters:
//   - id (string): The unique identifier.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// ServiceRegistrationRequest is a message sent to the bus to request the
//
// Summary: ServiceRegistrationRequest is a message sent to the bus to request the

// ServiceRegistrationRequest is a message sent to the bus to request the
//
// Summary: ServiceRegistrationRequest is a message sent to the bus to request the
type ServiceRegistrationRequest struct {
	BaseMessage
// ServiceRegistrationResult is a message published in response to a
//
// Summary: ServiceRegistrationResult is a message published in response to a
// ServiceRegistrationResult is a message published in response to a
//
// Summary: ServiceRegistrationResult is a message published in response to a
type ServiceRegistrationResult struct {
	BaseMessage
	ServiceKey          string
	DiscoveredTools     []*configv1.ToolDefinition
	DiscoveredResources []*configv1.ResourceDefinition
// ToolExecutionRequest is a message sent to the bus to request the execution of
//
// Summary: ToolExecutionRequest is a message sent to the bus to request the execution of
// ToolExecutionRequest is a message sent to the bus to request the execution of
//
// Summary: ToolExecutionRequest is a message sent to the bus to request the execution of
type ToolExecutionRequest struct {
	BaseMessage
	Context    context.Context
	ToolName   string
// ToolExecutionResult is a message published in response to a
//
// Summary: ToolExecutionResult is a message published in response to a
// ToolExecutionResult is a message published in response to a
//
// Summary: ToolExecutionResult is a message published in response to a
type ToolExecutionResult struct {
	BaseMessage
	Result json.RawMessage
// ServiceListRequest is a message sent to the bus to request a list of all
//
// Summary: ServiceListRequest is a message sent to the bus to request a list of all

// ServiceListRequest is a message sent to the bus to request a list of all
//
// Summary: ServiceListRequest is a message sent to the bus to request a list of all
// ServiceListResult is a message published in response to a
//
// Summary: ServiceListResult is a message published in response to a
}

// ServiceListResult is a message published in response to a
//
// Summary: ServiceListResult is a message published in response to a
type ServiceListResult struct {
// ServiceGetRequest is a message sent to the bus to request a specific service.
//
// Summary: ServiceGetRequest is a message sent to the bus to request a specific service.
	Services []*configv1.UpstreamServiceConfig
	Error    error
}

// ServiceGetRequest is a message sent to the bus to request a specific service.
// ServiceGetResult is a message published in response to a ServiceGetRequest.
//
// Summary: ServiceGetResult is a message published in response to a ServiceGetRequest.
// Summary: ServiceGetRequest is a message sent to the bus to request a specific service.
type ServiceGetRequest struct {
	BaseMessage
	ServiceName string
}

// ServiceGetResult is a message published in response to a ServiceGetRequest.
//
// Summary: ServiceGetResult is a message published in response to a ServiceGetRequest.
type ServiceGetResult struct {
	BaseMessage
	Service *configv1.UpstreamServiceConfig
	Error   error
}
