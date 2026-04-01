# MCP Any

## Project Identity
**What is this?** MCP Any is the Universal Adapter for AI agents.
**Why does it exist?** It empowers developers to build secure, capability-enabled APIs (REST, gRPC, GraphQL, Command-line) through lightweight YAML/JSON configurations, serving as an ultimate universal bridge and eliminating the need to write custom boilerplate adapters.

## Quick Start
```bash
git clone https://github.com/mcpany/core.git
cd core
# Install dependencies
# Ensure `bazelisk` is installed and available in your `PATH`.
bazelisk run //server/cmd/mcpany -- -config examples/hello_world.yaml
```

## Developer Workflow
```bash
# How to run `make test`
make test

# How to run `make lint`
make lint

# How to build builds
bazelisk build //server/cmd/mcpany
```

## Architecture
MCP Any relies on a "Configuration over Code" pattern. Users deploy a single binary which reads dynamically loaded capability definitions. The architecture supports gRPC, OpenAPI, HTTP, GraphQL, and CLI tools.
Features include:
- Dynamic Tool Registration.
- Safety Policies.
- Upstream Authentication.
- Multi-Tenant isolation.

### High-Level Flow
```mermaid
graph TD
    classDef agent fill:#f9f,stroke:#333,stroke-width:2px;
    classDef core fill:#bbf,stroke:#333,stroke-width:2px;
    classDef target fill:#bfb,stroke:#333,stroke-width:2px;

    A[Client Application / Agent Framework]:::agent -->|Model Context Protocol| B(MCP Any Adapter Hub):::core

    subgraph Config [Capability Configurations]
        B --> C{Dynamic Router}
    end

    C -->|REST/OpenAPI Specs| D[REST/HTTP API]:::target
    C -->|Protobuf/Reflection| E[gRPC Service]:::target
    C -->|GraphQL Schema| F[GraphQL Endpoint]:::target
    C -->|Standard I/O| G[Command Line Tools]:::target
```
