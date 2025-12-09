# 🏁 Getting Started

This guide provides a step-by-step walkthrough to get the MCP Any up and running on your local machine. By following these instructions, you'll be able to build the project, run the server, and verify that everything is working correctly.

## Prerequisites

Before you begin, ensure you have the following installed:

- [Go](https://golang.org/doc/install) (version 1.24.3 or higher)
- [Docker](https://docs.docker.com/get-docker/)
- [Make](https://www.gnu.org/software/make/)

## Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/mcpany/core.git
   cd core
   ```

2. **Build the application:**
   This command will **automatically install all required dependencies** (tools, linters, generators), generate the necessary protobuf files, and build the server executable into `./build/bin/server`.

   ```bash
   make build
   ```

## Running the Server

After building the project, you can run the server application:

```bash
make run
```

This will start the MCP Any server. It acts as a unified host for all your configured tools (gRPC, HTTP, etc.), managing them concurrently. By default, the server will listen for JSON-RPC requests on port `50050`.

You should see log messages indicating that the server has started, for example:

```
INFO Starting MCP Any server locally... service=mcpany
INFO Configuration jsonrpc-port=50050 registration-port= grpc-port= stdio=false config-paths=[] service=mcpany
INFO Application started service=mcpany
INFO Application started service=mcpany
```

## 🐳 Running with Docker Compose

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

## ☸️ Running with Helm

For deployments to Kubernetes, a Helm chart is available in the `helm/mcpany` directory. See the [Helm chart README](../helm/mcpany/README.md) for detailed instructions.
