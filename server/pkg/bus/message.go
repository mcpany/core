// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

import (
	"context"
	"encoding/json"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Message represents the public Message entity.
//
// Summary: Defines the structured data model representing a .
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

// BaseMessage represents the public BaseMessage entity.
//
// Summary: Defines the structured data model representing a message.
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

// CorrelationID serves as a public interface for interacting with CorrelationID.
//
// Summary: Correlation the id appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *BaseMessage) CorrelationID() string {
	return m.CID
}

// SetCorrelationID serves as a public interface for interacting with SetCorrelationID.
//
// Summary: Set the correlation id appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *BaseMessage) SetCorrelationID(id string) {
	m.CID = id
}

// ServiceRegistrationRequest represents the public ServiceRegistrationRequest entity.
//
// Summary: Defines the structured data model representing a registration request.
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

// ServiceRegistrationResult represents the public ServiceRegistrationResult entity.
//
// Summary: Defines the structured data model representing a registration result.
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

// ToolExecutionRequest represents the public ToolExecutionRequest entity.
//
// Summary: Defines the structured data model representing a execution request.
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

// ToolExecutionResult represents the public ToolExecutionResult entity.
//
// Summary: Defines the structured data model representing a execution result.
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

// ServiceListRequest represents the public ServiceListRequest entity.
//
// Summary: Defines the structured data model representing a list request.
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

// ServiceListResult represents the public ServiceListResult entity.
//
// Summary: Defines the structured data model representing a list result.
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

// ServiceGetRequest represents the public ServiceGetRequest entity.
//
// Summary: Defines the structured data model representing a get request.
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

// ServiceGetResult represents the public ServiceGetResult entity.
//
// Summary: Defines the structured data model representing a get result.
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
