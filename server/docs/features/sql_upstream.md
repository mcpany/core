# SQL Upstream

The SQL Upstream allows you to expose a SQL database as a set of MCP tools.

## Features

- **Safe Queries**: Define specific SQL queries that can be executed as tools.
- **Parameter Binding**: Use named parameters to safely inject user input.
- **Read-Only Mode**: Enforce read-only access for safety.
- **Multiple Dialects**: Supports PostgreSQL, MySQL, and SQLite.

## Configuration

```yaml
upstream_services:
  - name: "my-db"
    sql_service:
      driver: "postgres"
      dsn: "postgres://user:pass@localhost:5432/dbname"
      calls:
        get_user:
          query: "SELECT * FROM users WHERE id = $1"
          parameter_order: ["id"]
          input_schema:
            type: "object"
            properties:
              id:
                type: "string"
            required: ["id"]
          output_schema:
            type: "object"
            properties:
              id: {type: "string"}
              name: {type: "string"}
```

## Supported Drivers

- `postgres` (PostgreSQL)
- `mysql` (MySQL)
- `sqlite` (SQLite)
