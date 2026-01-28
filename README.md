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

## 🚀 Elevator Pitch

MCP Any revolutionizes how you interact with the Model Context Protocol (MCP). It is not just another MCP proxy or aggregator—it is a powerful **Universal Adapter** that turns _any_ API into an MCP-compliant server through simple configuration.

Traditional MCP adoption requires running a separate server binary for every tool or service you want to expose. This leads to "binary fatigue," complex local setups, and maintenance nightmares.

**MCP Any solves this with a Single Binary approach:**

1.  **Install once**: Run a single `mcpany` server instance.
2.  **Configure everything**: Load lightweight YAML/JSON configurations to capability-enable different APIs (REST, gRPC, GraphQL, Command-line).
3.  **Run anywhere**: No need for `npx`, `python`, or language-specific runtimes for each tool.

## 🏗️ Architecture & Design Patterns

MCP Any is built on a **Configuration over Code** philosophy. We believe you shouldn't have to write and maintain new code just to expose an existing API to your AI assistant.

### Tech Stack
- **Server**: Written in **Go** for high performance, concurrency, and single-binary deployment.
- **UI**: built with **Next.js** and **React** for a modern, responsive management dashboard.
- **Communication**: Uses **gRPC** for internal service communication and **JSON-RPC** for MCP compliance.
- **Storage**: **SQLite** (embedded) or **Postgres** for persistence of audit logs and cache.

### Design Patterns
- **Universal Adapter**: Instead of wrapping internal API calls with code (Wrapper pattern), MCP Any maps existing API endpoints to MCP tools via configuration (Adapter pattern).
- **Hexagonal Architecture**: Core logic is isolated from upstream adapters (HTTP, gRPC, SQL), allowing easy extension.
- **Centralized Governance**: Authentication, rate limiting, and observability are handled centrally, avoiding "Sidecar hell".

| Feature | Traditional MCP Server | MCP Any |
| :--- | :--- | :--- |
| **Approach** | **Code-Driven Wrapper** | **Config-Driven Adapter** |
| **Deployment** | 1 Binary per Service | 1 Binary for All |
| **Updates** | Recompile & Redistribute | Update Config & Reload |
| **Maintenance** | High (N dependencies) | Low (1 Core Server) |

## ✨ Key Features

- **Dynamic Config Reloading**: Hot-swap registry without restarting.
- **Auto-Discovery**: Automatically register tools from OpenAPI, gRPC reflection, or GraphQL schemas.
- **Broad Protocol Support**: REST (OpenAPI), gRPC, GraphQL, SQL, WebSocket, WebRTC.
- **Safety & Security**: Granular access control, block dangerous operations (e.g., DELETE), redacting sensitive data.
- **Observability**: Built-in audit logging, semantic caching, and network topology visualization.
- **Management Dashboard**: A comprehensive UI to manage services, view metrics, and test tools.

## 🚀 Getting Started

Ready to give your AI access to real-time data? Follow these steps to go from `git clone` to "Hello World".

### 1. Prerequisites

- **Go**: Version 1.23+ installed.
- **Docker**: For running tests and building images.
- **Make**: For build automation.

### 2. Installation

Clone the repository and build the server:

```bash
# 1. Clone the repository
git clone https://github.com/mcpany/core.git
cd core

# 2. Prepare dependencies (protoc, linters, plugins)
make prepare

# 3. Build the server
make build
```

The binary will be available at `build/bin/server`.

### 3. Run "Hello World" (Weather Service)

We'll use the pre-built `wttr.in` configuration to check the weather.

```bash
# Run the server with the example configuration
./build/bin/server run --config-path server/examples/popular_services/wttr.in/config.yaml
```

The server is now running on `http://localhost:50050`.

### 4. Verify

You can test it using the `mcp-cli` or `gemini` CLI. Or simply use `curl` to verify the server is up:

```bash
curl http://localhost:50050/sse
```

To use it with **Gemini CLI**:

```bash
gemini mcp add --transport http --trust mcpany http://localhost:50050
gemini -m gemini-2.5-flash -p "What is the weather in London?"
```

## 🛠️ Development

We welcome contributions! Here is how to work on the codebase.

### Running Tests
Run the full test suite (unit, integration, E2E):

```bash
make test
```

### Linting
Ensure your code meets quality standards:

```bash
make lint
```

### Building
Compile the project:

```bash
make build
```

### Documentation Check
Verify that your code is fully documented:

```bash
# For Go
go run server/tools/check_doc.go server/

# For TypeScript
python3 server/tools/check_ts_doc.py
```

## ⚙️ Configuration

MCP Any is configured via environment variables and YAML/JSON files.

### Environment Variables

| Variable | Description | Default |
| :--- | :--- | :--- |
| `MCPANY_MCP_LISTEN_ADDRESS` | Address for the MCP HTTP server. | `:50050` |
| `MCPANY_API_KEY` | **Secret**: Master API key for securing the server. | `""` |
| `MCPANY_CONFIG_PATH` | Comma-separated paths to config files. | `""` |
| `MCPANY_LOG_LEVEL` | Log verbosity (`debug`, `info`). | `info` |
| `MCPANY_DB_PATH` | Path to the SQLite database. | `data/mcpany.db` |

For a complete reference, see [server/docs/reference/configuration.md](server/docs/reference/configuration.md).

## 🖥️ Management Dashboard

The UI provides real-time metrics, service management, and an interactive playground.

1.  **Navigate to UI**: `cd ui`
2.  **Install**: `npm install`
3.  **Run**: `npm run dev` (Access at `http://localhost:9002`)

## 🤝 Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) (if available) and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## 📄 License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.
