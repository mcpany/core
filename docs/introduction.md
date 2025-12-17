# 🚀 Introduction

Welcome to MCP Any! This document provides a high-level overview of the project, its purpose, and its core components.

## The Goal: Elimination of Custom MCP Servers

The primary goal of **MCP Any** is to eliminate the requirement to implement, compile, and maintain new MCP servers for every API call you want to expose.

Traditionally, if you wanted to expose an API as an MCP tool, you would need to write a specific server wrapper for it. With MCP Any, you can achieve this using **pure configuration**.

## ❓ Philosophy: Configuration over Code

We believe you shouldn't have to write and maintain new code just to expose an existing API to your AI assistant.

- **Metamcp / Onemcp vs. MCP Any**: While other tools might proxy existing MCP servers (aggregator pattern), **MCP Any** creates them from scratch using your existing upstream APIs.
- **No More "Sidecar hell"**: Instead of running 10 different containers for 10 different tools, run 1 `mcpany` container loaded with 10 config files.
- **Ops Friendly**: Centralize authentication, rate limiting, and observability in one robust layer.

### Comparison with Traditional MCP Servers

Unlike traditional "Wrapper" MCP servers (like `mcp-server-postgres`, `mcp-server-github`, etc.) which are compiled binaries dedicated to a single service, **MCP Any** is a generic runtime.

| Feature           | Traditional MCP Server (e.g., `mcp-server-postgres`)                    | MCP Any                                                                         |
| :---------------- | :---------------------------------------------------------------------- | :------------------------------------------------------------------------------ |
| **Architecture**  | **Code-Driven Wrapper**: Wraps internal API calls with MCP annotations. | **Config-Driven Adapter**: Maps existing API endpoints to MCP tools via config. |
| **Deployment**    | **1 Binary per Service**: Need 10 different binaries for 10 services.   | **1 Binary for All**: One `mcpany` binary handles N services.                   |
| **Updates**       | **Recompile & Redistribute**: Internal API change = New Binary release. | **Update Config**: API change = Edit YAML/JSON file & reload.                   |
| **Maintenance**   | **High**: Manage dependencies/versions for N projects.                  | **Low**: Upgrade one core server; just swap config files.                       |
| **Extensibility** | Write code (TypeScript/Python/Go).                                      | Write JSON/YAML.                                                                |

Most "popular" MCP servers today are bespoke binaries. If the upstream API changes, you must wait for the maintainer to update the code, release a new version, and then you must redeploy. With **MCP Any**, you simply update your configuration file to match the new API signature—zero downtime, zero recompilation.

## ✨ Key Features

- **Dynamic Tool Registration & Auto-Discovery**: Automatically discover and register tools from various backend services. For gRPC and OpenAPI, simply provide the server URL or spec URL—MCP Any handles the rest (no manual tool definition required).
- **Multiple Service Types**: Supports a wide range of service types, including:
  - **gRPC**: Register services from `.proto` files or by using gRPC reflection.
  - **OpenAPI**: Ingest OpenAPI (Swagger) specifications to expose RESTful APIs as tools.
  - **HTTP**: Expose any HTTP endpoint as a tool.
  - **GraphQL**: Expose a GraphQL API as a set of tools, with the ability to customize the selection set for each query.
- **Advanced Service & Safety Policies**:
  - **Safety**: Control which tools are exposed to the AI to limit context (reduce hallucinations) and prevent dangerous actions (e.g., blocking `DELETE` operations).
  - **Performance**: Configure [Caching](features/caching/README.md) and Rate Limiting to optimize performance and protect upstream services.
- **MCP Any Proxy**: Proxy and re-expose tools from another MCP Any instance.
- **Upstream Authentication**: Securely connect to your backend services using:
  - **API Keys**
  - **Bearer Tokens**
  - **Basic Auth**
  - **mTLS**
- **Unified API**: Interact with all registered tools through a single, consistent API based on the [Model Context Protocol](https://modelcontext.protocol.ai/).
- **Multi-User & Multi-Profile**: Securely support multiple users with distinct profiles, each with its own set of enabled services and granular authentication.
- **Extensible**: Designed to be easily extended with new service types and capabilities.
