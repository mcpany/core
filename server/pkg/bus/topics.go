// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package bus

const (
	// ServiceRegistrationRequestTopic defines the NATS subject for publishing service registration requests.
	// Summary: Defines ServiceRegistrationRequestTopic.
//
// Summary: Represents ->:ServiceRegistrationRequestTopic
//
//
// Summary: Represents ->:ServiceRegistrationResultTopic
//
//
// Summary: Represents ->:ServiceListRequestTopic
//
//
// Summary: Represents ->:ServiceListResultTopic
//
//
// Summary: Represents ->:ServiceGetRequestTopic
//
//
// Summary: Represents ->:ServiceGetResultTopic
//
//
// Summary: Represents ->:ToolExecutionRequestTopic
//
//
// Summary: Represents ->:ToolExecutionResultTopic
//
	ServiceRegistrationRequestTopic = "service_registration_requests"
	// ServiceRegistrationResultTopic defines the NATS subject for receiving service registration outcomes.
	// Summary: Defines ServiceRegistrationResultTopic.
//
// Summary: Represents ->:ServiceRegistrationResultTopic
//
	ServiceRegistrationResultTopic = "service_registration_results"
	// ServiceListRequestTopic defines the NATS subject for requesting a list of registered services.
	// Summary: Defines ServiceListRequestTopic.
//
// Summary: Represents ->:ServiceListRequestTopic
//
	ServiceListRequestTopic = "service_list_requests"
	// ServiceListResultTopic defines the NATS subject for receiving the list of services.
	// Summary: Defines ServiceListResultTopic.
//
// Summary: Represents ->:ServiceListResultTopic
//
	ServiceListResultTopic = "service_list_results"
	// ServiceGetRequestTopic defines the NATS subject for requesting details of a specific service.
	// Summary: Defines ServiceGetRequestTopic.
//
// Summary: Represents ->:ServiceGetRequestTopic
//
	ServiceGetRequestTopic = "service_get_requests"
	// ServiceGetResultTopic defines the NATS subject for receiving service details.
	// Summary: Defines ServiceGetResultTopic.
//
// Summary: Represents ->:ServiceGetResultTopic
//
	ServiceGetResultTopic = "service_get_results"
	// ToolExecutionRequestTopic defines the NATS subject for submitting tool execution requests.
	// Summary: Defines ToolExecutionRequestTopic.
//
// Summary: Represents ->:ToolExecutionRequestTopic
//
	ToolExecutionRequestTopic = "tool_execution_requests"
	// ToolExecutionResultTopic defines the NATS subject for receiving tool execution results.
	// Summary: Defines ToolExecutionResultTopic.
//
// Summary: Represents ->:ToolExecutionResultTopic
//
	ToolExecutionResultTopic = "tool_execution_results"
)
