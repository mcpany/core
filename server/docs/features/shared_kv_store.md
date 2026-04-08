# Shared Key-Value Store (Blackboard)

The embedded SQLite "Blackboard" acts as a shared Key-Value store, providing reliable state management for multi-agent systems and preventing multi-agent hallucinations.

## Configuration

The Shared KV Store can be configured via:

```yaml
blackboard:
  db_path: "/var/lib/mcpany/blackboard.db"
```

## How it works

The Blackboard serves as a central repository for shared variables and intermediate results. When an agent updates a value, it becomes instantly available to other agents within the swarm or intent-scope. By enforcing agent-aware row-level security, it prevents cross-agent state injections by enforcing access controls based on the agent's identity and assigned intent-scope.
