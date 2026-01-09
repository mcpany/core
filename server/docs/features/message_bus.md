# Message Bus Integration

MCP Any uses a message bus for asynchronous communication and decoupled architecture.

## Supported Brokers

- **[NATS](nats.md)**: Lightweight and high-performance.
- **[Kafka](kafka.md)**: Robust and scalable for high-throughput.

## Use Cases

- **Audit Logging**: Asynchronously ship audit logs to external systems.
- **Events**: Publish system events (e.g., tool execution, errors).
- **Decoupling**: Separate the core server from heavy processing tasks.
