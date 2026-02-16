# Dynamic Tool Registration

MCP Any supports dynamic registration and auto-discovery of tools from upstream services.

## Overview

Instead of manually defining every tool, you can point MCP Any to a service definition (e.g., OpenAPI spec, gRPC reflection, GraphQL schema), and it will automatically generate the corresponding MCP tools.

## Supported Sources

- **OpenAPI**: Parses Swagger/OpenAPI specs to create tools for each operation.
- **gRPC**: Uses Server Reflection or Proto files to discover methods and messages.
- **GraphQL**: Introspects the schema to create tools for Queries and Mutations.

## Configuration

To enable dynamic registration, you generally set `auto_discover_tool: true` (or it is implied by the specific service config) and provide the necessary connection details.

### OpenAPI

You can provide the OpenAPI spec either via a URL or directly as content.

**Using a Spec URL:**

```yaml
upstream_services:
  - name: "petstore"
    auto_discover_tool: true
    openapi_service:
      address: "https://petstore.swagger.io/v2"
      spec_url: "https://petstore.swagger.io/v2/swagger.json"
```

**Using Spec Content:**

```yaml
upstream_services:
  - name: "internal-api"
    auto_discover_tool: true
    openapi_service:
      address: "http://localhost:8080"
      spec_content: |
        openapi: 3.0.0
        info:
          title: Internal API
          version: 1.0.0
        paths:
          /users:
            get:
              operationId: listUsers
              summary: List all users
              responses:
                '200':
                  description: A list of users
```

### gRPC

For gRPC services, you can use Server Reflection or provide Proto files.

**Using Server Reflection:**

```yaml
upstream_services:
  - name: "payment-service"
    auto_discover_tool: true
    grpc_service:
      address: "localhost:50051"
      use_reflection: true
      tls_config:
        insecure_skip_verify: true # For local development
```

**Using Proto Files:**

```yaml
upstream_services:
  - name: "inventory-service"
    auto_discover_tool: true
    grpc_service:
      address: "localhost:50052"
      proto_definitions:
        - proto_file:
            file_name: "inventory.proto"
            file_path: "./protos/inventory.proto"
```

### GraphQL

For GraphQL, MCP Any introspects the schema from the endpoint.

```yaml
upstream_services:
  - name: "shopify-store"
    auto_discover_tool: true
    graphql_service:
      address: "https://shop.myshopify.com/api/2023-01/graphql.json"
```

## Benefits

- **Reduced Maintenance**: Automatic updates when upstream APIs change.
- **Consistency**: Ensures tools match the actual service capabilities.
- **Instant Onboarding**: Connect to any standard API without writing adapter code.
