# MCP Any - Gold Standard Universal Adapter

[![Documentation: Gold Standard](https://img.shields.io/badge/Documentation-Gold%20Standard-gold.svg)](https://github.com/mcpany/core)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## Elevator Pitch
**What is this project and why does it exist?**

**MCP Any** is a universal adapter that instantly turns your existing APIs into [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) compliant tools. Traditional MCP adoption often requires writing a separate server binary for every tool, leading to "binary fatigue" and significant maintenance overhead. MCP Any solves this problem by providing a single, unified server that acts as a gateway to multiple services, defined purely through lightweight configuration files. It unifies your infrastructure into a single, secure, and observable MCP endpoint, allowing you to focus on capabilities rather than plumbing.

## Architecture
**High-Level Overview**

MCP Any utilizes a modular, adapter-based architecture to decouple the MCP protocol from upstream API specifics. Built with Go for performance and concurrency, it serves as a robust middleware between AI clients and your infrastructure.

**Core Components:**
1.  **Core Server**: A high-performance Go runtime that handles the MCP protocol (JSON-RPC) and manages client sessions.
2.  **Service Registry**: The central nervous system of MCP Any. It implements the `ServiceRegistryInterface` to manage the lifecycle of upstream services. It handles dynamic loading, hot-reloading, and health checking of services defined in configuration.
3.  **Upstream Adapters**: Specialized implementations of the `Upstream` interface that translate MCP requests into protocol-specific calls:
    *   **HTTP**: Proxies requests to REST/JSON APIs with powerful parameter mapping and transformation templates.
    *   **gRPC**: Uses reflection to dynamically discover and invoke methods on gRPC services without generating code.
    *   **Command**: Safely executes local CLI tools or scripts in a controlled environment.
    *   **Filesystem**: Provides secure access to local or remote (S3, GCS) filesystems.
4.  **Policy Engine & Middleware**: A security layer that enforces authentication, rate limiting, DLP (Data Loss Prevention), and audit logging.

## Getting Started
Follow these steps to get up and running with MCP Any immediately.

```bash
# 1. Clone the repository
git clone https://github.com/mcpany/core.git
cd core

# 2. Install dependencies
make prepare

# 3. Build the binary
make build

# 4. Run the server (Hello World)
./build/bin/server run --config-path server/config.minimal.yaml
```

Once the server is running, you can connect an AI client (like Claude Desktop or Gemini CLI) to `http://localhost:50050` and try out the available tools, such as asking for the weather.

## Development
We adhere to a strict development workflow to ensure code quality and maintainability.

*   **Testing:** Run all unit and integration tests to ensure code correctness. We practice proactive testing and continuous integration.
    ```bash
    make test
    ```
*   **Linting:** We enforce 100% documentation coverage and strict style guides using `make lint`.
    ```bash
    make lint
    ```
*   **Building:** Compile the server binary and UI assets.
    ```bash
    make build
    ```

## Configuration
MCP Any is configured via environment variables and YAML/JSON configuration files. This allows for flexible deployment across different environments.

### Environment Variables
| Variable | Description | Default |
|----------|-------------|---------|
| `MCPANY_MCP_LISTEN_ADDRESS` | MCP server's bind address (host:port) | `50050` |
| `MCPANY_CONFIG_PATH` | Comma-separated paths to config files/dirs | `[]` |
| `MCPANY_API_KEY` | Master API key for securing the server | Empty (No Auth) |

### Required Secrets
Sensitive information (like upstream API keys) must **never** be hardcoded in configuration files. Instead, use environment variable references.
```yaml
upstreamAuth:
  apiKey:
    value: "${OPENAI_API_KEY}" # References env var
```
Ensure `OPENAI_API_KEY` (or your specific secret) is set in the server's environment before starting.
