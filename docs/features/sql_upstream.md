# SQL Upstream

**Status**: Implemented

The SQL Upstream service allows you to expose a SQL database (PostgreSQL, SQLite, MySQL) as a set of MCP Tools. This enables agents to run pre-defined SQL queries against your database securely.

## Configuration

You define SQL queries in your `config.yaml` under `upstream_services`. Each query becomes a tool.

```yaml
upstream_services:
  - name: "users-db"
    sql_service:
      driver: "postgres"
      dsn: "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
      calls:
        get_user_by_email:
          input_schema:
            type: object
            properties:
              email:
                type: string
          query: "SELECT * FROM users WHERE email = $1"
          parameter_order:
            - "email"
          output_schema:
            type: array
            items:
              type: object
```

## Security

*   **Parameterized Queries**: All inputs are strictly parameterized to prevent SQL injection.
*   **Read-Only**: It is best practice to use a read-only database user for the DSN if the agent should not modify data.
