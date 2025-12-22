[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Test](https://github.com/mcpany/core/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/mcpany/core/actions/workflows/ci.yml)
[![GoDoc](https://godoc.org/github.com/mcpany/core?status.png)](https://pkg.go.dev/github.com/mcpany/core)
[![GoReportCard](https://goreportcard.com/badge/github.com/mcpany/core)](https://goreportcard.com/report/github.com/mcpany/core)
[![codecov](https://codecov.io/gh/mcpany/core/branch/main/graph/badge.svg)](https://codecov.io/gh/mcpany/core)

<p align="center">
  <img src="server/docs/images/logo.png" alt="MCP Any Logo" width="200"/>
</p>

# MCP Any: Configuration-Driven MCP Server

**One server, Infinite possibilities.**

MCP Any revolutionizes how you interact with the Model Context Protocol (MCP). It is not just another MCP proxy or aggregator—it is a powerful **Universal Adapter** that turns _any_ API into an MCP-compliant server through simple configuration.

Traditional MCP adoption requires running a separate server binary for every tool or service you want to expose. This leads to "binary fatigue," complex local setups, and maintenance nightmares.

**MCP Any solves this with a Single Binary approach:**

1.  **Install once**: Run a single `mcpany` server instance.
2.  **Configure everything**: Load lightweight YAML/JSON configurations to capability-enable different APIs (REST, gRPC, GraphQL, Command-line).
3.  **Run anywhere**: No need for `npx`, `python`, or language-specific runtimes for each tool.

## ❓ Philosophy: Configuration over Code

We believe you shouldn't have to write and maintain new code just to expose an existing API to your AI assistant.

- **Metamcp / Onemcp vs. MCP Any**: While other tools might proxy existing MCP servers (aggregator pattern), **MCP Any** creates them from scratch using your existing upstream APIs.
- **No More "Sidecar hell"**: Instead of running 10 different containers for 10 different tools, run 1 `mcpany` container loaded with 10 config files.
- **Ops Friendly**: Centralize authentication, rate limiting, and observability in one robust layer.

### Comparison with Traditional MCP Servers

Unlike traditional "Wrapper" MCP servers (like `mcp-server-postgres`, `mcp-server-github`, etc.) which are compiled binaries dedicated to a single service, **MCP Any** is a generic runtime.

| Feature           | Traditional MCP Server (e.g., `mcp-server-postgres`)                    | MCP Any                                                                         |
| :---------------- | :---------------------------------------------------------------------- | :------------------------------------------------------------------------------ |
| **Architecture**  | **Code-Driven Wrapper**: Wraps internal API calls with MCP annotations. | **Config-Driven Adapter**: Maps existing API endpoints to MCP tools via config. |
| **Deployment**    | **1 Binary per Service**: Need 10 different binaries for 10 services.   | **1 Binary for All**: One `mcpany` binary handles N services.                   |
| **Updates**       | **Recompile & Redistribute**: Internal API change = New Binary release. | **Update Config**: API change = Edit YAML/JSON file & reload.                   |
| **Maintenance**   | **High**: Manage dependencies/versions for N projects.                  | **Low**: Upgrade one core server; just swap config files.                       |
| **Extensibility** | Write code (TypeScript/Python/Go).                                      | Write JSON/YAML.                                                                |

Most "popular" MCP servers today are bespoke binaries. If the upstream API changes, you must wait for the maintainer to update the code, release a new version, and then you must redeploy. With **MCP Any**, you simply update your configuration file to match the new API signature—zero downtime, zero recompilation.

## ✨ Key Features

- **Dynamic Tool Registration & Auto-Discovery**: Automatically discover and register tools from various backend services. For gRPC and OpenAPI, simply provide the server URL or spec URL—MCP Any handles the rest (no manual tool definition required).
- **Multiple Service Types**: Supports a wide range of service types, including:
  - **gRPC**: Register services from `.proto` files or by using gRPC reflection.
  - **OpenAPI**: Ingest OpenAPI (Swagger) specifications to expose RESTful APIs as tools.
  - **HTTP**: Expose any HTTP endpoint as a tool.
  - **GraphQL**: Expose a GraphQL API as a set of tools, with the ability to customize the selection set for each query.
  - **SQL**: Connect to SQL databases (Postgres, SQLite, MySQL) and expose safe queries as tools.
  - **WebSocket**: Connect to WebSocket servers.
  - **WebRTC**: Connect to WebRTC services.
- **Advanced Service & Safety Policies**:
  - **Safety**: Control which tools are exposed to the AI to limit context (reduce hallucinations) and prevent dangerous actions (e.g., blocking `DELETE` operations).
  - **Performance**: Configure [Caching](server/docs/caching.md) and Rate Limiting to optimize performance and protect upstream services.
  - **Audit Logging**: Keep a tamper-evident record of all tool executions in a JSON file or **SQLite database** (using SHA-256 hash chaining) for compliance and security auditing.
- **MCP Any Proxy**: Proxy and re-expose tools from another MCP Any instance.
- **Upstream Authentication**: Securely connect to your backend services using:
  - **API Keys**
  - **Bearer Tokens**
  - **Basic Auth**
  - **mTLS**
- **Unified API**: Interact with all registered tools through a single, consistent API based on the [Model Context Protocol](https://modelcontext.protocol.ai/).
- **Multi-User & Multi-Profile**: Securely support multiple users with distinct profiles, each with its own set of enabled services and granular authentication.
- **Advanced Configuration**: Customize tool behavior with [Merge Strategies and Profile Filtering](server/docs/feature/merge_strategy.md).
- **Extensible**: Designed to be easily extended with new service types and capabilities.

## ⚡ Quick Start (5 Minutes)

Ready to give your AI access to real-time data? Let's connect a public Weather API to **Gemini CLI** (or any MCP client) using MCP Any.

### 1. Prerequisites

- **Go**: Ensure you have [Go](https://go.dev/doc/install) installed (1.23+ recommended).
- **Gemini CLI**: If not installed, see the [installation guide](https://docs.cloud.google.com/gemini/docs/codeassist/gemini-cli).

_(Prefer building from source? See [Getting Started](server/docs/developer_guide.md) for build instructions.)_

### 2. Configuration

We will use the pre-built `wttr.in` configuration available in the examples directory: `server/examples/popular_services/wttr.in/config.yaml`.

### Quick Start: Weather Service

1.  **Run the Server:**

    Choose one of the following methods to run the server.

    **Option 1: Remote Configuration (Recommended)**

    Fastest way to get started. No need to clone the repository.

    ```bash
    docker run -d --rm --name mcpany-server \
      -p 50050:50050 \
      ghcr.io/mcpany/server:dev-latest \
      run --config-path https://raw.githubusercontent.com/mcpany/core/main/server/examples/popular_services/wttr.in/config.yaml
    ```

    **Option 2: Local Configuration**

    Best if you want to modify the configuration or use your own. Requires cloning the repository.

    ```bash
    # Clone the repository
    git clone https://github.com/mcpany/core.git
    cd core

    # Run with local config mounted
    docker run -d --rm --name mcpany-server \
      -p 50050:50050 \
      -v $(pwd)/server/examples/popular_services/wttr.in/config.yaml:/config.yaml \
      ghcr.io/mcpany/server:dev-latest \
      run --config-path /config.yaml
    ```

    > **Tip:** Need detailed logs? Add the `--debug` flag to the end of the `run` command.

2.  **Connect Gemini CLI:**

    ```bash
    gemini mcp add --transport http --trust mcpany http://localhost:50050
    ```

3.  **Chat!**

    Ask your AI about the weather:

    ```bash
    gemini -m gemini-2.5-flash -p "What is the weather in London?"
    ```

    The AI will:

    1.  **Call** the tool (e.g., `wttrin_<hash>.get_weather`).
    2.  `mcpany` will **proxy** the request to `https://wttr.in`.
    3.  The AI receives the JSON response and answers your question!

Ask about the moon phase:

```bash
gemini -m gemini-2.5-flash -p "What is the moon phase?"
```

The AI will:

1.  **Call** the `get_moon_phase` tool.
2.  `mcpany` will **proxy** the request to `https://wttr.in/moon`.
3.  The AI receives the ASCII art response and describes it!

For more complex examples, including gRPC, OpenAPI, and authentication, check out [server/docs/reference/configuration.md](server/docs/reference/configuration.md).

## 💡 More Usage

Once the server is running, you can interact with it using its JSON-RPC API.

- For detailed configuration options, see **[Configuration Reference](server/docs/reference/configuration.md)**.
- For instructions on how to connect `mcpany` with your favorite AI coding assistant (Claude Desktop, Cursor, VS Code, JetBrains, Cline), see the **[Integration Guide](server/docs/integrations.md)**.
- For hands-on examples, see the **[Examples](server/docs/examples.md)** and the **[Profile Authentication Example](server/examples/profile_example/README.md)**.
- For monitoring metrics, see **[Monitoring](server/docs/monitoring.md)**.

## 🛠️ Development Guide

### Prerequisites

- **Go**: Version 1.23+
- **Docker**: For running tests and building images.
- **Protoc**: For generating protobuf files (handled by `make prepare`).

### Setup

Run the following command to set up the development environment, including installing tools and dependencies:

```bash
make prepare
```

### Building

To build the server binary locally:

```bash
make build
```

The binary will be located at `build/bin/server`.

### Testing

To run the test suite:

```bash
make test
```

This includes unit tests, integration tests, and end-to-end (E2E) tests.

### Linting

To run linters:

```bash
make lint
```

We use `golangci-lint` and `pre-commit` hooks.

### Project Structure

- **`server/cmd/`**: Entry points for the applications (server, worker, webhooks).
- **`server/pkg/`**: Core library code.
    - **`config/`**: Configuration loading and validation.
    - **`mcpserver/`**: Core MCP server implementation.
    - **`tool/`**: Tool management and execution logic.
    - **`upstream/`**: Upstream service integrations (gRPC, HTTP, etc.).
    - **`util/`**: Utility functions.
- **`proto/`**: Protocol Buffer definitions.
- **`server/examples/`**: Example configurations.
- **`server/docs/`**: Detailed documentation.

### Code Standards

- **Documentation**: All public functions, methods, and types must be documented. Use `go run server/tools/check_doc.go server/` to verify documentation coverage.
- **Linting**: Ensure `make lint` passes before submitting code.
- **Testing**: Ensure `make test` passes.

### Running Locally

After building, you can run the server locally:

```bash
./build/bin/server run --config-path server/examples/popular_services/wttr.in/config.yaml
```

## 🤝 Contributing

Contributions are welcome! Please feel free to open an issue or submit a pull request.

## 🗺️ Roadmap

Check out our [Roadmap](server/docs/roadmap.md) to see what we're working on and what's coming next.

## 📄 License

This project is licensed under the terms of the [LICENSE](LICENSE) file.
