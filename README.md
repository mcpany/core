[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Test](https://github.com/mcpany/core/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/mcpany/core/actions/workflows/ci.yml)
[![GoDoc](https://godoc.org/github.com/mcpany/core?status.png)](https://pkg.go.dev/github.com/mcpany/core)
[![GoReportCard](https://goreportcard.com/badge/github.com/mcpany/core)](https://goreportcard.com/report/github.com/mcpany/core)
[![codecov](https://codecov.io/gh/mcpany/core/branch/main/graph/badge.svg)](https://codecov.io/gh/mcpany/core)

<p align="center">
  <img src="docs/images/logo.png" alt="MCP Any Logo" width="200"/>
</p>

# MCP Any: Configuration-Driven MCP Server

**Eliminate the need to build and maintain custom MCP servers for every API.**

MCP Any empowers you to create robust Model Context Protocol (MCP) servers using **simple configurations**. Instead of writing code, compiling binaries, and managing complex deployments, you define your tools, resources, and prompts in portable configuration files.

## Why MCP Any?

- **No Code Required**: Create fully functional MCP servers for your APIs just by writing a config file.
- **Shareable Configurations**: Share your MCP server setups publicly. Users don't need to download unsafe binaries or set up complex environments—they just load your config.
- **Local & Secure**: Host your MCP server locally. Connect to your private or public APIs without sending sensitive data through third-party remote servers. Perfect for both personal and enterprise use.
- **Universal Adapter**: Dynamically acts as a bridge for gRPC services, RESTful APIs (via OpenAPI), and command-line tools, exposing them as standardized MCP tools.

## Key Features

- **Dynamic Tool Registration**: Automatically discover and register tools from various backend services, either through a dynamic gRPC API or a static configuration file.
- **Multiple Service Types**: Supports a wide range of service types, including:
  - **gRPC**: Register services from `.proto` files or by using gRPC reflection.
  - **OpenAPI**: Ingest OpenAPI (Swagger) specifications to expose RESTful APIs as tools.
  - **HTTP**: Expose any HTTP endpoint as a tool.
- **GraphQL**: Expose a GraphQL API as a set of tools, with the ability to customize the selection set for each query.
- **Advanced Service Policies**: Configure [Caching](docs/caching.md) and Rate Limiting to optimize performance and protect upstream services.
- **MCP Any Proxy**: Proxy and re-expose tools from another MCP Any instance.
- **Upstream Authentication**: Securely connect to your backend services using:
  - **API Keys**
  - **Bearer Tokens**
  - **Basic Auth**
  - **mTLS**
- **Unified API**: Interact with all registered tools through a single, consistent API based on the [Model Context Protocol](https://modelcontext.protocol.ai/).
- **Extensible**: Designed to be easily extended with new service types and capabilities.

### Defining Prompts

MCP Any allows you to define and execute prompts directly from your configuration files. This is useful for integrating with AI models and other services that require dynamic, template-based inputs.

Here's an example of how to define a prompt in your `config.yaml`:

```yaml
upstreamServices:
  - name: "my-prompt-service"
    httpService:
      address: "https://api.example.com"
      prompts:
        - name: "my-prompt"
          description: "A sample prompt"
          messages:
            - role: "user"
              text:
                text: "Hello, {{name}}!"
```

You can then execute this prompt by sending a `prompts/get` request to the server:

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "prompts/get", "params": {"name": "my-prompt-service.my-prompt", "arguments": {"name": "world"}}, "id": 1}' \
  http://localhost:50050
```

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

### Running with Docker

You can also run the server using Docker. The official image is available on GitHub Container Registry.

1.  **Pull the latest image:**

    ```bash
    docker pull ghcr.io/mcpany/core:latest
    ```

2.  **Run the server:**

    ```bash
    docker run --rm -p 50050:50050 -p 50051:50051 ghcr.io/mcpany/core:latest
    ```

    This will start the server and expose the JSON-RPC and gRPC ports to your local machine.

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
```

To run the server with this configuration, use the following command:

```bash
make run ARGS="--config-paths ./config.yaml"
```

The server also supports configuration via environment variables. For example, you can set the JSON-RPC port with `MCPANY_JSONRPC_PORT=6000`.

### Validating Configuration

You can validate your configuration files without starting the server using the `config validate` command:

```bash
./build/bin/server config validate --config-path ./config.yaml
```

If the configuration is valid, the command will print a success message and exit with a status code of 0. If the configuration is invalid, the command will print an error message and exit with a non-zero status code.

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

- **GraphQL Services**: Register a GraphQL service and customize the selection set for a query.

  ```yaml
  upstreamServices:
    - name: "my-graphql-service"
      graphqlService:
        address: "http://localhost:8080/graphql"
        calls:
          - name: "user"
            selectionSet: "{ id name }"
  ```

- **Stdio Services**: Wrap a command-line tool that communicates over stdio.

  ```yaml
  upstreamServices:
    - name: "my-stdio-service"
      stdioService:
        command: "my-tool"
        args: ["--arg1", "value1"]
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

  You can also source secrets from environment variables. This is the recommended approach for production environments.

  ```yaml
  upstreamServices:
    - name: "my-secure-service"
      httpService:
        address: "https://api.example.com"
        # ...
      upstreamAuthentication:
        apiKey:
          headerName: "X-API-Key"
          apiKey:
            environmentVariable: "MY_API_KEY"
  ```

- **mTLS**: Configure mTLS for an upstream service.

  ```yaml
  upstreamServices:
    - name: "my-mtls-service"
      httpService:
        address: "https://api.example.com"
        # ...
      upstreamAuthentication:
        mtls:
          clientCertPath: "/path/to/client.crt"
          clientKeyPath: "/path/to/client.key"
          caCertPath: "/path/to/ca.crt"
  ```

- **Resilience**: Configure retries for a gRPC service.
- **Server Authentication**: Secure the MCP server with an API key.

  ```yaml
  global_settings:
    api_key: "my-secret-key"
  ```

### Remote Configurations

In addition to loading configuration files from the local filesystem, MCP Any can also load configurations from remote URLs. This allows you to easily share and reuse configurations without having to manually copy and paste files.

To load a remote configuration, simply provide the URL as a value to the `--config-paths` flag:

```bash
make run ARGS="--config-paths https://example.com/my-config.yaml"
```

**Security Warning:** Loading configurations from remote URLs can be dangerous if you do not trust the source. Only load configurations from trusted sources to avoid potential security risks.

### Loading Configurations from Git Repositories

In addition to loading configuration files from the local filesystem and remote URLs, MCP Any can also load configurations from Git repositories. This allows you to easily share and reuse configurations without having to manually copy and paste files.

To load a configuration from a Git repository, use the `--config-git-url` and `--config-git-path` flags. The `--config-git-url` flag specifies the URL of the Git repository, and the `--config-git-path` flag specifies the path to the configuration file within the repository.

Here is an example of how to load a configuration from a Git repository:

```bash
make run ARGS="--config-git-url https://github.com/mcpany/examples.git --config-git-path basic-http-service/config.yaml"
```

This command will clone the `mcpany/examples` repository, load the `config.yaml` file from the `basic-http-service` directory, and then start the server with the loaded configuration.

## Usage

Once the server is running, you can interact with it using its JSON-RPC API. For instructions on how to connect `mcpany` with your favorite AI coding assistant, see the **[Integration Guide](docs/integrations.md)**.

### Configuration Generator

MCP Any includes a CLI tool to help you generate configuration files interactively. To use it, run the following command:

```bash
go run cmd/mcp-any-cli/main.go
```

The tool will prompt you for the information needed to generate a configuration file for a specific service type.

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

1.  **Start the services:**

    ```bash
    docker compose up --build
    ```

    This command will build the Docker images for both the `mcpany` server and the echo server, and then start them. The `mcpany` server is configured via `docker/config.docker.yaml` to automatically discover the echo server.

2.  **Test the setup:**
    Once the services are running, you can call the `echo` tool from the `http-echo-server` through the `mcpany` JSON-RPC API:

    ```bash
    curl -X POST -H "Content-Type: application/json" \
      -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "docker-http-echo/-/echo", "arguments": {"message": "Hello from Docker!"}}, "id": 3}' \
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
