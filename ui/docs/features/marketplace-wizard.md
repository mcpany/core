# Marketplace Config Wizard

The Marketplace Config Wizard simplifies the process of instantiating upstream MCP servers. It provides a guided experience to configure server parameters, authentication, and webhooks.

## Smart Templates

For popular MCP servers, the wizard offers "Smart Forms" that replace generic environment variable inputs with tailored configuration fields.

### PostgreSQL

Instantiate a PostgreSQL MCP server by simply providing the connection string.

![Postgres Wizard](../screenshots/wizard_postgres.png)

The wizard automatically constructs the correct command arguments:
`npx -y @modelcontextprotocol/server-postgres <connection-url>`

### Filesystem

Configure the Filesystem MCP server by adding allowed directories through a list interface.

The wizard constructs:
`npx -y @modelcontextprotocol/server-filesystem <path1> <path2> ...`

## Custom Configuration

For other servers, or for advanced customization, the "Manual / Custom" template allows full control over the executable path, arguments, and environment variables.
