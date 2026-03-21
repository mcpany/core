# Automated Documentation Generation

MCP Any includes a built-in tool to generate Markdown documentation for your configured services and tools. This ensures that your documentation always stays in sync with your configuration.

## Usage

You can generate documentation using the `config doc` command:

```bash
mcpany config doc --config-path /path/to/your/config.yaml
```

This will print the generated Markdown to stdout. You can redirect it to a file:

```bash
mcpany config doc --config-path config.yaml > docs/api.md
```

## What gets generated?

The documentation includes:
*   Service names and descriptions
*   Tool names, descriptions, and arguments
*   Prompts and resources (if applicable)

This is useful for providing up-to-date reference material for the AI assistants or developers using your MCP Any server.
