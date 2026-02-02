# MCP Any: The Universal MCP Adapter

**One server, Infinite possibilities.**

## 1. Project Identity

**What is this?**
**MCP Any** is a universal adapter that instantly turns existing APIs into [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) compliant tools. It serves as a configuration-driven gateway, bridging the gap between your AI agents and any backend service (REST, gRPC, GraphQL, or CLI) without requiring custom code for each integration.

**Why does it exist?**
Traditional MCP adoption suffers from "binary fatigue"—developers must build and maintain a separate server binary for every tool they wish to expose. **MCP Any** eliminates this burden by providing a single, high-performance runtime that:
1.  **Unifies** multiple upstream services into one secure MCP endpoint.
2.  **Decouples** protocol details from agent logic via the Adapter Pattern.
3.  **Enforces** security policies (authentication, rate limiting, DLP) centrally.

**The Solution:** Don't write code to expose your APIs. Just configure them.

## 2. Quick Start

Get up and running with a functional MCP server in under 5 minutes.

### Prerequisites
*   [Go 1.23+](https://go.dev/doc/install)
*   [Docker](https://docs.docker.com/get-docker/) (optional, for containerized execution)
*   `make`

### Installation & Run

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/mcpany/core.git
    cd core
    ```

2.  **Install dependencies:**
    ```bash
    make prepare
    ```

3.  **Build the server:**
    ```bash
    make build
    ```
    *Output binary located at `build/bin/server`.*

4.  **Run with example configuration:**
    ```bash
    ./build/bin/server run --config-path server/examples/popular_services/wttr.in/config.yaml
    ```

### Verification
Once running, verify the server is healthy:
```bash
curl http://localhost:50050/health
```

Connect your AI client (e.g., Claude Desktop, Gemini CLI):
```bash
gemini mcp add --transport http --trust mcpany http://localhost:50050
```

## 3. Developer Workflow

We adhere to strict engineering standards. Follow this workflow to contribute.

### Testing
Run the comprehensive test suite (Unit, Integration, E2E):
```bash
make test
```

### Linting & Compliance
We enforce **100% documentation coverage** for public interfaces and strict style guides.
```bash
make lint
```
*Note: This runs `check-go-doc` and `check-ts-doc` to ensure all exported symbols are documented.*

### Building
Compile the server binary and UI assets:
```bash
make build
```

### Code Generation
Regenerate Protocol Buffers and mocks after modifying definitions:
```bash
make gen
```

## 4. Architecture

MCP Any operates as a centralized middleware layer. It allows you to define "Upstream Services" in YAML/JSON, which are then projected as MCP Tools, Resources, and Prompts.

**High-Level Design:**

1.  **Core Server (Go):** High-concurrency runtime handling the MCP protocol lifecycle.
2.  **Service Registry:** Dynamically loads and manages tool definitions from config.
3.  **Adapters:** Protocol-specific modules that translate MCP requests:
    *   `http`: For REST/JSON APIs.
    *   `grpc`: For gRPC microservices.
    *   `command`: For local CLI tools (sandboxed).
4.  **Policy Engine:** Intercepts calls to enforce RBAC, input validation, and redaction.

```mermaid
graph TD
    User[User / AI Agent] -->|MCP Protocol| Server[MCP Any Server]

    subgraph "MCP Any Core"
        Server --> Registry[Service Registry]
        Registry -->|Config| Config[Configuration Files]
        Registry -->|Policy| Auth[Policy Engine]
    end

    subgraph "Upstream Services"
        Registry -->|gRPC| ServiceA[gRPC Service]
        Registry -->|HTTP| ServiceB[REST API]
        Registry -->|OpenAPI| ServiceC[OpenAPI Spec]
        Registry -->|CMD| ServiceD[Local Command]
    end
```

### Key Design Patterns
*   **Adapter Pattern:** Translates diverse upstream protocols into a uniform MCP interface.
*   **Configuration as Code:** All behavior is defined in declarative configuration files, enabling version-controlled infrastructure.
*   **Sidecar/Gateway:** Deploys easily as a Kubernetes sidecar or a standalone gateway.

## License
Apache 2.0 - See [LICENSE](LICENSE) for details.
