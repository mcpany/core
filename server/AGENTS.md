# MCP Any - Agent Instructions

## Project Goal
The primary goal of this project is to **eliminate the requirement to implement new MCP servers for doing API calls**.

Instead of writing code to create a new MCP server, users should be able to:
1.  **Configure everything**: Load lightweight YAML/JSON configurations to capability-enable different APIs (REST, gRPC, GraphQL, Command-line).
2.  **Run anywhere**: Run a single `mcpany` server instance that acts as a **Universal Adapter**.
3.  **Share configurations**: Share these configurations publicly, allowing others to use the same capabilities without managing binaries or dependencies.

## Key Principles for Documentation and Development

-   **Configuration over Code**: Emphasize that adding new capabilities is done through config files, not by writing Go/Python code.
-   **Single Binary**: Users install one `mcpany` binary (or Docker image) and load N config files. Avoid "binary fatigue".
-   **Universal Adapter**: We support gRPC, OpenAPI, HTTP, GraphQL, and even ensuring CLI commands can be turned into MCP tools.
-   **Security**: The server runs locally. We prioritize security features like:
    -   **Strict Egress Policy**: Control where the server can connect.
    -   **Context Propagation**: Securely pass authentication headers.
    -   **Input Validation**: Prevent command injection and ensuring safe execution.

## Key Features to Highlight

-   **Dynamic Tool Registration & Auto-Discovery**: We can discover tools from Proto files, OpenAPI specs, or Reflection.
-   **Safety Policies**: We allow users to block dangerous operations (e.g. `DELETE`) or restrict access to specific URLs.
-   **Upstream Authentication**: We handle API keys, Bearer tokens, and mTLS so the AI doesn't have to see them.
-   **Multi-User & Multi-Profile**: We support complex multi-tenancy use cases.

## Documentation Standards

We adhere to a **Gold Standard** for documentation. All public API elements (functions, methods, types, constants) must have comprehensive docstrings.

### Structure

Every docstring must follow this structure:

1.  **Summary**: A concise, one-line action statement (e.g., "Calculates the user's tax liability").
2.  **Parameters**: Name, Type, and a meaningful description of what it represents.
3.  **Returns**: Type and description of the output.
4.  **Errors**: Explicitly list exceptions or error states this function might trigger.
5.  **Side Effects**: Note if it modifies global state, writes to DB, or makes network calls.

### Example

```go
// CalculateTax computes the tax amount for a given transaction.
//
// Parameters:
//   - amount (float64): The transaction amount.
//   - rate (float64): The tax rate as a percentage.
//
// Returns:
//   - float64: The calculated tax amount.
//   - error: An error if the rate is negative.
//
// Errors:
//   - Returns "invalid rate" if the rate is less than 0.
//
// Side Effects:
//   - None.
func CalculateTax(amount, rate float64) (float64, error) {
    // ...
}
```

### Constraints

*   **No Empty Calories**: Avoid comments like `// Sets the ID`. Instead, use `// Sets the unique request ID used for distributed tracing`.
*   **Completeness**: Do not skip "simple" functions. Consistency is key.
