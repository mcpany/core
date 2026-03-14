// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0
// Summary: ServiceRegistrationRequestTopic defines the NATS subject for publishing service registration requests.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: ServiceRegistrationResultTopic defines the NATS subject for receiving service registration outcomes.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: ServiceListRequestTopic defines the NATS subject for requesting a list of registered services.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: ServiceListResultTopic defines the NATS subject for receiving the list of services.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: ServiceGetRequestTopic defines the NATS subject for requesting details of a specific service.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: ServiceGetResultTopic defines the NATS subject for receiving service details.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: ToolExecutionRequestTopic defines the NATS subject for submitting tool execution requests.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// Summary: ToolExecutionResultTopic defines the NATS subject for receiving tool execution results.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
package bus

const (
	ServiceRegistrationRequestTopic	= "service_registration_requests"

	ServiceRegistrationResultTopic	= "service_registration_results"

	ServiceListRequestTopic	= "service_list_requests"

	ServiceListResultTopic	= "service_list_results"

	ServiceGetRequestTopic	= "service_get_requests"

	ServiceGetResultTopic	= "service_get_results"

	ToolExecutionRequestTopic	= "tool_execution_requests"

	ToolExecutionResultTopic	= "tool_execution_results"
)
