# Hot Reloading

MCP Any supports dynamic configuration reloading without restarting the server.

## How it works

The server watches the configuration file (and referenced files) for changes. When a change is detected:

1. The server debounces the events to avoid rapid reloads.
2. It parses the new configuration.
3. If the configuration is valid, it applies the changes (e.g., updating upstream services, policies).
4. If the configuration is invalid, it logs an error and keeps the old configuration active.

## Supported Changes

- Adding/Removing Upstream Services
- Updating Authentication/Policy configurations
- Modifying Logging settings

## Best Practices

- Use atomic saves (e.g., `mv new.yaml config.yaml`) to ensure the server reads a complete file.
