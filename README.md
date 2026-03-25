# MCP Any - The Universal Adapter for APIs

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/mcpany/core)](https://goreportcard.com/report/github.com/mcpany/core)
[![Tests](https://github.com/mcpany/core/actions/workflows/test.yml/badge.svg)](https://github.com/mcpany/core/actions)

## 🚀 Elevator Pitch

**What is MCP Any?**

MCP Any is a Universal Adapter that eliminates the need to write new Model Context Protocol (MCP) servers for every API you want to connect to your AI assistants.

Instead of writing code to wrap a REST API, a gRPC service, or a GraphQL endpoint, you simply write a lightweight YAML or JSON configuration file. MCP Any acts as a bridge, reading your configuration and automatically exposing those APIs as standard MCP tools that any compatible AI (like Claude or ChatGPT) can use.

**Why does it exist?**

Writing a custom MCP server for every single API creates "binary fatigue" and maintenance nightmares. You shouldn't have to be a backend developer just to give your AI access to a new data source. MCP Any democratizes API access by turning configuration into capability. Run one binary, load endless possibilities.

## 🏗️ Architecture

MCP Any is built around a flexible, plugin-style architecture driven entirely by configuration:

*   **Universal Adapter Core**: The main server runs locally and speaks the standard MCP protocol to clients (like Claude Desktop).
*   **Dynamic Tool Discovery**: Tools are discovered and registered automatically from various sources:
    *   OpenAPI (Swagger) Specifications
    *   gRPC Server Reflection or Proto files
    *   GraphQL Introspection
    *   Local CLI Commands
*   **Security & Policy Engine**: Because it connects AI to your real APIs, security is paramount. It includes:
    *   **Strict Egress Policies**: You control exactly which domains and IPs the server can communicate with.
    *   **Context Propagation**: Securely handles passing authentication tokens (API Keys, Bearer, mTLS) to upstream services without ever exposing them to the AI.
    *   **Input Validation & Sanitization**: Blocks dangerous operations and validates inputs before they hit your APIs.
*   **Multi-Tenancy**: Built-in support for multiple users and profiles, allowing tailored access controls per session.

## 🏁 Getting Started

From `git clone` to your first AI-powered API call in minutes.

### 1. Prerequisites

*   [Go 1.21+](https://golang.org/dl/) installed.
*   A compatible MCP Client (like [Claude Desktop](https://claude.ai/download)).

### 2. Installation

Clone the repository and build the binary:

```bash
git clone https://github.com/mcpany/core.git mcpany
cd mcpany
make build # Or `go build -o mcpany ./server/cmd/server`
```

### 3. Hello World Configuration

Create a simple configuration file `weather.yaml` to expose a public API:

```yaml
version: "v1"
profiles:
  - name: default
    services:
      - name: weather_api
        type: openapi
        openapi:
          url: "https://api.weather.gov/openapi.json"
        allowed_hosts:
          - "api.weather.gov"
```

### 4. Run the Server

Start MCP Any, pointing it to your configuration:

```bash
./mcpany --config weather.yaml
```

### 5. Connect your AI

Configure your MCP Client to start the `mcpany` process. The AI will immediately discover the tools exposed by the API and can start making requests!

## 🛠️ Development

We welcome contributions! Here is how to get your development environment set up.

### Running Tests

To run the comprehensive test suite:

```bash
make test
```

For dockerized tests:

```bash
make docker-test
```

### Linting

We enforce strict documentation and code quality standards. Ensure your code passes the linter before submitting a PR:

```bash
make lint
```

For dockerized linting:

```bash
make docker-lint
```

### Building

To compile the binary:

```bash
make build
```

## ⚙️ Configuration

MCP Any is highly configurable via YAML/JSON.

### Environment Variables

*   `MCP_CONFIG`: Path to the configuration file (default: `config.yaml`).
*   `MCP_PORT`: Port to run the server on (if using HTTP transport).
*   `LOG_LEVEL`: Set logging verbosity (`debug`, `info`, `warn`, `error`).

### Secrets Management

**Do not hardcode secrets in your configuration files.**

MCP Any supports loading secrets from environment variables. Use the `${ENV_VAR}` syntax in your YAML:

```yaml
services:
  - name: internal_db
    type: graphql
    url: "https://internal.api/graphql"
    auth:
      type: bearer
      token: "${INTERNAL_API_TOKEN}" # Loaded from environment
```

See the [Configuration Reference](server/docs/reference/configuration.md) for full details on all available options, policies, and service types.
