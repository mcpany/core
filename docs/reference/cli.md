# CLI Reference

This document provides a reference for the MCP Any command-line interface (CLI).

## `docs`

Generate Markdown documentation for registered tools based on the configuration. This command loads your configuration files, initializes the tools (including dynamic discovery for services that support it), and outputs a Markdown file listing all available tools and their input schemas.

### Usage

```bash
mcpany docs [flags]
```

### Flags

| Flag | Shorthand | Description | Default |
| :--- | :--- | :--- | :--- |
| `--config-path` | | Paths to configuration files or directories. | `[]` |
| `--output` | `-o` | Output file path. If not specified, prints to stdout. | `""` |
| `--profiles` | | Comma-separated list of active profiles. | `default` |

### Examples

**Generate documentation to stdout:**

```bash
mcpany docs --config-path config.yaml
```

**Save documentation to a file:**

```bash
mcpany docs --config-path config.yaml -o tools.md
```

**Generate documentation for a specific profile:**

```bash
mcpany docs --config-path config.yaml --profiles prod -o production_tools.md
```
