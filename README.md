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

## ❓ Why MCP Any?

- **No Code Required**: Create fully functional MCP servers for your APIs just by writing a config file.
- **Unified Runtime**: Stop spawning a new npx or python process for every single tool. Run one efficient `mcpany` server that manages all your connections and tools in a single place.
- **Shareable Configurations**: Share your MCP server setups publicly. Users don't need to download unsafe binaries or set up complex environments—they just load your config.
- **Local & Secure**: Host your MCP server locally. Connect to your private or public APIs without sending sensitive data through third-party remote servers. Perfect for both personal and enterprise use.
- **Universal Adapter**: Dynamically acts as a bridge for gRPC services, RESTful APIs (via OpenAPI), and command-line tools, exposing them as standardized MCP tools.

## ✨ Key Features

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
- **Multi-User & Multi-Profile**: Securely support multiple users with distinct profiles, each with its own set of enabled services and granular authentication.
- **Extensible**: Designed to be easily extended with new service types and capabilities.

## ⚡ Quick Start (5 Minutes)

Ready to give your AI access to real-time data? Let's connect a public Weather API to **Gemini CLI** (or any MCP client) using MCP Any.

### 1. Prerequisites

- **Go**: Ensure you have [Go](https://go.dev/doc/install) installed (1.23+ recommended).
- **Gemini CLI**: If not installed, see the [installation guide](https://docs.cloud.google.com/gemini/docs/codeassist/gemini-cli).

_(Prefer building from source? See [Getting Started](docs/getting_started.md) for build instructions.)_

### 2. Configuration

We will use the pre-built `wttr.in` configuration available in the examples directory: `examples/popular_services/wttr.in/config.yaml`.

### Quick Start: Weather Service

1.  **Run the Server (Docker):**

    ```bash
    docker run -d --rm --name mcpany-server \
      -v $(pwd)/examples:/examples \
      -p 50050:50050 \
      ghcr.io/mcpany/server:dev-latest \
      run --config-path /examples/popular_services/wttr.in/config.yaml
    ```

    > **Tip:** successful debugging requires detailed logs? Add the `--debug` flag to the end of the command:
    >
    > ```bash
    > docker run -d --rm --name mcpany-server \
    >   -v $(pwd)/examples:/examples \
    >   -p 50050:50050 \
    >   ghcr.io/mcpany/server:dev-latest \
    >   run --config-path /examples/popular_services/wttr.in/config.yaml --debug
    > ```
    >
    > You can then inspect the logs (including API call details) with:
    >
    > ```bash
    > docker logs mcpany-server
    > ```

2.  **Connect Gemini CLI:**

    ```bash
    gemini mcp add --transport http --trust mcpany http://localhost:50050
    ```

3.  **Ask about the weather:**

    ````bash
    gemini -m gemini-2.5-flash -p "What is the weather in London?"
    ```ignored if tools are working.)_
    ````

### 6. Chat!

Ask your AI about the weather:

gemini -m gemini-2.5-flash -p "What is the weather in London?"

````

The AI will:

1.  **Call** the tool (e.g., `wttrin_<hash>.get_weather`).
2.  `mcpany` will **proxy** the request to `https://wttr.in`.
3.  The AI receives the JSON response and answers your question!

Ask about the moon phase:

```bash
gemini -m gemini-2.5-flash -p "What is the moon phase?"
````

The AI will:

1.  **Call** the `get_moon_phase` tool.
2.  `mcpany` will **proxy** the request to `https://wttr.in/moon`.
3.  The AI receives the ASCII art response and describes it!

For more complex examples, including gRPC, OpenAPI, and authentication, check out [docs/reference/configuration.md](docs/reference/configuration.md).

## 💡 More Usage

Once the server is running, you can interact with it using its JSON-RPC API.

- For detailed configuration options, see **[Configuration Reference](docs/reference/configuration.md)**.
- For instructions on how to connect `mcpany` with your favorite AI coding assistant, see the **[Integration Guide](docs/integrations.md)**.
- For hands-on examples, see the **[Examples](docs/examples.md)**.
- For monitoring metrics, see **[Monitoring](docs/monitoring.md)**.

## 🤝 Contributing

Contributions are welcome! Please feel free to open an issue or submit a pull request.

## 🗺️ Roadmap

Check out our [Roadmap](docs/roadmap.md) to see what we're working on and what's coming next.

## 📄 License

This project is licensed under the terms of the [LICENSE](LICENSE) file.
