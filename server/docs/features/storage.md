# Storage Configuration

MCP Any supports different storage backends for persisting configuration (Upstream Services, Profiles, etc.). The storage backend is configured via the global settings.

## Supported Drivers

### SQLite (Default)

SQLite is the default storage driver. It uses a local file to store data.

**Configuration:**

```yaml
global_settings:
  db_driver: sqlite
  db_path: mcpany.db # Path to the SQLite database file
```

### PostgreSQL

PostgreSQL is supported for High Availability (HA) deployments.

**Configuration:**

```yaml
global_settings:
  db_driver: postgres
  db_dsn: "postgres://user:password@localhost:5432/mcpany?sslmode=disable"
```

## Migration

Currently, MCP Any automatically initializes the schema on startup for both SQLite and PostgreSQL.

**Note:** Automatic schema migration for updates is not yet fully supported. Please ensure you back up your data before upgrading.
