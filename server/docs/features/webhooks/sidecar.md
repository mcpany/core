# Webhooks Sidecar

The Webhooks Sidecar pattern allows you to offload webhook processing and context optimization.

## Overview

Run a lightweight sidecar container alongside your main application or MCP server to handle incoming webhooks, perform pre-processing (like token counting or summarization), and then forward the relevant information.

## Benefits

- **Context Optimization**: Summarize large payloads before they hit the context window.
- **Security**: Validate signatures and filter events at the edge.
- **Scalability**: Handle high-volume webhooks without blocking the main server.
