# Feature Inventory: MCP Any

This is the living masterlist of priority features for MCP Any.

## Priority 0: Core Infrastructure & Safety
| Feature Name | Description | Status | Source |
| :--- | :--- | :--- | :--- |
| **Policy Firewall** | Rego/CEL based hooking for tool calls (Zero Trust). | Planned | Server Roadmap |
| **HITL Middleware** | Suspension protocol for user approval flows. | Planned | Server Roadmap |
| **Recursive Context Protocol** | Standardize headers for Subagent inheritance. | Planned | Server Roadmap |
| **Context Inheritance (Universal)** | Cross-framework (Claude/Gemini) context propagation. | **New (2026-02-23)** | Market Sync |

## Priority 1: State & Developer Experience
| Feature Name | Description | Status | Source |
| :--- | :--- | :--- | :--- |
| **Shared KV Store** | Embedded SQLite "Blackboard" for agent swarms. | Planned | Server Roadmap |
| **Tool Playground** | Interactive UI for tool testing and discovery. | In Development | UI Roadmap |
| **Live Marble Diagrams** | Visualization of concurrent agent flows. | Planned | UI Roadmap |
| **Isolated Agent Comms** | Secure named pipes/sockets for inter-agent calls. | **New (2026-02-23)** | Market Sync |
| **Agent Trust Scoring** | Integrity verification for local MCP servers. | **New (2026-02-23)** | Market Sync |

## Priority 2: Enterprise & Scalability
| Feature Name | Description | Status | Source |
| :--- | :--- | :--- | :--- |
| **Compliance Reporting** | Automated SOC2/GDPR audit logs. | Planned | Server Roadmap |
| **Plugin Marketplace** | In-app browser for community MCP servers. | Planned | UI Roadmap |
| **Cost & Metrics Dashboard** | Real-time token usage and latency visualization. | Planned | UI Roadmap |

## Suggested Deprecations / Re-prioritizations
- **Reprioritize**: *Recursive Context Protocol* to **P0** (due to subagent context gap findings).
- **Add**: *Isolated Agent Comms* (P1) to mitigate local execution vulnerabilities.
