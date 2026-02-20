// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

const (
	// ServiceRegistrationRequestTopic defines the NATS subject for publishing service registration requests.
	//
	// Summary: Topic for service registration requests.
	ServiceRegistrationRequestTopic = "service_registration_requests"

	// ServiceRegistrationResultTopic defines the NATS subject for receiving service registration outcomes.
	//
	// Summary: Topic for service registration results.
	ServiceRegistrationResultTopic = "service_registration_results"

	// ServiceListRequestTopic defines the NATS subject for requesting a list of registered services.
	//
	// Summary: Topic for service list requests.
	ServiceListRequestTopic = "service_list_requests"

	// ServiceListResultTopic defines the NATS subject for receiving the list of services.
	//
	// Summary: Topic for service list results.
	ServiceListResultTopic = "service_list_results"

	// ServiceGetRequestTopic defines the NATS subject for requesting details of a specific service.
	//
	// Summary: Topic for single service retrieval requests.
	ServiceGetRequestTopic = "service_get_requests"

	// ServiceGetResultTopic defines the NATS subject for receiving service details.
	//
	// Summary: Topic for single service retrieval results.
	ServiceGetResultTopic = "service_get_results"

	// ToolExecutionRequestTopic defines the NATS subject for submitting tool execution requests.
	//
	// Summary: Topic for tool execution requests.
	ToolExecutionRequestTopic = "tool_execution_requests"

	// ToolExecutionResultTopic defines the NATS subject for receiving tool execution results.
	//
	// Summary: Topic for tool execution results.
	ToolExecutionResultTopic = "tool_execution_results"
)
