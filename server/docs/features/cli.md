# mcpany: The MCP Any CLI

`mcpany` is the command-line interface for managing and debugging your MCP Any installation.

## Features

- **Configuration Validation**: Check your config files for errors before deploying.
- **Doctor**: Run a health check on your environment and server.

## Usage

### Validation

```bash
mcpany config validate --config-path ./config.yaml
```

### Doctor

```bash
mcpany doctor
```
