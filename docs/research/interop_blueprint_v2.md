# Interop Blueprint V2

This document details the standards and frameworks supported by the Universal Agent Bus (UAB) for inter-agent communication.

## Supported Protocols

1. **Agent Context Protocol (ACP)**
   - Integrates memory shard sharing and universal context synchronization logic.
   - Provides an implementation through the new `ACPAdapter`.
   - Ensures memory sync logic triggers on the `acp_context_sync` capability.

2. **Agent-to-Agent (A2A)**
   - Foundational messaging logic and protocols integrated into the hub.

3. **Model Context Protocol (MCP)**
   - Underlying core abstraction mechanism over models.

## Integrated Framework Adapters

| Framework   | Supported Capabilities                               |
|-------------|------------------------------------------------------|
| **OpenClaw**| `adaptive_reasoning`, `context_sync`                 |
| **CrewAI**  | `task_delegation`, `role_discovery`                  |
| **AutoGen** | `multi_agent_chat`, `subagent_exec`                  |
| **ACP**     | `acp_context_sync`, `a2a_messaging`                  |

All framework adapters now support standard multimodal context ingestion via `MemoryShard` definitions (UMMB Memory Sync Shards), ensuring unified memory sharing across autonomous environments without breaking existing architectural latency constraints.
