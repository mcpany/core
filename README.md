# MCP Any - Universal Agent Infrastructure

## Project Identity
MCP Any is a Universal Adapter designed to eliminate the requirement to implement new MCP servers for doing API calls. It allows you to configure everything through lightweight YAML/JSON configurations to capability-enable different APIs (REST, gRPC, GraphQL, Command-line) and run a single `mcpany` server instance that acts as a secure, universal bridge.

## Quick Start
1. Clone the repository:
   ```bash
   git clone https://github.com/mcpany/core.git
   cd core
   ```
2. Install frontend dependencies:
   ```bash
   cd ui
   npm install
   cd ..
   ```
3. Run the backend application (Ensure `bazelisk` is in your PATH):
   ```bash
   bazelisk run //server/cmd/mcpany -- -config examples/hello_world.yaml
   ```

## Developer Workflow
To set up your development environment and verify code:
- Run tests:
  ```bash
  make test
  ```
- Run linter:
  ```bash
  make lint
  ```
- Build the binary:
  ```bash
  bazelisk build //server/cmd/mcpany
  ```

## Architecture
MCP Any relies on a "Configuration over Code" pattern. Users deploy a single binary which reads dynamically loaded capability definitions. The architecture supports gRPC, OpenAPI, HTTP, GraphQL, and CLI tools. Key architectural features include:
- **Dynamic Tool Registration**: Discovers tools automatically from Proto, OpenAPI specs, or Reflection.
- **Safety Policies**: Pluggable middlewares that block dangerous operations or restrict URL access.
- **Upstream Authentication**: Handles authentication transparently (API keys, mTLS, Bearer tokens).
- **Multi-Tenant**: Supports complex multi-user/multi-profile isolation.
