# Agent Debugger & Inspector

The Agent Debugger is a middleware that monitors and records HTTP traffic for inspection. It allows developers to "replay" traffic and inspect requests and responses to debug agent interactions.

## Configuration

To enable the debugger, configure it in your `config.yaml` (example):

```yaml
debugger:
  enabled: true
  size: 100 # Number of entries to keep in the ring buffer
```

## API

The debugger exposes an endpoint (typically `/debug/entries`) to retrieve the recorded traffic logs.
