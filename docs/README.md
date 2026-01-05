# MCP Any Documentation

Welcome to the official documentation for MCP Any. This documentation provides everything you need to understand, use, and extend `mcpany`.

Whether you are a developer looking to integrate your services, an operator deploying `mcpany` in a production environment, or a contributor wanting to improve the project, this is the place to start.

## 🚀 Getting Started

If you are new to `mcpany`, we recommend starting here.

- **[Introduction](introduction.md)**: What is `mcpany` and why should you use it?

## 🧠 Core Concepts

Understand the foundational principles and design of `mcpany`.

- **[Architecture](architecture.md)**: A deep dive into the internal architecture of the `mcpany` server.

## ✨ Features

- **[Authentication](features/authentication/README.md)**: Secure your server and upstream connections.
- **[Caching](features/caching/README.md)**: Improve performance with response caching.
- **[Webhooks](features/webhooks/README.md)**: Intercept and modify tool executions.
- **[Connection Pooling](features/connection-pooling/README.md)**: Manage upstream connections efficiently.
- **[Rate Limiting](features/rate-limiting/README.md)**: Protect upstream services.
- **[Resilience](features/resilience/README.md)**: Circuit breakers and retries.
- **[Monitoring](features/monitoring/README.md)**: Metrics and observability.
- **[Profiles](features/profiles_and_policies/README.md)**: Environment-specific configurations.
- **[Prompts](features/prompts/README.md)**: Reusable prompts for AI interactions.

## 🔌 Integrations

Learn how to connect `mcpany` with your favorite AI assistants and other tools.

- **[Integration Guide](integrations.md)**: Detailed instructions for integrating `mcpany` with clients like the Gemini CLI, Claude, and more.

## 💡 Examples

Explore practical examples of how to use `mcpany` with different types of backend services.

- **[Examples Overview](examples.md)**: A guide to the included examples, from simple HTTP servers to complex gRPC services.

## 💻 Development

For those who want to contribute to the `mcpany` project.

- **[Developer Guide](developer_guide.md)**: Instructions on how to set up your development environment, run tests, and contribute code.

## 📚 Reference

Detailed technical reference documentation.

- **[Configuration Reference](reference/configuration.md)**: A complete reference for all configuration options.
- **[API Reference](reference/api.md)**: A full reference for the `mcpany` JSON-RPC API.
- **[Command-Line Reference](reference/cli.md)**: A guide to the `mcpany` command-line interface.
- **[Service Types](reference/service_types.md)**: A detailed explanation of the supported service types.
