# 👨‍💻 Developer Guide

This guide is for developers who want to contribute to the MCP Any. It provides information about the development environment, build process, and other useful tips.

## Development Setup

### Prerequisites

- **Go**: Ensure you have a recent version of Go installed (see `go.mod` for the exact version). You can find installation instructions on the [official Go website](https://golang.org/doc/install).
- **Docker**: Required for building and running Docker images, especially for end-to-end tests.
- **Make**: Used for simplifying common development tasks.
- **Python**: Required for installing and running pre-commit hooks.

### Tool Installation

This project uses a `Makefile` to automate the installation of all necessary development tools, including `protoc`, Go protobuf plugins, linters, and pre-commit hooks.

To install everything you need, simply run:

```bash
make prepare
```

Alternatively, you can install the tools manually:

- **protoc**: The `protoc` compiler is required to generate Go code from `.proto` files.
  - **Find the latest release:** Go to the [protobuf GitHub releases page](https://github.com/protocolbuffers/protobuf/releases).
  - **Download the archive:** Find the `protoc-*-<OS>-<ARCH>.zip` file that matches your operating system and architecture.
  - **Install:** Unzip the archive and move the `bin/protoc` executable to a directory that is in your system's `PATH`.

This command will download and install the correct versions of the tools into a `./build/env` directory, ensuring a consistent development environment without polluting your global system paths.

## Code Structure Overview

The MCP Any codebase is organized into several key packages:

- **`cmd/server`**: Contains the `main` application entry point and command-line interface setup using Cobra.
- **`pkg/app`**: Implements the core application logic, orchestrating the different components.
- **`pkg/service`**: Provides resilience patterns and utilities.
- **`pkg/serviceregistry`**: Handles the registration of upstream services.
- **`pkg/upstream`**: Contains the implementations for connecting to and interacting with various upstream services (gRPC, HTTP, etc.).
- **`pkg/tool`**: Manages the lifecycle, indexing, and execution of tools.
- **`pkg/transformer`**: Handles the conversion of data between the internal MCP Any format and the format of the upstream services.
- **`proto`**: Contains all the protobuf definitions for the project, including API contracts and configuration structures.
- **`tests`**: Contains integration and end-to-end tests.

## Working with Services

MCP Any allows you to extend its capabilities by registering external services, which are then exposed as "Tools."

### Registering Services

Services can be registered with MCP Any in two ways:

1. **Dynamically via the gRPC Registration API**: The `RegistrationService` (`proto/api/v1/registration.proto`) provides an API for registering services at runtime.
2. **Statically via a configuration file**: Services can be defined in a YAML configuration file and loaded when the server starts.

### Service Configuration

The `UpstreamService` message (`proto/config/v1/config.proto`) is the core configuration for defining a service.

For a comprehensive reference for all configuration options, please see the [Configuration Reference](./reference/configuration.md).

### Configuration Examples

Below are some examples of how to configure different upstream services using a static YAML configuration file.

#### gRPC Service with Reflection

This example configures a gRPC service and uses gRPC reflection to automatically discover its tools.

```yaml
upstreamServices:
  - name: "grpc_weather"
    grpcService:
      address: "localhost:50051"
      useReflection: true
```

#### HTTP Service with API Key Authentication

This example configures a generic HTTP service and demonstrates how to secure the connection to the upstream service using an API key.

```yaml
upstreamServices:
  - name: "http_echo"
    httpService:
      address: "http://localhost:8080"
      calls:
        echo:
          id: "echo"
          endpoint_path: "/echo"
          method: "HTTP_METHOD_POST"
    upstream_authentication:
      api_key:
        header_name: "X-Api-Key"
        api_key:
          plain_text: "your-secret-api-key"
```

#### OpenAPI Service

This example configures a service from an OpenAPI specification. MCP Any will parse the specification to discover the available tools.

```yaml
upstreamServices:
  - name: "openapi_petstore"
    openapiService:
      address: "https://petstore.swagger.io/v2"
      spec_content: |
        # You can paste an OpenAPI spec here directly
        # or provide a path to a file using `spec_url`.
        openapi: "2.0"
        info:
          title: "Simple Pet Store API"
          version: "1.0.0"
        paths:
          /pets:
            get:
              operationId: listPets
              responses:
                "200":
                  description: "A paged array of pets"
```

## Generating Documentation

You can automatically generate Markdown documentation for your `mcpany` configuration using the `mcpany` CLI.

```bash
mcpany config doc --config-path ./config.yaml
```

This command will output a Markdown formatted list of all available tools, their descriptions, and input schemas, which is useful for sharing with consumers of your MCP server.

## Makefile Commands

This project uses a Makefile to simplify common development tasks. Run `make` or `make help` to see a list of all available commands.

- `make prepare`: Installs all necessary development tools.
- `make run`: Builds and runs the main server application locally.
- `make build`: Builds the main server application binary.
- `make test`: Runs all unit and integration tests.
- `make lint`: Runs all linters and formatters using pre-commit.
- `make gen`: Generates Go code from protobuf files.
- `make build-docker`: Builds the Docker image for the server.
- `make clean`: Cleans up build artifacts and generated files.
