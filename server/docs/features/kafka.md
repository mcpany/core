# Kafka Integration

MCP Any supports using **Kafka** as a message bus for internal communication and event handling. This allows for high-throughput event streaming and integration with existing Kafka infrastructure.

## Configuration

To use Kafka as the message bus, configure the `message_bus` section in your `global_settings`.

```yaml
global_settings:
  message_bus:
    kafka:
      brokers:
        - "localhost:9092"
      topic_prefix: "mcpany-"
      consumer_group: "mcpany-group" # Optional: set for Queue semantics, unset for Broadcast
```

## Features

*   **Pub/Sub Messaging**: Used for internal events and communication.
*   **Scalability**: Allows multiple `mcpany` instances to communicate.
*   **Durability**: Messages are persisted in Kafka.

## Consumer Groups

*   If `consumer_group` is specified, `mcpany` instances will share the load (Queue semantics) for subscriptions that share the same group ID.
*   If `consumer_group` is not specified (default), each `mcpany` instance will generate a unique consumer group ID, resulting in Broadcast semantics (all instances receive all messages).
