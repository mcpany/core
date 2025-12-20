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
