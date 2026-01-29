[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Test](https://github.com/mcpany/core/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/mcpany/core/actions/workflows/ci.yml)
[![GoDoc](https://godoc.org/github.com/mcpany/core?status.png)](https://pkg.go.dev/github.com/mcpany/core)
[![GoReportCard](https://goreportcard.com/badge/github.com/mcpany/core)](https://goreportcard.com/report/github.com/mcpany/core)
[![codecov](https://codecov.io/gh/mcpany/core/branch/main/graph/badge.svg)](https://codecov.io/gh/mcpany/core)

<p align="center">
  <img src="server/docs/images/logo.png" alt="MCP Any Logo" width="200"/>
</p>

# MCP Any: The Universal MCP Adapter

**One server, Infinite possibilities.**

## 1. Project Identity

**What is this?**
MCP Any is a configuration-driven **Universal Adapter** that turns *any* API (REST, gRPC, GraphQL, Command-line) into a Model Context Protocol (MCP) compliant server.

**Why does it exist?**
Traditional MCP adoption suffers from "binary fatigue"—requiring a separate server binary for every tool. MCP Any solves this by allowing you to run a single binary that acts as a gateway to multiple services, defined purely through lightweight configuration files.

**The Elevator Pitch:**
Don't write code to expose your APIs to AI agents. Just configure them. MCP Any unifies your backend services into a single, secure, and observable MCP endpoint.

## 2. Quick Start

Get up and running with a weather service example in minutes.

### Prerequisites
*   [Go 1.23+](https://go.dev/doc/install)
*   [Docker](https://docs.docker.com/get-docker/) (optional, for containerized run)

### Commands

**Option 1: Docker (Fastest)**

```bash
docker run -d --rm --name mcpany-server \
  -p 50050:50050 \
  ghcr.io/mcpany/server:dev-latest \
  run --config-path https://raw.githubusercontent.com/mcpany/core/main/server/examples/popular_services/wttr.in/config.yaml
```

**Option 2: Build from Source**

```bash
# 1. Clone the repository
git clone https://github.com/mcpany/core.git
cd core

# 2. Install dependencies and build
make prepare
make build

# 3. Run the server (using the example weather config)
./build/bin/server run --config-path server/examples/popular_services/wttr.in/config.yaml
```

**Connect your Client:**
Once running, connect your MCP client (like Gemini CLI or Claude Desktop) to `http://localhost:50050`.

```bash
gemini mcp add --transport http --trust mcpany http://localhost:50050
```

## 3. Developer Workflow

For contributors and developers extending the core platform.

**Run Tests**
Execute all unit, integration, and end-to-end tests.
```bash
make test
```

**Lint Code**
Ensure code quality and style compliance (Go & TypeScript).
```bash
make lint
```

**Build Artifacts**
Compile the server binary and frontend assets.
```bash
make build
```

**Generate Code**
Regenerate Protocol Buffers and other auto-generated files.
```bash
make gen
```

## 4. Architecture

MCP Any acts as a centralized middleware between AI Agents (Clients) and your Upstream Services.

**High-Level Summary:**
1.  **Core Server**: A Go-based runtime that speaks the MCP protocol.
2.  **Service Registry**: Dynamically loads tool definitions from configuration.
3.  **Adapters**: Specialized modules that translate MCP requests into upstream calls (gRPC, HTTP, OpenAPI, etc.).
4.  **Policy Engine**: Enforces authentication, rate limiting, and security policies.

```mermaid
graph TD
    User[User / AI Agent] -->|MCP Protocol| Server[MCP Any Server]

    subgraph "MCP Any Core"
        Server --> Registry[Service Registry]
        Registry -->|Config| Config[Configuration]
        Registry -->|Policy| Auth[Authentication & Policy]
    end

    subgraph "Upstream Services"
        Registry -->|gRPC| ServiceA[gRPC Service]
        Registry -->|HTTP| ServiceB[REST API]
        Registry -->|OpenAPI| ServiceC[OpenAPI Spec]
        Registry -->|CMD| ServiceD[Local Command]
    end
```

### Key Features
*   **Dynamic Config Reloading**: Hot-swap registry without restarts.
*   **Broad Protocol Support**: gRPC, OpenAPI, HTTP, GraphQL, SQL, WebSocket, WebRTC.
*   **Safety Policies**: Block dangerous operations (e.g., DELETE) and limit access.
*   **Network Topology**: Visual graph of clients, services, and tools.
*   **Observability**: Real-time metrics and audit logging.
*   **Security**: Upstream authentication (API Keys, OAuth, mTLS) and multi-user profiles.

### Troubleshooting
*   **Protobuf Errors**: Run `make prepare` to install dependencies.
*   **Docker Issues**: Ensure user is in the `docker` group or use `sudo`.
*   **Port Conflicts**: Check ports 50050, 50051, and 9002.

### Key Documentation
*   **[Developer Guide](server/docs/developer_guide.md)**: Detailed internal architecture and contribution guide.
*   **[Configuration Reference](server/docs/reference/configuration.md)**: Full syntax for defining services.
*   **[Integrations](server/docs/integrations.md)**: How to use with Claude, Cursor, VS Code, etc.
*   **[Examples](server/docs/examples.md)**: Hands-on examples.

---

<p align="center">
  Generated with ❤️ by the MCP Any Team.
</p>
