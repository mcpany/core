[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Test](https://github.com/mcpany/core/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/mcpany/core/actions/workflows/ci.yml)
[![GoDoc](https://godoc.org/github.com/mcpany/core?status.png)](https://pkg.go.dev/github.com/mcpany/core)
[![GoReportCard](https://goreportcard.com/badge/github.com/mcpany/core)](https://goreportcard.com/report/github.com/mcpany/core)
[![codecov](https://codecov.io/gh/mcpany/core/branch/main/graph/badge.svg)](https://codecov.io/gh/mcpany/core)

# MCP Any: Convert Anything to MCP Server

Why developing multiple MCP servers for each API when you can just have one to adapt to all?

MCP Any is a powerful and flexible server that acts as a universal adapter for backend services. It dynamically discovers and registers capabilities from various sources—such as gRPC services, RESTful APIs (via OpenAPI specifications), and even command-line tools—and exposes them as standardized "Tools." These tools can then be listed and executed through a unified API, simplifying the integration of diverse services into a single, coherent system.

## Architecture Overview

MCP Any is composed of a central server and a pluggable system of upstream services. The core components are:

-   **`cmd/server`**: The main entry point for the MCP Any server. It handles command-line parsing, configuration loading, and the initialization of all other components.
-   **`pkg/`**: This directory contains the core logic of the application, including:
    -   **`serviceregistry`**: Manages the lifecycle of all upstream services.
    -   **`bus`**: An event bus for asynchronous communication between components.
    -   **`tool`**, **`prompt`**, and **`resource`**: Managers for the core MCP concepts that are exposed by upstream services.
-   **Upstream Services**: These are the backend services that MCP Any connects to. Each upstream service type (e.g., gRPC, OpenAPI, HTTP) has a corresponding implementation in the `pkg/upstream` directory.

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

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/mcpany/core.git
    cd core
    ```

2.  **Build the application:**
    This command will install dependencies, generate code, and build the `server` binary.

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

By default, the server will start and listen for JSON-RPC requests on port `50050`.

### Running with Docker

You can also run the server using Docker. The official image is available on GitHub Container Registry.

1.  **Pull the latest image:**

    ```bash
    docker pull ghcr.io/mcpany/core:latest
    ```

2.  **Run the server:**

    ```bash
    docker run --rm -p 50050:50050 ghcr.io/mcpany/core:latest
    ```

    This will start the server and expose the JSON-RPC and gRPC ports to your local machine.

## Configuration

MCP Any can be configured to register services at startup using configuration files. You can specify one or more configuration files or directories using the `--config-paths` flag. The configuration files can be in YAML, JSON, or textproto format. For a detailed reference of all available configuration options, see the **[Configuration Reference](docs/configuration.md)**.

### Example Configuration

Here is an example of a `config.yaml` file that registers an HTTP service with a single tool:

```yaml
# config.yaml
upstreamServices:
  - name: "my-http-service"
    httpService:
      address: "https://api.example.com"
      calls:
        - toolName: "get_user"
          description: "Get user by ID"
          method: "GET"
          endpointPath: "/users/{userId}"
```

To run the server with this configuration, use the following command:

```bash
make run ARGS="--config-paths ./config.yaml"
```

The server also supports configuration via environment variables. For example, you can set the JSON-RPC port with `MCPANY_JSONRPC_PORT=6000`.

## Usage

Once the server is running, you can interact with it using its JSON-RPC API. For instructions on how to connect `mcpany` with your favorite AI coding assistant, see the **[Integration Guide](docs/integrations.md)**.

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
  -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "my-http-service.get_user", "arguments": {"userId": "123"}}, "id": 2}' \
  http://localhost:50050
```

### Listing Prompts

To see the list of all registered prompts, you can send a `prompts/list` request.

```bash
curl -X POST -H "Content-Type: application/json" \
    -d '{"jsonrpc": "2.0", "method": "prompts/list", "id": 3}' \
    http://localhost:50050
```

### Getting a Prompt

To retrieve a specific prompt, send a `prompts/get` request with the prompt's name.

```bash
curl -X POST -H "Content-Type: application/json" \
    -d '{"jsonrpc": "2.0", "method": "prompts/get", "params": {"name": "my-prompt-service.my_prompt"}, "id": 4}' \
    http://localhost:50050
```

## Examples

For hands-on examples of how to use `mcpany` with different upstream service types and AI tools like Gemini CLI, please see the [examples](examples) directory. Each example includes a README file with detailed instructions.

## Running with Docker Compose

For a containerized setup, you can use the provided `docker-compose.yml` file. This will build and run the `mcpany` server along with a sample `http-echo-server` to demonstrate how `mcpany` connects to other services in a Docker network.

1.  **Start the services:**

    ```bash
    docker compose up --build
    ```

    This command will build the Docker images for both the `mcpany` server and the echo server, and then start them. The `mcpany` server is configured via `docker/config.docker.yaml` to automatically discover the echo server.

2.  **Test the setup:**
    Once the services are running, you can call the `echo` tool from the `http-echo-server` through the `mcpany` JSON-RPC API:

    ```bash
    curl -X POST -H "Content-Type: application/json" \
      -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "docker-http-echo.echo", "arguments": {"message": "Hello from Docker!"}}, "id": 3}' \
      http://localhost:50050
    ```

    You should receive a response echoing your message.

3.  **Shut down the services:**

    ```bash
    docker compose down
    ```

## Running with Helm

For deployments to Kubernetes, a Helm chart is available in the `helm/mcpany` directory. See the [Helm chart README](helm/mcpany/README.md) for detailed instructions.

## Contributing

Contributions are welcome! Please feel free to open an issue or submit a pull request.

## License

This project is licensed under the terms of the [LICENSE](LICENSE) file.
