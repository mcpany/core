// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

const (
	// ServiceRegistrationRequestTopic defines the NATS subject for publishing service registration requests.
	// Summary: Bus topic for service registration requests.
	ServiceRegistrationRequestTopic = "service_registration_requests"
	// ServiceRegistrationResultTopic defines the NATS subject for receiving service registration outcomes.
	// Summary: Bus topic for service registration results.
	ServiceRegistrationResultTopic = "service_registration_results"
	// ServiceListRequestTopic defines the NATS subject for requesting a list of registered services.
	// Summary: Bus topic for service listing requests.
	ServiceListRequestTopic = "service_list_requests"
	// ServiceListResultTopic defines the NATS subject for receiving the list of services.
	// Summary: Bus topic for service listing results.
	ServiceListResultTopic = "service_list_results"
	// ServiceGetRequestTopic defines the NATS subject for requesting details of a specific service.
	// Summary: Bus topic for single service retrieval requests.
	ServiceGetRequestTopic = "service_get_requests"
	// ServiceGetResultTopic defines the NATS subject for receiving service details.
	// Summary: Bus topic for single service retrieval results.
	ServiceGetResultTopic = "service_get_results"
	// ToolExecutionRequestTopic defines the NATS subject for submitting tool execution requests.
	// Summary: Bus topic for tool execution requests.
	ToolExecutionRequestTopic = "tool_execution_requests"
	// ToolExecutionResultTopic defines the NATS subject for receiving tool execution results.
	// Summary: Bus topic for tool execution results.
	ToolExecutionResultTopic = "tool_execution_results"
)
