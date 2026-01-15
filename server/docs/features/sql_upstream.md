# SQL Upstream

The SQL Upstream allows you to expose a SQL database as a set of MCP tools.

## Features

- **Safe Queries**: Define specific SQL queries that can be executed as tools.
- **Parameter Binding**: Use named parameters (or positional depending on driver) to safely inject user input.
- **Read-Only Mode**: Enforce read-only access for safety (via database user permissions or careful query design).
- **Multiple Dialects**: Supports PostgreSQL, MySQL, and SQLite.

## Configuration

```yaml
upstreams:
  my-db:
    type: sql
    config:
      driver: postgres
      dsn: "postgres://user:pass@localhost:5432/dbname"
      # Map of tool names to query definitions
      calls:
        get_user:
          # The SQL query to execute.
          # Use $1, $2, etc. for Postgres; ? for MySQL/SQLite
          query: "SELECT * FROM users WHERE id = $1"
          # Helper description for the LLM
          # description: "Get user by ID" (Note: description is inferred from tool name or set manually if supported)

          # Order of parameters to bind to the query placeholders ($1, $2...)
          parameter_order:
            - "id"

          # JSON Schema for the input arguments
          input_schema:
            type: object
            properties:
              id:
                type: string
                description: "The user ID"
            required: ["id"]
```
