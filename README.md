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

MCP Any is a **Universal Adapter** that transforms _any_ API (REST, gRPC, GraphQL, Database) into a Model Context Protocol (MCP) compliant server through declarative configuration.

**Why does this exist?**
Traditional MCP adoption requires running a separate binary for every tool (`mcp-server-postgres`, `mcp-server-github`, etc.). This creates "binary fatigue," complex local orchestration, and maintenance overhead.

**MCP Any solves this:**
1.  **Single Binary**: Run one `mcpany` instance.
2.  **Configuration over Code**: Enable capabilities via lightweight YAML/JSON files.
3.  **Unified Ops**: Centralized authentication, caching, rate limiting, and observability.

---

## 🚀 Quick Start

### For Users (Docker)
The fastest way to get started without installing Go or building from source.

```bash
# Run with a sample configuration (Weather Service)
docker run -d --rm --name mcpany-server \
  -p 50050:50050 \
  ghcr.io/mcpany/server:dev-latest \
  run --config-path https://raw.githubusercontent.com/mcpany/core/main/server/examples/popular_services/wttr.in/config.yaml
```

**Connect your Client:**
```bash
# Gemini CLI example
gemini mcp add --transport http --trust mcpany http://localhost:50050
```

### For Developers (Source)
To build and contribute to the project.

```bash
# 1. Clone the repository
git clone https://github.com/mcpany/core.git
cd core

# 2. Install dependencies (requires Go 1.23+ and Make)
make prepare

# 3. Build the binary
make build

# 4. Run the server
./build/bin/server run --config-path server/examples/popular_services/wttr.in/config.yaml
```

---

## 🛠️ Developer Workflow

We use `make` to automate common development tasks.

| Command | Description |
| :--- | :--- |
| `make test` | **Run all tests** (Unit, Integration, E2E). |
| `make lint` | **Run linters** (Go, protobuf) to ensure code quality. |
| `make build` | **Compile the project** to `build/bin/server`. |
| `make gen` | **Regenerate code** from Protocol Buffers definitions. |
| `make clean` | Remove build artifacts and generated files. |

**Prerequisites:**
*   **Go**: Version 1.23+
*   **Docker**: For running tests and building images.
*   **Make**: Standard build tool.

---

## 🏗️ Architecture

**Philosophy: Configuration over Code**

MCP Any acts as a dynamic runtime adapter. Instead of writing code to wrap an API, you define the mapping in configuration.

| Feature | Traditional MCP Server | MCP Any |
| :--- | :--- | :--- |
| **Architecture** | **Code-Driven Wrapper**: Compiled binary per service. | **Config-Driven Adapter**: One runtime, many configs. |
| **Deployment** | **N Binaries**: Complex orchestration. | **1 Binary**: Simple deployment. |
| **Updates** | **Recompile**: Wait for new release. | **Reload**: Update config and hot-swap. |

### System Design
*   **Core Server (`server/`)**: Go-based high-performance runtime.
    *   **Service Registry**: Manages active tools and upstream connections.
    *   **Upstream Adapters**: Generic clients for HTTP, gRPC, SQL, etc.
    *   **Middleware**: Handles Auth, Logging, Rate Limiting.
*   **Management UI (`ui/`)**: Next.js/React dashboard for visualization and config.
*   **Protocol Buffers (`proto/`)**: Defines the internal and external APIs.

### Key Features
*   **Dynamic Reloading**: Hot-swap configurations without downtime.
*   **Auto-Discovery**: Automatically ingest OpenAPI/Swagger and gRPC Reflection.
*   **Security**: Granular Safety Policies, Audit Logging, and Upstream Authentication.
*   **Observability**: Real-time metrics and Network Topology visualization.

For deep dives, see:
*   [Developer Guide](server/docs/developer_guide.md)
*   [Configuration Reference](server/docs/reference/configuration.md)

---

## ⚙️ Configuration

MCP Any is configured via YAML/JSON files.

```bash
# Example: Run with multiple config files
mcpany run --config-path ./config/github.yaml --config-path ./config/postgres.yaml
```

**Environment Variables:**
*   `MCPANY_API_KEY`: Secure the server.
*   `MCPANY_LOG_LEVEL`: `debug`, `info`, `warn`, `error`.
*   `MCPANY_MCP_LISTEN_ADDRESS`: Default `:50050`.

---

## 🤝 Contributing

We welcome contributions! Please ensure you:
1.  Run `make test` before submitting.
2.  Ensure `make lint` passes.
3.  Add documentation for any new public interfaces.

## 📄 License

Apache 2.0 - See [LICENSE](LICENSE) for details.
