// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

const (
// Summary: ServiceRegistrationRequestTopic defines the NATS subject for publishing service registration requests. Defines ServiceRegistrationRequestTopic.
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
	ServiceRegistrationRequestTopic = "service_registration_requests"
// Summary: ServiceRegistrationResultTopic defines the NATS subject for receiving service registration outcomes. Defines ServiceRegistrationResultTopic.
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
	ServiceRegistrationResultTopic = "service_registration_results"
// Summary: ServiceListRequestTopic defines the NATS subject for requesting a list of registered services. Defines ServiceListRequestTopic.
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
	ServiceListRequestTopic = "service_list_requests"
// Summary: ServiceListResultTopic defines the NATS subject for receiving the list of services. Defines ServiceListResultTopic.
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
	ServiceListResultTopic = "service_list_results"
// Summary: ServiceGetRequestTopic defines the NATS subject for requesting details of a specific service. Defines ServiceGetRequestTopic.
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
	ServiceGetRequestTopic = "service_get_requests"
// Summary: ServiceGetResultTopic defines the NATS subject for receiving service details. Defines ServiceGetResultTopic.
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
	ServiceGetResultTopic = "service_get_results"
// Summary: ToolExecutionRequestTopic defines the NATS subject for submitting tool execution requests. Defines ToolExecutionRequestTopic.
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
	ToolExecutionRequestTopic = "tool_execution_requests"
// Summary: ToolExecutionResultTopic defines the NATS subject for receiving tool execution results. Defines ToolExecutionResultTopic.
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
	ToolExecutionResultTopic = "tool_execution_results"
)
