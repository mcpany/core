# Dynamic Tool Registration

MCP Any supports dynamic registration and auto-discovery of tools from upstream services.

## Overview

Instead of manually defining every tool, you can point MCP Any to a service definition (e.g., OpenAPI spec, gRPC reflection), and it will automatically generate the corresponding MCP tools.

## Supported Sources

- **OpenAPI**: Parses Swagger/OpenAPI specs to create tools for each operation.
- **gRPC**: Uses Server Reflection to discover methods and messages.

## Benefits

- **Reduced Maintenance**: Automatic updates when upstream APIs change.
- **Consistency**: Ensures tools match the actual service capabilities.
