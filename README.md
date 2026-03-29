# MCP Any - Universal Agent Infrastructure

## Project Identity
**What is this?** MCP Any is the ultimate developer entry point and Universal Adapter designed to eliminate the requirement to implement new MCP (Model Context Protocol) servers for doing API calls.

**Why does it exist?** It allows you to configure everything through lightweight YAML/JSON configurations to capability-enable different APIs (REST, gRPC, GraphQL, Command-line) and run a single `mcpany` server instance that acts as a secure, universal bridge. Instead of writing custom boilerplate adapters for each service, you configure MCP Any to handle it automatically.

## Quick Start

Follow these steps to go from `git clone` to `Hello World`.

### 1. Clone the repository
```bash
git clone https://github.com/mcpany/core.git
cd core
```

### 2. Install Dependencies
Ensure `bazelisk` and `make` are installed and available in your `PATH`.

### 3. Run the App
Run the server with the included Hello World example configuration:
```bash
bazelisk run //server/cmd/mcpany -- -config examples/hello_world.yaml
```
*The server will start and you can connect your MCP-compatible client to the local instance.*

## Developer Workflow

We use `make` and `bazelisk` for common development tasks.

- **Lint the code:**
  Run the linter to ensure your code matches style guidelines:
  ```bash
  make lint
  ```

- **Run the tests:**
  Execute all tests using Bazel:
  ```bash
  bazelisk test //...
  ```
  Or via the Makefile:
  ```bash
  make test
  ```

- **Build the binary:**
  Build the server executable:
  ```bash
  bazelisk build //server/cmd/mcpany
  ```

## Architecture
MCP Any relies on a "Configuration over Code" pattern. Users deploy a single binary which reads dynamically loaded capability definitions. The architecture supports gRPC, OpenAPI, HTTP, GraphQL, and CLI tools.

Key architectural features include:
- **Dynamic Tool Registration**: Discovers tools automatically from Proto, OpenAPI specs, or Reflection.
- **Safety Policies**: Pluggable middlewares that block dangerous operations or restrict URL access (e.g. Audit, DLP, Rate Limiting).
- **Upstream Authentication**: Handles authentication transparently to connected capabilities (API keys, mTLS, Bearer tokens).
- **Multi-Tenant**: Supports complex multi-user/multi-profile isolation.

### High-Level Flow
```mermaid
graph TD
    A[Client Application (LLM/Agent)] --> B(MCP Any Adapter)
    B --> C{Capability Configuration}
    C -->|REST/OpenAPI| D[REST API]
    C -->|gRPC| E[gRPC Service]
    C -->|GraphQL| F[GraphQL Endpoint]
    C -->|CLI| G[Command Line Tool]
```

## Configuration

MCP Any requires configurations to be provided via YAML/JSON.

- **Configs:** Place capability configurations in the `./configs` directory. These define what upstream capabilities your MCP Any server exposes.
- **Environment Variables:** Set any secret values in environment variables (e.g., `OPENAI_API_KEY`) and reference them dynamically in your config files to ensure security boundaries are maintained.
- **Global Settings:** Important settings (like allowed IPs, logging level, rate limits) can be provided through a global configuration structure or CLI flags.
