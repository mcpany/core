# Dynamic Tool Registration

MCP Any supports dynamic registration and auto-discovery of tools from upstream services.

## Overview

Instead of manually defining every tool, you can point MCP Any to a service definition (e.g., OpenAPI spec, gRPC reflection, GraphQL schema), and it will automatically generate the corresponding MCP tools.

## Supported Sources

- **OpenAPI**: Parses Swagger/OpenAPI specs to create tools for each operation.
- **gRPC**: Uses Server Reflection to discover methods and messages.
- **GraphQL**: Introspects the schema to create tools for Queries and Mutations.

## Benefits

- **Reduced Maintenance**: Automatic updates when upstream APIs change.
- **Consistency**: Ensures tools match the actual service capabilities.

## Configuration Examples

### OpenAPI Service

```yaml
upstream_services:
  - name: "petstore-api"
    openapi_service:
      address: "https://petstore.swagger.io/v2"
      spec_url: "https://petstore.swagger.io/v2/swagger.json"
      # Optional: Override/Customize specific tools
      tools:
        - name: "getPetById"
          description: "Retrieves a pet by its ID"
```

### gRPC Service

```yaml
upstream_services:
  - name: "payment-service"
    grpc_service:
      address: "localhost:50051"
      use_reflection: true
      tls_config:
        insecure_skip_verify: true
```
