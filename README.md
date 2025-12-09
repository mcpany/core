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
- **Extensible**: Designed to be easily extended with new service types and capabilities.

## ⚡ Quick Start (5 Minutes)

Ready to give your AI access to real-time data? Let's connect a public Weather API to **Gemini CLI** (or any MCP client) using MCP Any.

### 1. Prerequisites

- **Docker**: Ensure you have [Docker](https://docs.docker.com/get-docker/) installed and running.
- **Gemini CLI**: If not installed, see the [installation guide](https://docs.cloud.google.com/gemini/docs/codeassist/gemini-cli).

_(Prefer building from source? See [Getting Started](docs/getting_started.md) for build instructions.)_

### 2. Create Configuration

Create a file named `weather-config.yaml` in your workspace. We will wrap `wttr.in`, a simple public weather service.

**File Architecture:**

```
.
└── weather-config.yaml # The config file we are creating
```

**`weather-config.yaml`**:

```yaml
global_settings:
  mcp_listen_address: "0.0.0.0:50050"
  log_level: "INFO"

upstream_services:
  - name: "weather"
    http_service:
      address: "https://wttr.in"
      calls:
        get_weather:
          id: "get_weather"
          endpoint_path: "/{city}?format=j1"
          method: "HTTP_METHOD_GET"
          parameters:
            - schema:
                name: "city"
                type: "STRING"
      tools:
        - name: "get_weather"
          description: "Get current weather for a specific city."
          call_id: "get_weather"
          input_schema:
            type: "object"
            properties:
              city:
                type: "string"
```

### 3. Run with Docker

Start the server using the official image. We mount the current directory so the container can read your config.

```bash
docker run --rm \
  -v $(pwd)/weather-config.yaml:/weather-config.yaml \
  -p 50050:50050 \
  ghcr.io/mcpany/core:latest \
  run --config-path /weather-config.yaml
```

The server will start and listen on port `50050`.

### 4. Connect Your AI

Tell your AI assistant how to reach the server.

**For Gemini CLI:**

Connect to the running HTTP server:

```bash
gemini mcp add --transport http --trust weather-server http://localhost:50050
```

_(Note: You may see an unrelated "GitHub requires OAuth" error in the CLI; this can be ignored if tools are working.)_

### 5. Chat!

Ask your AI about the weather:

```bash
gemini -m gemini-2.5-flash -p "What is the weather in London?"
```

The AI will:

1.  **Call** the `weather.get_weather` tool with `city="London"`.
2.  `mcpany` will **proxy** the request to `https://wttr.in/London?format=j1`.
3.  The AI receives the JSON response and answers your question!

---

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
