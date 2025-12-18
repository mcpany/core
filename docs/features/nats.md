# NATS Integration

MCP Any supports using **NATS** as a message bus for internal communication and event handling. This allows for decoupling components and building reactive architectures.

## Configuration

To use NATS as the message bus, configure the `message_bus` section in your `global_settings`.

```yaml
global_settings:
  message_bus:
    nats:
      server_url: "nats://localhost:4222"
```

## Features

*   **Pub/Sub Messaging**: Used for internal events and communication.
*   **Scalability**: Allows multiple `mcpany` instances to communicate (e.g., for future distributed features).

## Future Capabilities

*   **Transport Protocol**: In the future, NATS may be used as a primary transport protocol for MCP communication, allowing clients to connect via NATS subjects instead of HTTP/gRPC.
