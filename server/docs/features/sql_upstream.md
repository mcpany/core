# SQL Upstream

The SQL Upstream allows you to expose a SQL database as a set of MCP tools.

## Features

- **Safe Queries**: Define specific SQL queries that can be executed as tools.
- **Parameter Binding**: Use named parameters to safely inject user input.
- **Read-Only Mode**: Enforce read-only access for safety.
- **Multiple Dialects**: Supports PostgreSQL, MySQL, and SQLite.

## Configuration

```yaml
upstreams:
  my-db:
    type: sql
    config:
      driver: postgres
      dsn: "postgres://user:pass@localhost:5432/dbname"
      calls:
        get_user:
          sql: "SELECT * FROM users WHERE id = :id"
          description: "Get user by ID"
          args:
            id: string
```
