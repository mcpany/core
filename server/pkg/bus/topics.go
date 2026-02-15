package bus

const (
	// ServiceRegistrationRequestTopic defines the NATS subject for publishing service registration requests.
	ServiceRegistrationRequestTopic = "service_registration_requests"
	// ServiceRegistrationResultTopic defines the NATS subject for receiving service registration outcomes.
	ServiceRegistrationResultTopic = "service_registration_results"
	// ServiceListRequestTopic defines the NATS subject for requesting a list of registered services.
	ServiceListRequestTopic = "service_list_requests"
	// ServiceListResultTopic defines the NATS subject for receiving the list of services.
	ServiceListResultTopic = "service_list_results"
	// ServiceGetRequestTopic defines the NATS subject for requesting details of a specific service.
	ServiceGetRequestTopic = "service_get_requests"
	// ServiceGetResultTopic defines the NATS subject for receiving service details.
	ServiceGetResultTopic = "service_get_results"
	// ToolExecutionRequestTopic defines the NATS subject for submitting tool execution requests.
	ToolExecutionRequestTopic = "tool_execution_requests"
	// ToolExecutionResultTopic defines the NATS subject for receiving tool execution results.
	ToolExecutionResultTopic = "tool_execution_results"
)
