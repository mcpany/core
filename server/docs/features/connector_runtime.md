# Universal Connector Runtime

The Universal Connector Runtime allows running MCP connectors (stdio-based tools) as managed sidecar processes.

## Usage

```bash
connector-runtime -name <connector_name> -sidecar
```

## Features

- **Sidecar Mode**: Runs alongside the main application.
- **Lifecycle Management**: Handles startup and shutdown of connector processes.
