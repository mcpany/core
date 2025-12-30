# Message Bus Integration

MCP Any supports integration with message brokers to enable asynchronous communication and event-driven architectures. This allows MCP Any to act as a bridge between AI agents and your messaging infrastructure.

## Supported Brokers

- [**NATS**](nats.md): Lightweight, high-performance messaging system.
- [**Kafka**](kafka.md): Distributed event streaming platform.

## Common Use Cases

1.  **Event Triggers**: Agents can "subscribe" to topics and receive notifications when messages arrive (via Tool calls or future "Server-Sent Events").
2.  **Asynchronous Actions**: Agents can publish messages to a topic to trigger background jobs (e.g., "Process this order").
3.  **Log Streaming**: Stream audit logs or operational events to a central bus for analysis.

## Configuration

Message bus configuration is typically handled at the service level or global settings depending on the use case.

See the specific documentation for [NATS](nats.md) and [Kafka](kafka.md) for detailed configuration examples.
