// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhooks

// Handler defines the interface for handling webhook requests.
// Implementations of this interface process incoming webhook events.
//
// Summary: Interface for webhook handlers.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type Handler interface {
	// Handle processes the webhook request.
	//
	// Parameters:
	//   w: http.ResponseWriter. The HTTP response writer to write the response to.
	//   r: *http.Request. The HTTP request containing the webhook payload.
	//
	// Returns:
	//
	//	None.
	//
	// Side Effects:
// NewRegistry creates and initializes a new Registry instance.
//
// Register registers a handler with a specific name.
// If a handler with the same name already exists, it will be overwritten.
//
// Summary: Registers a webhook handler.
//
// Parameters:
//   - name: string. The name/path to register the handler under.
// Get retrieves a handler by its name.
//
// Summary: Retrieves a webhook handler by name.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Parameters:
//   - name: string. The name of the handler to retrieve.
//
// Returns:
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - Handler: The registered handler, if found.
//   - bool: True if the handler exists, false otherwise.
//
// Side Effects:
//   - None.
// Errors:
//   - triggers relevant error states on failure.
func (r *Registry) Get(name string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
//   - None.
// Errors:
// Parameters:
//
// Summary: Registers a webhook handler.
//
// If a handler with the same name already exists, it will be overwritten.
// Register registers a handler with a specific name.
//
// NewRegistry creates and initializes a new Registry instance.
	// Side Effects:
	//
	//	None.
	//
	// Returns:
	//
	//   r: *http.Request. The HTTP request containing the webhook payload.
	//   w: http.ResponseWriter. The HTTP response writer to write the response to.
	// Parameters:
	//
	// Handle processes the webhook request.
	h, ok := r.hooks[name]
	return h, ok
}
