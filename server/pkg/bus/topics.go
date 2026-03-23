// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

const (
	// ServiceRegistrationRequestTopic defines the NATS subject for publishing service registration requests.
	//
	// Summary: Defines the topic used to broadcast new service registrations across the cluster.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: The topic name
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	ServiceRegistrationRequestTopic = "service_registration_requests"

	// ServiceRegistrationResultTopic defines the NATS subject for receiving service registration outcomes.
	//
	// Summary: Defines the topic used to listen for the result of a registration request.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: The topic name
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	ServiceRegistrationResultTopic = "service_registration_results"

	// ServiceListRequestTopic defines the NATS subject for requesting a list of registered services.
	//
	// Summary: Defines the topic used to request all known services from the registry.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: The topic name
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	ServiceListRequestTopic = "service_list_requests"

	// ServiceListResultTopic defines the NATS subject for receiving the list of services.
	//
	// Summary: Defines the topic used to reply with the current list of known services.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: The topic name
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	ServiceListResultTopic = "service_list_results"

	// ServiceGetRequestTopic defines the NATS subject for requesting details of a specific service.
	//
	// Summary: Defines the topic used to query a specific service's detailed configuration.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: The topic name
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	ServiceGetRequestTopic = "service_get_requests"

	// ServiceGetResultTopic defines the NATS subject for receiving service details.
	//
	// Summary: Defines the topic used to return the detailed configuration of a requested service.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: The topic name
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	ServiceGetResultTopic = "service_get_results"

	// ToolExecutionRequestTopic defines the NATS subject for submitting tool execution requests.
	//
	// Summary: Defines the topic used to dispatch an MCP tool call to the appropriate upstream adapter.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: The topic name
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	ToolExecutionRequestTopic = "tool_execution_requests"

	// ToolExecutionResultTopic defines the NATS subject for receiving tool execution results.
	//
	// Summary: Defines the topic used to receive the final output from an upstream tool execution.
	//
	// Parameters:
	//   - none
	//
	// Returns:
	//   - string: The topic name
	//
	// Errors:
	//   - none
	//
	// Side Effects:
	//   - none
	ToolExecutionResultTopic = "tool_execution_results"
)
