# MCP-XY: Model Context Protocol eXtensions

MCP-XY is a powerful and flexible server that acts as a universal adapter for backend services. It dynamically discovers and registers capabilities from various sources—such as gRPC services, RESTful APIs (via OpenAPI specifications), and even command-line tools—and exposes them as standardized "Tools." These tools can then be listed and executed through a unified API, simplifying the integration of diverse services into a single, coherent system.

## Key Features

- **Dynamic Tool Registration**: Automatically discover and register tools from various backend services, either through a dynamic gRPC API or a static configuration file.
- **Multiple Service Types**: Supports a wide range of service types, including:
  - **gRPC**: Register services from `.proto` files or by using gRPC reflection.
  - **OpenAPI**: Ingest OpenAPI (Swagger) specifications to expose RESTful APIs as tools.
  - **HTTP**: Expose any HTTP endpoint as a tool.
  - **Stdio**: Wrap any command-line tool that communicates over standard I/O.
  - **MCP-XY Proxy**: Proxy and re-expose tools from another MCP-XY instance.
- **Upstream Authentication**: Securely connect to your backend services using:
  - **API Keys**
  - **Bearer Tokens**
  - **Basic Auth**
- **Unified API**: Interact with all registered tools through a single, consistent API based on the [Model Context Protocol](httpshttps://modelcontext.protocol.ai/).
- **Extensible**: Designed to be easily extended with new service types and capabilities.

## Getting Started

Follow these instructions to get MCP-XY set up and running on your local machine.

### Prerequisites

Before you begin, ensure you have the following installed:
- [Go](https://golang.org/doc/install) (version 1.21 or later)
- [Docker](https://docs.docker.com/get-docker/)
- [Make](https://www.gnu.org/software/make/)

### Installation & Setup

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/mcpxy/core.git
    cd core
    ```

2.  **Install dependencies and generate code:**
    This command will download the necessary Go modules and generate the required protobuf files.
    ```bash
    make prepare
    ```

### Running the Server

You can run the MCP-XY server using a `make` command, which handles building and running the application.

```bash
make server
```

By default, the server will start and listen for JSON-RPC requests on port `50050` and gRPC registration requests on port `50051`.

## Configuration

MCP-XY can be configured to register services at startup using configuration files. You can specify one or more configuration files or directories using the `--config-paths` flag.

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
make server ARGS="--config-paths ./config.yaml"
```

The server also supports configuration via environment variables. For example, you can set the JSON-RPC port with `MCPXY_JSONRPC_PORT=6000`.

## Usage

Once the server is running, you can interact with it using its JSON-RPC API. For instructions on how to connect `mcpxy` with your favorite AI coding assistant, see the **[Integration Guide](docs/integrations.md)**.

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

For a containerized setup, you can use the provided `docker-compose.yml` file. This will build and run the `mcpxy` server along with a sample `http-echo-server` to demonstrate how `mcpxy` connects to other services in a Docker network.

1.  **Start the services:**
    ```bash
    docker-compose up --build
    ```
    This command will build the Docker images for both the `mcpxy` server and the echo server, and then start them. The `mcpxy` server is configured via `docker/config.docker.yaml` to automatically discover the echo server.

2.  **Test the setup:**
    Once the services are running, you can call the `echo` tool from the `http-echo-server` through the `mcpxy` JSON-RPC API:
    ```bash
    curl -X POST -H "Content-Type: application/json" \
      -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "docker-http-echo/-/echo", "arguments": {"message": "Hello from Docker!"}}, "id": 3}' \
      http://localhost:50050
    ```
    You should receive a response echoing your message.

3.  **Shut down the services:**
    ```bash
    docker-compose down
    ```

## Development

The following commands are available for development:

- `make help`: Show this help message.
- `make server`: Run the main server application.
- `make build`: Build the main server application.
- `make test`: Run all tests.
- `make check`: Run all checks (lint, vet, etc.).
- `make proto-gen`: Generate protobuf files.
- `make docker-build`: Build the Docker image for the server.
- `make clean`: Clean up build artifacts.

## Contributing

Contributions are welcome! Please feel free to open an issue or submit a pull request.

## License

This project is licensed under the terms of the [LICENSE](LICENSE) file.