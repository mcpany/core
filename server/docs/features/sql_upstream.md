# SQL Upstream

The SQL Upstream allows the MCP Any server to query SQL databases and expose them as tools.

## Configuration

The SQL provider is configured as an `upstream_service` with the `sql_service` block.

### Fields

| Field    | Type     | Description                                               |
| -------- | -------- | --------------------------------------------------------- |
| `driver` | `string` | The database driver to use. Supported: `mysql`, `postgres`, `sqlite`. |
| `dsn`    | `string` | The Data Source Name (connection string).                 |
| `calls`  | `map`    | A map of defined SQL queries to expose as tools.          |

### Call Definition

Each entry in `calls` defines a tool.

| Field          | Type     | Description                                           |
| -------------- | -------- | ----------------------------------------------------- |
| `query`        | `string` | The SQL query to execute. Use `?` for placeholders.   |
| `input_schema` | `object` | JSON Schema for the input arguments.                  |
| `output_schema`| `object` | JSON Schema for the output.                           |

### Example Configuration

#### PostgreSQL

```yaml
upstream_services:
  - name: "users-db"
    sql_service:
      driver: "postgres"
      dsn: "postgres://user:password@localhost:5432/dbname?sslmode=disable"
      calls:
        get_user_by_email:
          query: "SELECT id, name, email FROM users WHERE email = $1"
          input_schema:
            type: "object"
            properties:
              email:
                type: "string"
            required: ["email"]
```

#### SQLite

```yaml
upstream_services:
  - name: "local-db"
    sql_service:
      driver: "sqlite"
      dsn: "./data.db"
      calls:
        list_items:
          query: "SELECT * FROM items"
```
