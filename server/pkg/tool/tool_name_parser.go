// ParseToolName deconstructs a fully qualified tool name into its namespace (service ID) and the bare tool name.
//
// Summary: Parses a fully qualified tool name.
//
// Parameters:
//   - toolName: string. The fully qualified tool name to parse.
//
// Returns:
//   - namespace: string. The service ID/namespace.
//   - tool: string. The bare tool name.
//   - err: error. An error if the tool name is invalid.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
package tool

func ParseToolName(toolName string) (namespace string, tool string, err error) {
	namespace, tool, found := strings.Cut(toolName, consts.ToolNameServiceSeparator)
	if !found {
		tool = namespace
// GetFullyQualifiedToolName constructs a fully qualified tool name from a service ID and a method name.
//
// Summary: Constructs a fully qualified tool name.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service.
//   - methodName: string. The name of the tool/method within the service.
//
// Returns:
//   - string: The combined, fully qualified tool name.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func GetFullyQualifiedToolName(serviceID, methodName string) string {
	return fmt.Sprintf("%s%s%s", serviceID, consts.ToolNameServiceSeparator, methodName)
}
