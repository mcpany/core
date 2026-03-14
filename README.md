# MCP Any - Gold Standard Universal Adapter

[![Documentation: Gold Standard](https://img.shields.io/badge/Documentation-Gold%20Standard-gold.svg)](https://github.com/mcpany/core)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflow/status/mcpany/core/ci.yml?branch=main)](https://github.com/mcpany/core/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/mcpany/core)](https://goreportcard.com/report/github.com/mcpany/core)

## Project Identity

**What is this?**

**MCP Any** is a universal adapter that instantly turns your existing APIs into [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) compliant tools. It acts as a configuration-driven gateway, bridging the gap between your backend services (REST, gRPC, OpenAPI, Command-line) and AI agents.

**Why does it exist?**

Traditional MCP adoption often requires writing a separate server binary for every tool, leading to "binary fatigue" and significant maintenance overhead. MCP Any solves this problem by providing a single, unified server that acts as a gateway to multiple services, defined purely through lightweight configuration files. It unifies your infrastructure into a single, secure, and observable MCP endpoint, allowing you to focus on capabilities rather than plumbing.

## Quick Start

Follow these exact commands to clone, install dependencies, and run the app.

### Prerequisites

*   [Go 1.23+](https://go.dev/doc/install) (for building from source)
*   `make` (for build automation)
*   [Docker](https://docs.docker.com/get-docker/) (optional, for containerized run)

### One-Shot Setup

```bash
git clone https://github.com/mcpany/core.git
cd core
make prepare
make build
./build/bin/server run --config-path server/config.minimal.yaml
```

Once the server is running, you can verify its health:
```bash
curl http://localhost:50050/health
```

## Developer Workflow

We adhere to a strict development workflow to ensure code quality and maintainability.

### 1. Building the Code
Compile the server binary and UI assets.
```bash
make build
```

### 2. Testing
Run all unit and integration tests to ensure code correctness. We practice proactive testing and continuous integration.
```bash
make test
```

### 3. Linting and Documentation
We enforce **100% documentation coverage** and strict style guides. Run linters to ensure compliance with our gold standards:
```bash
make lint
```

### 4. Code Generation
Regenerate Protocol Buffers and other auto-generated files if you modify `.proto` definitions.
```bash
make gen
```

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

**Request Flow:**

1.  **Client Request:** An AI agent (e.g., Claude) sends a JSON-RPC request (e.g., `tools/call`) to the MCP Any Core Server.
2.  **Authentication:** The server verifies the request's API Key or Session Token.
3.  **Policy Check:** The Policy Engine evaluates the request against active Profiles and DLP rules. Blocked requests are rejected immediately.
4.  **Routing:** The Service Registry resolves the requested tool to a specific Upstream Adapter.
5.  **Adaptation:** The Upstream Adapter transforms the MCP request into the target protocol (e.g., constructs an HTTP request or gRPC message).
6.  **Execution:** The adapter communicates with the upstream service.
7.  **Response Transformation:** The upstream response is received, transformed back into MCP format (e.g., `CallToolResult`), and returned to the client.

## Contributing

We welcome contributions! Please read [AGENTS.md](server/AGENTS.md) for our coding standards, documentation requirements, and development workflow.

## License

This project is licensed under the terms of the [Apache 2.0 License](LICENSE).
