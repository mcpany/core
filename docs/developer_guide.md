# 👨‍💻 Developer Guide

This guide is for developers who want to contribute to the MCP-X. It provides information about the development environment, build process, and other useful tips.

## Development Setup

### Prerequisites

- **Go**: Ensure you have a recent version of Go installed. You can find installation instructions on the [official Go website](https://golang.org/doc/install).
- **Docker**: Required for building and running the Docker images.
- **Make**: Used for simplifying common development tasks.

### Tool Installation

1.  **Install `protoc` (Protobuf Compiler)**

    The `protoc` compiler is required to generate Go code from `.proto` files.
    - **Find the latest release:** Go to the [protobuf GitHub releases page](https://github.com/protocolbuffers/protobuf/releases).
    - **Download the archive:** Find the `protoc-*-<OS>-<ARCH>.zip` file that matches your operating system and architecture.
    - **Install:** Unzip the archive and move the `bin/protoc` executable to a directory that is in your system's `PATH`.

2.  **Install Go Protobuf Plugins**

    These plugins are used by `protoc` to generate Go-specific code.

    ```bash
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    ```

    Make sure your Go bin directory (`$GOPATH/bin` or `$HOME/go/bin`) is in your system's `PATH`.

## Code Structure Overview

The MCP-X codebase is organized into several key packages:

- **`cmd/server`**: Contains the `main` application entry point.
- **`pkg/server`**: Implements the core MCP-X server logic.
- **`pkg/apiserver`**: Provides the gRPC API for service registration.
- **`pkg/grpc`**: Houses modules related to gRPC service integration.
- **`pkg/openapi`**: Contains modules for integrating services defined by OpenAPI specifications.
- **`proto`**: Contains all the protobuf definitions for the project.

## Working with Services

MCP-X allows you to extend its capabilities by registering external services, which are then exposed as "Tools".

### Registering Services

Services can be registered with MCP-X in two ways:

1.  **Dynamically via the Registration API**: The `RegistrationService` (`proto/api/v1/registration.proto`) provides an API for registering services at runtime.
2.  **Statically via a configuration file**: Services can be defined in a YAML configuration file and loaded when the server starts.

### Service Configuration

The `McpxServerConfig` message (`proto/config/v1/config.proto`) is the root configuration for the entire MCP-X server. It defines the global settings, upstream services, frontend services, and service bindings.

For a comprehensive reference for all configuration options, please see the [Configuration Reference](./reference/configuration.md).

### Configuration Examples

Below are some examples of how to configure different upstream services using a static YAML configuration file.

#### gRPC Service with Reflection

This example configures a gRPC service and uses gRPC reflection to automatically discover its tools.

```yaml
upstream_services:
  - id: "grpc-calculator-service"
    name: "grpc_calculator"
    service_config:
      grpc_service:
        address: "localhost:50051"
        use_reflection: true
```

#### HTTP Service with API Key Authentication

This example configures a generic HTTP service and demonstrates how to secure the connection to the upstream service using an API key.

```yaml
upstream_services:
  - id: "http-echo-service"
    name: "http_echo"
    service_config:
      http_service:
        address: "http://localhost:8080"
        calls:
          - operation_id: "echo"
            endpoint_path: "/echo"
            method: "HTTP_METHOD_POST"
    upstream_authentication:
      api_key:
        header_name: "X-Api-Key"
        api_key: "your-secret-api-key"
```

#### OpenAPI Service

This example configures a service from an OpenAPI specification. MCP-X will parse the specification to discover the available tools.

```yaml
upstream_services:
  - id: "openapi-petstore-service"
    name: "openapi_petstore"
    service_config:
      openapi_service:
        address: "https://petstore.swagger.io/v2"
        openapi_spec: |
          # You can paste an OpenAPI spec here directly
          # or provide a path to a file.
          swagger: "2.0"
          info:
            title: "Simple Pet Store API"
            version: "1.0.0"
          paths:
            /pets:
              get:
                operationId: listPets
                responses:
                  '200':
                    description: "A paged array of pets"
```

## Makefile Commands

This project uses a Makefile to simplify common development tasks.

- `make help`: Show this help message.
- `make server`: Run the main server application.
- `make build`: Build the main server application.
- `make test`: Run all tests.
- `make check`: Run all checks (lint, vet, etc.).
- `make proto-gen`: Generate protobuf files.
- `make docker-build`: Build the docker image for the server.
- `make clean`: Clean up build artifacts.
