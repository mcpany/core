// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

const (
	// ServiceRegistrationRequestTopic defines the NATS subject for publishing service registration requests.
//
// Summary: Defines the NATS subject for publishing service registration requests.
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
	// ServiceRegistrationResultTopic defines the NATS subject for receiving service registration outcomes.
//
// Summary: Defines the NATS subject for receiving service registration outcomes.
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
	// ServiceListRequestTopic defines the NATS subject for requesting a list of registered services.
//
// Summary: Defines the NATS subject for requesting a list of registered services.
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
	// ServiceListResultTopic defines the NATS subject for receiving the list of services.
//
// Summary: Defines the NATS subject for receiving the list of services.
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
	// ServiceGetRequestTopic defines the NATS subject for requesting details of a specific service.
//
// Summary: Defines the NATS subject for requesting details of a specific service.
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
	// ServiceGetResultTopic defines the NATS subject for receiving service details.
//
// Summary: Defines the NATS subject for receiving service details.
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
	// ToolExecutionRequestTopic defines the NATS subject for submitting tool execution requests.
//
// Summary: Defines the NATS subject for submitting tool execution requests.
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
	// ToolExecutionResultTopic defines the NATS subject for receiving tool execution results.
//
// Summary: Defines the NATS subject for receiving tool execution results.
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
