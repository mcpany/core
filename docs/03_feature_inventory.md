# Feature Inventory

## Master List of Priority Features

| Priority | Feature Name | Category | Status | Description |
| :--- | :--- | :--- | :--- | :--- |
| **P0** | **Policy Firewall Engine** | Security | Proposed | Rego/CEL based hooking for tool calls to enforce Zero Trust. |
| **P0** | **HITL Middleware** | Safety | Proposed | Suspension protocol for user approval flows (Human-in-the-Loop). |
| **P0** | **Isolation via Named Pipes** | Security | **New** | Use isolated Docker-bound named pipes for inter-agent comms instead of HTTP. |
| **P1** | **Recursive Context Protocol** | Comms | In Progress | Standardized headers for subagent context and scope inheritance. |
| **P1** | **Shared KV Store** | State | Proposed | Embedded SQLite "Blackboard" tool for agents to share state. |
| **P1** | **Lethal Trifecta Detection** | Security | **New** | Heuristic detection of agent sessions bridging private, untrusted, and external data. |
| **P1** | **Subagent Context Propagation** | Comms | **New** | Automatic propagation of auth and security scopes to spawned subagents. |
| **P1** | **Tool Playground & Explorer** | DevX | In Progress | Visual UI for discovering and testing tools with auto-generated forms. |
| **P1** | **Live Marble Diagrams** | Observability| Proposed | Reactive visualization of concurrent agent flows and tool calls. |
| **P2** | **Granular Scopes** | Security | In Progress | Capability-based token system (e.g., `fs:read:/tmp`). |
| **P2** | **Plugin Marketplace** | Ecosystem | Proposed | In-app browser to discover and install community MCP servers. |
| **P2** | **Agent Black Box Player** | Debugging | Proposed | Timeline-based replay of recorded agent sessions. |

## Feature Proposals: 2026-02-07

### 1. Isolation via Named Pipes
**Context:** Local HTTP tunneling is vulnerable to host-level network sniffing and unauthorized access.
**Proposal:** Replace local network listeners with named pipes bound within Docker containers or isolated namespaces.

### 2. Lethal Trifecta Detection
**Context:** Recent "Lethal Trifecta" attacks exploit agents that bridge internal and external data.
**Proposal:** Implement a monitor that flags or blocks sessions when an agent attempts to use tools from conflicting security domains simultaneously.

### 3. Subagent Context Propagation
**Context:** Subagents often lose the context of the parent agent, leading to re-authentication loops or security gaps.
**Proposal:** A standardized `x-mcp-context` header that carries encrypted scope and session data.
