// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

import (
	"context"
	"encoding/json"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Message message represents a message.
//
// Summary: Message represents a message.
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

// BaseMessage baseMessage represents a base message.
//
// Summary: BaseMessage represents a base message.
type BaseMessage struct {
	CID string `json:"cid"`
}

// CorrelationID returns the correlation ID of the message. This ID is used to associate requests with their corresponding responses in asynchronous workflows.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - string: The resulting string.
//
// Errors: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes CorrelationID operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (m *BaseMessage) CorrelationID() string {
	return m.CID
}

// SetCorrelationID sets the correlation ID for the message. This is typically called by the message publisher to assign a unique ID to a request.
//
// Parameters: - None.
//   - id (string): The id parameter.
//
// Returns: - None.
//   - None.
//
// Errors: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Updates SetCorrelationID operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (m *BaseMessage) SetCorrelationID(id string) {
	m.CID = id
}

// ServiceRegistrationRequest serviceRegistrationRequest represents a service registration request.
//
// Summary: ServiceRegistrationRequest represents a service registration request.
type ServiceRegistrationRequest struct {
	BaseMessage
	Context context.Context
	Config  *configv1.UpstreamServiceConfig
}

// ServiceRegistrationResult serviceRegistrationResult represents a service registration result.
//
// Summary: ServiceRegistrationResult represents a service registration result.
type ServiceRegistrationResult struct {
	BaseMessage
	ServiceKey          string
	DiscoveredTools     []*configv1.ToolDefinition
	DiscoveredResources []*configv1.ResourceDefinition
	Error               error
}

// ToolExecutionRequest toolExecutionRequest represents a tool execution request.
//
// Summary: ToolExecutionRequest represents a tool execution request.
type ToolExecutionRequest struct {
	BaseMessage
	Context    context.Context
	ToolName   string
	ToolInputs json.RawMessage
}

// ToolExecutionResult toolExecutionResult represents a tool execution result.
//
// Summary: ToolExecutionResult represents a tool execution result.
type ToolExecutionResult struct {
	BaseMessage
	Result json.RawMessage
	Error  error
}

// ServiceListRequest serviceListRequest represents a service list request.
//
// Summary: ServiceListRequest represents a service list request.
type ServiceListRequest struct {
	BaseMessage
}

// ServiceListResult serviceListResult represents a service list result.
//
// Summary: ServiceListResult represents a service list result.
type ServiceListResult struct {
	BaseMessage
	Services []*configv1.UpstreamServiceConfig
	Error    error
}

// ServiceGetRequest serviceGetRequest represents a service get request.
//
// Summary: ServiceGetRequest represents a service get request.
type ServiceGetRequest struct {
	BaseMessage
	ServiceName string
}

// ServiceGetResult serviceGetResult represents a service get result.
//
// Summary: ServiceGetResult represents a service get result.
type ServiceGetResult struct {
	BaseMessage
	Service *configv1.UpstreamServiceConfig
	Error   error
}
