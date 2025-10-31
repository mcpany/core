[![Test](https://github.com/mcpany/core/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/mcpany/core/actions/workflows/ci.yml)
[![GoDoc](https://godoc.org/github.com/mcpany/core?status.png)](https://pkg.go.dev/github.com/mcpany/core)
[![GoReportCard](https://goreportcard.com/badge/github.com/mcpany/core)](https://goreportcard.com/report/github.com/mcpany/core)
[![codecov](https://codecov.io/gh/mcpany/core/branch/main/graph/badge.svg)](https://codecov.io/gh/mcpany/core)

# MCP Any: Convert Anything to MCP Server

Why developing multiple MCP servers for each API when you can just have one to adapt to all?

MCP Any is a powerful and flexible server that acts as a universal adapter for backend services. It dynamically discovers and registers capabilities from various sources—such as gRPC services, RESTful APIs (via OpenAPI specifications), and even command-line tools—and exposes them as standardized "Tools." These tools can then be listed and executed through a unified API, simplifying the integration of diverse services into a single, coherent system.

## Architecture

MCP Any is built on a modular and extensible architecture. The core components are:

- **MCP Server**: The main server that implements the [Model Context Protocol](https://modelcontext.protocol.ai/). It handles incoming requests and orchestrates the other components.
- **Service Registry**: Manages the lifecycle of upstream services. It is responsible for creating and registering services from configuration files or dynamic registration requests.
- **Tool Manager**: Keeps track of all the tools that are discovered from the upstream services. It provides a unified interface for executing tools, regardless of their underlying implementation.
- **Upstream Services**: These are the backend services that MCP Any connects to. Each service type (gRPC, HTTP, etc.) has a corresponding implementation that knows how to interact with the service and expose its capabilities as tools.
- **Connection Pool**: Manages connections to upstream services to improve performance and resource usage.

## Key Features

- **Dynamic Tool Registration**: Automatically discover and register tools from various backend services, either through a dynamic gRPC API or a static configuration file.
- **Multiple Service Types**: Supports a wide range of service types, including:
  - **gRPC**: Register services from `.proto` files or by using gRPC reflection.
  - **OpenAPI**: Ingest OpenAPI (Swagger) specifications to expose RESTful APIs as tools.
  - **HTTP**: Expose any HTTP endpoint as a tool.
  - **Stdio**: Wrap any command-line tool that communicates over standard I/O.
  - **MCP Any Proxy**: Proxy and re-expose tools from another MCP Any instance.
- **Upstream Authentication**: Securely connect to your backend services using:
  - **API Keys**
  - **Bearer Tokens**
  - **Basic Auth**
- **Unified API**: Interact with all registered tools through a single, consistent API based on the [Model Context Protocol](https://modelcontext.protocol.ai/).
- **Extensible**: Designed to be easily extended with new service types and capabilities.

## Getting Started

Follow these instructions to get MCP Any set up and running on your local machine.

### Prerequisites

Before you begin, ensure you have the following installed:

- [Go](https://golang.org/doc/install) (version 1.24.3 or later)
- [Docker](https://docs.docker.com/get-docker/)
- [Make](https://www.gnu.org/software/make/)

### Installation & Setup

1. **Clone the repository:**

   ```bash
   git clone https://github.com/mcpany/core.git
   cd core
   ```

2. **Build the application:**
   This command will install dependencies, generate code, and build the `mcpany` binary.

   ```bash
   make build
   ```

   The binary will be located at `./build/bin/server`.

### Running the Server

You can run the MCP Any server directly or by using a `make` command.

- **Directly:**

  ```bash
  ./build/bin/server
  ```

- **Via Make:**

  ```bash
  make run
  ```

By default, the server will start and listen for JSON-RPC requests on port `50050` and gRPC registration requests on port `50051`.

## Configuration

MCP Any can be configured to register services at startup using configuration files. You can specify one or more configuration files or directories using the `--config-paths` flag. The configuration files can be in YAML, JSON, or textproto format.

### Example Configuration

Here is an example of a `config.yaml` file that registers an HTTP service with a single tool:

```yaml
# config.yaml
upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "https://api.example.com"
      calls:
        - operationId: "get_user"
          description: "Get user by ID"
          method: "HTTP_METHOD_GET"
          endpointPath: "/users/{userId}"
          parameterMappings:
            - inputParameterName: "userId"
              targetParameterName: "userId"
              location: "PATH"
```

To run the server with this configuration, use the following command:

```bash
make run ARGS="--config-paths ./config.yaml"
```

The server also supports configuration via environment variables. For example, you can set the JSON-RPC port with `MCPANY_JSONRPC_PORT=6000`.

### Advanced Configuration

MCP Any supports a variety of advanced configuration options, including:

- **gRPC Services**: Register a gRPC service using reflection.

  ```yaml
  upstreamServices:
    - name: "my-grpc-service"
      grpcService:
        address: "localhost:50052"
        reflection:
          enabled: true
  ```

- **OpenAPI Services**: Register a service from an OpenAPI specification.

  ```yaml
  upstreamServices:
    - name: "my-openapi-service"
      openapiService:
        spec:
          path: "./openapi.json"
  ```

- **Authentication**: Configure authentication for an upstream service.

  ```yaml
  upstreamServices:
    - name: "my-secure-service"
      httpService:
        address: "https://api.example.com"
        # ...
      upstreamAuthentication:
        apiKey:
          headerName: "X-API-Key"
          apiKey: "my-secret-key"
  ```

## Usage

Once the server is running, you can interact with it using its JSON-RPC API. For instructions on how to connect `mcpany` with your favorite AI coding assistant, see the **[Integration Guide](docs/integrations.md)**.

## Examples

For hands-on examples of how to use `mcpany` with different upstream service types and AI tools like Gemini CLI, please see the [examples](examples) directory. Each example includes a README file with detailed instructions.

### Listing Tools

To see the list of all registered tools, you can send a `tools/list` request.

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}' \
  http://localhost:50050
```

### Calling a Tool

To execute a tool, send a `tools/call` request with the tool's name and arguments. Based on the example configuration above, here's how you would call the `get_user` tool:

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "my-http-service/-/get_user", "arguments": {"userId": "123"}}, "id": 2}' \
  http://localhost:50050
```

## Running with Docker Compose

For a containerized setup, you can use the provided `docker-compose.yml` file. This will build and run the `mcpany` server along with a sample `http-echo-server` to demonstrate how `mcpany` connects to other services in a Docker network.

1. **Start the services:**

   ```bash
   docker compose up --build
   ```

   This command will build the Docker images for both the `mcpany` server and the echo server, and then start them. The `mcpany` server is configured via `docker/config.docker.yaml` to automatically discover the echo server.

2. **Test the setup:**
   Once the services are running, you can call the `echo` tool from the `http-echo-server` through the `mcpany` JSON-RPC API:

   ```bash
   curl -X POST -H "Content-Type: application/json" \
     -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "docker-http-echo/-/echo", "arguments": {"message": "Hello from Docker!"}}, "id": 3}' \
     http://localhost:50050
   ```

   You should receive a response echoing your message.

3. **Shut down the services:**

   ```bash
   docker compose down
   ```

## Running with Helm

For deployments to Kubernetes, a Helm chart is available in the `helm/mcpany` directory. See the [Helm chart README](helm/mcpany/README.md) for detailed instructions.

## Development

The following commands are available for development:

- `make help`: Show this help message.
- `make run`: Run the main server application.
- `make build`: Build the main server application.
- `make test`: Run all tests.
- `make check`: Run all checks (lint, vet, etc.).
- `make proto-gen`: Generate protobuf files.
- `make docker-build`: Build the docker image for the server.
- `make clean`: Clean up build artifacts.

## Code Documentation

The Go code in this repository is fully documented with GoDoc comments. You can
view the documentation locally by running a GoDoc server:

```bash
godoc -http=:6060
```

Then, navigate to `http://localhost:6060/pkg/github.com/mcpany/core` in your
browser.

The `pkg` directory contains the core logic of the application, organized into
the following subpackages:

- **`app`**: The main application entry point and server lifecycle management.
- **`auth`**: Authentication strategies for both incoming requests and
  connections to upstream services.
- **`bus`**: A type-safe, topic-based event bus for inter-component
  communication.
- **`client`**: Interfaces and wrappers for gRPC and HTTP clients.
- **`config`**: Configuration loading, parsing, and validation.
- **`consts`**: Application-wide constants.
- **`logging`**: The global logger initialization and access.
- **`mcpserver`**: The core MCP server implementation, including request routing
  and handling.
- **`middleware`**: MCP middleware for handling concerns like logging, CORS, and
  authentication.
- **`pool`**: A generic connection pool for managing upstream client
  connections.
- **`serviceregistry`**: The service registry for managing the lifecycle of
  upstream services.
- **`tool`**: Tool management and execution.
- **`upstream`**: Implementations for various upstream service types (gRPC,
  HTTP, etc.).

For more detailed information on each package and its components, please refer
to the GoDoc comments in the source code.

## Contributing

Contributions are welcome! Please feel free to open an issue or submit a pull request.

## License

This project is licensed under the terms of the [LICENSE](LICENSE) file.
