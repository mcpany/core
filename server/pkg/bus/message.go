// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

import (
// Message defines the interface that all messages exchanged on the event bus must
// implement. It provides a standard way to manage correlation IDs for tracking
// requests and their corresponding responses.
//
// Summary: Represents a Message.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type Message interface {
	// CorrelationID returns the unique identifier used to correlate messages.
	//
	// Returns the result.
	CorrelationID() string
// BaseMessage provides a default implementation of the Message interface. It
// includes a correlation ID field (`CID`) and can be embedded in other message
// structs to provide a common mechanism for message tracking.
//
// Summary: Represents a BaseMessage.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
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
// ServiceRegistrationRequest is a message sent to the bus to request the
// registration of a new upstream service. It contains the service's
// configuration and the context for the request.
//
// ServiceRegistrationResult is a message published in response to a
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// ServiceRegistrationRequest. It contains the outcome of the registration
// process, including the generated service key, a list of any tools that were
// discovered, or an error if the registration failed.
//
// Summary: Represents a ServiceRegistrationResult.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type ServiceRegistrationResult struct {
	BaseMessage
// ToolExecutionRequest is a message sent to the bus to request the execution of
// a specific tool on an upstream service. It includes the name of the tool and
// its inputs in raw JSON format.
//
// Summary: Represents a ToolExecutionRequest.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type ToolExecutionRequest struct {
// ToolExecutionResult is a message published in response to a
// ToolExecutionRequest. It contains the result of the tool execution, in raw
// JSON format, or an error if the execution failed.
//
// Summary: Represents a ToolExecutionResult.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type ToolExecutionResult struct {
// ServiceListRequest is a message sent to the bus to request a list of all
// registered services.
//
// ServiceListResult is a message published in response to a
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// ServiceListRequest. It contains a list of all registered services.
//
// Summary: Represents a ServiceListResult.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type ServiceListResult struct {
	BaseMessage
// ServiceGetRequest is a message sent to the bus to request a specific service.
//
// Summary: Represents a ServiceGetRequest.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type ServiceGetRequest struct {
// ServiceGetResult is a message published in response to a ServiceGetRequest.
//
// Summary: Represents a ServiceGetResult.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type ServiceGetResult struct {
	BaseMessage
	Service *configv1.UpstreamServiceConfig
	Error   error
}
