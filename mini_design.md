# Mini-Design Doc: SQL Database Provider

## Goal
Enable MCP Any to connect to SQL databases (Postgres, SQLite, MySQL) and expose SQL queries as MCP tools. This allows LLMs to query databases directly using pre-defined safe queries.

## Changes

### 1. Protocol Buffers (`proto/config/v1`)
*   **`upstream_service.proto`**:
    *   Add `SqlUpstreamService` message:
        *   `driver`: Driver name (e.g., "postgres", "sqlite").
        *   `dsn`: Data Source Name (connection string).
        *   `calls`: Map of `SqlCallDefinition`.
    *   Update `UpstreamServiceConfig` oneof to include `SqlUpstreamService`.
*   **`call.proto`**:
    *   Add `SqlCallDefinition` message:
        *   `id`: Unique identifier.
        *   `query`: The SQL query string (e.g., `SELECT * FROM users WHERE id = $1`).
        *   `parameter_order`: List of input property names that map to query parameters in order.
        *   `input_schema`: JSON schema for the tool input.
        *   `output_schema`: JSON schema for the tool output.
        *   `cache`: Cache configuration.

### 2. Implementation (`server/pkg/upstream/sql`)
*   **`Upstream`**: Implements `upstream.Upstream`.
    *   `Register`: Connects to the database using `database/sql`.
    *   Iterates over configured `calls`.
    *   Registers a `SQLTool` for each call.
*   **`SQLTool`**: Implements `tool.Tool`.
    *   `Execute`:
        *   Extracts arguments from `ExecutionRequest` based on `parameter_order`.
        *   Executes the prepared statement/query.
        *   Scans rows into a map/slice structure.
        *   Returns JSON-serializable result.

### 3. Factory (`server/pkg/upstream/factory`)
*   Update `NewUpstream` to handle `SqlService` case.

### 4. Dependencies
*   Add `github.com/lib/pq` for Postgres support.
*   Use existing `modernc.org/sqlite` for SQLite support.

## Data Flow
User/LLM -> MCP Server (ExecuteTool) -> SQLTool -> database/sql -> DB -> Result -> MCP Server -> LLM.

## Edge Cases
*   **SQL Injection**: Prevented by using parameterized queries (e.g. `$1` in Postgres, `?` in SQLite/MySQL). `parameter_order` maps inputs to these placeholders.
*   **Type Conversion**: `database/sql` handles most types. Complex types might need explicit handling (e.g. JSONB), but for now standard types are supported.
*   **Connection Errors**: Handled gracefully, returning error to caller.
