# MCP Any - Universal Agent Infrastructure

## Project Identity
**What is this?** MCP Any is a Universal Adapter designed to eliminate the requirement to implement new MCP servers for doing API calls.
**Why does it exist?** It allows you to configure everything through lightweight YAML/JSON configurations to capability-enable different APIs (REST, gRPC, GraphQL, Command-line) and run a single `mcpany` server instance that acts as a secure, universal bridge.

## Quick Start
1. **Clone the repository:**
   ```bash
   git clone https://github.com/mcpany/core.git
   cd core
   ```
2. **Install Dependencies:**
   Ensure `bazelisk` and `make` are in your PATH.
3. **Run the App:**
   ```bash
   bazelisk run //server/cmd/mcpany -- -config examples/hello_world.yaml
   ```

## Developer Workflow
To set up your development environment, verify code, and build:
- **Lint the code:**
  ```bash
  make lint
  ```
- **Run the tests:**
  ```bash
  make test
  ```
- **Build the binary:**
  ```bash
  bazelisk build //server/cmd/mcpany
  ```

## Architecture
MCP Any relies on a "Configuration over Code" pattern. Users deploy a single binary which reads dynamically loaded capability definitions. The architecture supports gRPC, OpenAPI, HTTP, GraphQL, and CLI tools. Key architectural features include:
- **Dynamic Tool Registration**: Discovers tools automatically from Proto, OpenAPI specs, or Reflection.
- **Safety Policies**: Pluggable middlewares that block dangerous operations or restrict URL access.
- **Upstream Authentication**: Handles authentication transparently (API keys, mTLS, Bearer tokens).
- **Multi-Tenant**: Supports complex multi-user/multi-profile isolation.

## Configuration
MCP Any requires configurations to be provided via YAML/JSON.
- **Environment Variables:** Set any secret values in environment variables (e.g., `OPENAI_API_KEY`) and reference them in your config files.
- **Configs:** Place capability configurations in the `./configs` directory. Required secrets must be resolved dynamically to ensure security boundaries are maintained.
