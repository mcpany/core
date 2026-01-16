# SQL Upstream

The SQL Upstream allows you to expose a SQL database as a set of MCP tools.

## Features

- **Safe Queries**: Define specific SQL queries that can be executed as tools.
- **Parameter Binding**: Use positional parameters and map them to input arguments safely.
- **Read-Only Mode**: Enforce read-only access for safety.
- **Multiple Dialects**: Supports PostgreSQL, MySQL, and SQLite.

## Configuration

```yaml
upstreams:
  my-db:
    type: sql
    sql_service:
      driver: postgres
      dsn: "postgres://user:pass@localhost:5432/dbname"
      calls:
        get_user:
          query: "SELECT * FROM users WHERE id = $1"
          parameter_order: ["id"]
          input_schema:
            type: object
            properties:
              id:
                type: string
                description: "The user ID"
            required: ["id"]
          output_schema:
            type: array
            items:
              type: object
```

### Explanation

- **`driver`**: The database driver (`postgres`, `mysql`, `sqlite`).
- **`dsn`**: Connection string.
- **`calls`**: Map of tool names to SQL queries.
- **`query`**: The SQL query string. Use placeholders appropriate for your driver (e.g., `$1` for Postgres, `?` for MySQL/SQLite).
- **`parameter_order`**: List of input property names that map to the positional placeholders in the query.
- **`input_schema`**: JSON Schema defining the input arguments for the tool.
